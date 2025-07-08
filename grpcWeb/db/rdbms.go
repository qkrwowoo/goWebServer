package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	c "local/common"
	"strings"
	"sync"
	"time"
)

type rdbms struct {
	Mutex     sync.Mutex
	connQueue c.Queue
	Conn      []*sql.DB
	DB        DBinfo
}

/*
*******************************************************************************************
  - function	: LoadConfig
  - Description	: 환경파일 내부 RDBMS 연동 설정
  - Argument	: [ ]
  - Return		: [ ]
  - Etc         :

*******************************************************************************************
*/
func (r *rdbms) LoadConfig() {
	r.AllClose()
	r.DB.Info.LoadConfig()
	r.DB.init(c.CFG["DB"]["TYPE"].(string))
	r.Conn = make([]*sql.DB, r.DB.Info.Thread)
	if r.connQueue.V != nil && r.connQueue.V.Len() > 0 {
		r.connQueue.Clear()
	}
	r.connQueue.CreateQ()
	r.DB.AllOpen(r)
}

/*
*******************************************************************************************
  - function	: Default
  - Description	: 기본(하드코딩) RDBMS 연동 설정
  - Argument	: [ ]
  - Return		: [ ]
  - Etc         :

*******************************************************************************************
*/
func (r *rdbms) Default(dbType string) {
	r.DB.Info.dbtype = strings.ToLower(dbType)
	switch r.DB.Info.dbtype {
	// case "oracle", "godror":
	// 	r.DB.Info.Open = oracle_Open
	// 	r.DB.Info.AllOpen = oracle_AllOpen
	// 	r.DB.Info.dbtype = "godror"
	// 	r.DB.Info.port = "1521"
	case "mssql":
		r.DB.Open = mssql_Open
		r.DB.AllOpen = mssql_AllOpen
		r.DB.Info.dbtype = "sqlserver"
		r.DB.Info.port = "1433"
	case "mysql":
		r.DB.Open = mysql_Open
		r.DB.AllOpen = mysql_AllOpen
		r.DB.Info.dbtype = "mysql"
		r.DB.Info.port = "3306"
	default:
		r.DB.Open = mysql_Open
		r.DB.AllOpen = mysql_AllOpen
		r.DB.Info.dbtype = "mysql"
		r.DB.Info.port = "3306"
	}
	r.DB.Info.ip = "127.0.0.1"
	r.DB.Info.Ipaddr = r.DB.Info.ip + ":" + r.DB.Info.port

	r.DB.Info.ID = "testuser"
	r.DB.Info.PW = "password123"
	r.DB.Info.SID = "test"

	r.DB.Info.Timeout = 10000
	r.DB.Info.duration = time.Duration(r.DB.Info.Timeout) * time.Millisecond
	r.DB.Info.Thread = 10

	r.connQueue.CreateQ()
	r.Conn = make([]*sql.DB, r.DB.Info.Thread)
	r.DB.AllOpen(r)
}

/*
*******************************************************************************************
  - function	: AllClose
  - Description	: RDBMS 전체 접속 해제
  - Argument	: [ ]
  - Return		: [ ]
  - Etc         :

*******************************************************************************************
*/
func (r *rdbms) AllClose() {
	for i := 0; i < r.DB.Info.Thread; i++ {
		if len(r.Conn) > i && r.Conn[i] != nil {
			r.Conn[i].Close()
			r.Conn[i] = nil
		}
	}
}

/*
*******************************************************************************************
  - function	: GetDBConn
  - Description	: Connection 세션 조회
  - Argument	: [ (*context.Context) TIMEOUT 설정 ]
  - Return		: [ (interface{}) Redis Connection, (error) 오류 ]
  - Etc         : Multi-Thread 기반일 때, 아직 회수되지 않은 Connection 사용 방지를 위하여
    #             Queue 사용으로 유효한 Connection 사용 보장.
    #             Connection 이 끊긴 경우, 재연결 시도.

*******************************************************************************************
*/
func (r *rdbms) GetDBConn(ctx *context.Context) (interface{}, error) {
	var temp interface{}
	for {
		if temp = r.connQueue.PopQ(); temp != nil {
			break
		}
		if (*ctx).Err() != nil {
			return nil, (*ctx).Err()
		}
		time.Sleep(10 * time.Millisecond)
	}
	conn := temp.(*sql.DB)
	if err := conn.PingContext(*ctx); err != nil {
		c.Logging.Write(c.LogWARN, "DB Connection Broken[%s]... Try ReConnect ", err.Error())
		temp, err = r.DB.Open(r.DB, ctx)
		if err != nil {
			r.connQueue.PushQ(&sql.DB{})
			return nil, err
		} else {
			return temp, nil
		}
	} else {
		return temp, nil
	}
}

/*
*******************************************************************************************
  - function	: Do
  - Description	: RDBMS 쿼리 수행
  - Argument	: [ (*context.Context) TIMEOUT 설정, (string) Query문 ]
  - Return		: [ (table) 결과 테이블, (error) 오류 ]
  - Etc         :

*******************************************************************************************
*/
func (r *rdbms) Do(ctx *context.Context, query string) (table, error) {
	var conn interface{}
	var err error
	if conn, err = r.GetDBConn(ctx); conn == nil || err != nil {
		return table{}, fmt.Errorf("no connection idle [%s]", err.Error())
	}
	defer r.connQueue.PushQ(conn)
	switch query[0] {
	case 's', 'S':
		return r.Select(ctx, query, conn.(*sql.DB))
	case 'i', 'I', 'u', 'U', 'd', 'D':
		return r.Query(ctx, query, conn.(*sql.DB))
	case 'e', 'E', 'c', 'C':
		return r.Procedure(ctx, query, conn.(*sql.DB))
	default:
		cmd := strings.Split(query, " ")
		return table{}, fmt.Errorf("not support query [%s]", cmd[0])
	}
}

/*
*******************************************************************************************
  - function	: Select
  - Description	: Select 수행
  - Argument	: [ (*context.Context) TIMEOUT 설정, (string) Query문, (*sql.DB) 커넥션 ]
  - Return		: [ (table) 결과 테이블, (error) 오류 ]
  - Etc         :

*******************************************************************************************
*/
func (db *rdbms) Select(ctx *context.Context, query string, conn *sql.DB) (table, error) {
	var retnTables table
	var err error

	// query 실행
	rows, err := conn.QueryContext(*ctx, query)
	if err != nil {
		return table{}, err
	}
	defer rows.Close()

	// query 결과
	culums, err := rows.Columns()
	if err != nil {
		return table{}, err
	}

	retnTables.Cols = make([]column, len(culums))
	retnTables.Tuples = make([]map[string]string, 0)

	//// 뭐나옴?
	//test, err := rows.ColumnTypes()
	//for _, v := range test {
	//	fmt.Println(" select 3.2 ", v.Name(), v.DatabaseTypeName())
	//}

	tuple := make([]interface{}, len(culums))
	for i := range tuple {
		tuple[i] = new(sql.RawBytes)
	}
	// 결과 저장
	for rows.Next() {
		err = rows.Scan(tuple...)
		if err != nil {
			return table{}, err
		}
		tempMap := make(map[string]string)

		for i, v := range tuple {
			value := *(v.(*sql.RawBytes))
			if len(value) == 0 {
				tempMap[culums[i]] = ""
			} else {
				tempMap[culums[i]] = string(value)
			}
		}
		retnTables.Tuples = append(retnTables.Tuples, tempMap)
	}
	err = rows.Err()
	if err != nil {
		return table{}, err
	}
	if len(retnTables.Tuples) == 0 {
		retnTables.Status = false
		return table{}, fmt.Errorf("no data")
	}

	retnTables.Status = true
	return retnTables, nil
}

/*
*******************************************************************************************
  - function	: Query
  - Description	: Query 수행
  - Argument	: [ (*context.Context) TIMEOUT 설정, (string) Query문, (*sql.DB) 커넥션 ]
  - Return		: [ (table) 결과 테이블, (error) 오류 ]
  - Etc         : insert, delete, update 를 수행.

*******************************************************************************************
*/
func (db *rdbms) Query(ctx *context.Context, query string, conn *sql.DB) (table, error) {
	var retnTables table
	var err error

	tx, err := conn.Begin()
	if err != nil {
		return table{}, err
	}

	result_exce, err := tx.ExecContext(*ctx, query)
	defer tx.Rollback()
	if (*ctx).Err() != nil {
		return table{}, (*ctx).Err()
	} else if err != nil {
		return table{}, err
	}

	err = tx.Commit()
	if err != nil {
		return table{}, err
	}

	lowerCmd := strings.ToLower(strings.Split(query, " ")[0])
	ret, err := result_exce.RowsAffected()
	if err != nil {
		return table{}, err
	} else if ret <= 0 && lowerCmd != "create" { // create 문은 rows affected 가 없음
		return table{}, errors.New("no Rows Affected")
	}
	retnTables.Retn = int(ret)
	retnTables.Status = true
	return retnTables, nil
}

/*
*******************************************************************************************
  - function	: Procedure
  - Description	: Procedure 수행
  - Argument	: [ (*context.Context) TIMEOUT 설정, (string) Query문, (*sql.DB) 커넥션 ]
  - Return		: [ (table) 결과 테이블, (error) 오류 ]
  - Etc         : 테스트환경에서의 RDBMS 구축이 아직이라 보류

*******************************************************************************************
*/
func (db *rdbms) Procedure(ctx *context.Context, query string, conn *sql.DB) (table, error) {
	return table{}, nil
}
