package db

import (
	"context"
	"database/sql"
	"fmt"
	c "local/common"
	"sync"

	_ "github.com/denisenkom/go-mssqldb"
)

/*
*******************************************************************************************
  - function	: mssql_Open
  - Description	: MsSQL 단일 접속
  - Argument	: [ (DBinfo) DB연동정보, (*context.Context) TIMEOUT 설정 ]
  - Return		: [ (interface{}) MySQL Connection, (error) 오류 ]
  - Etc         :

*******************************************************************************************
*/
func mssql_Open(db DBinfo, ctx *context.Context) (interface{}, error) {
	db.Info.connectQuery = fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;", db.Info.ip, db.Info.port, db.Info.ID, db.Info.PW, db.Info.SID)
	c.Logging.Write(c.LogDEBUG, "connectQuery [%.]", db.Info.connectQuery)
	conn, err := sql.Open(db.Info.dbtype, db.Info.connectQuery)
	if err != nil {
		c.Logging.Write(c.LogERROR, "MsSQL Open Failed [%.][%s]", db.Info.connectQuery, err.Error())
		return nil, err
	}
	if err = conn.PingContext(*ctx); err != nil {
		c.Logging.Write(c.LogERROR, "MsSQL Connect Failed [%.][%s]", db.Info.connectQuery, err.Error())
		return nil, err
	}

	c.Logging.Write(c.LogTRACE, "MySQL Connect Success")
	return conn, nil
}

/*
*******************************************************************************************
  - function	: mssql_AllOpen
  - Description	: MsSQL 전체 접속
  - Argument	: [ (interface{}) DB연동정보 ]
  - Return		: [ (error) 오류 ]
  - Etc         : Thread 개수 만큼 동시 접속할 수 있는 Connection 생성

*******************************************************************************************
*/
func mssql_AllOpen(in interface{}) error {
	r := in.(*rdbms)

	var wg sync.WaitGroup
	wg.Add(r.DB.Info.Thread)
	for i := 0; i < r.DB.Info.Thread; i++ {
		index := i
		go func(*sync.WaitGroup, *rdbms, int) {
			ctx, cancel := context.WithTimeout(context.Background(), r.DB.Info.duration)
			defer cancel()
			conn, err := mssql_Open(r.DB, &ctx)
			if err != nil {
				r.connQueue.PushQ(&sql.DB{})
			} else {
				r.Conn[index] = conn.(*sql.DB)
				r.connQueue.PushQ(r.Conn[index])
			}
			wg.Done()
		}(&wg, r, index)
	}
	wg.Wait()
	return nil
}
