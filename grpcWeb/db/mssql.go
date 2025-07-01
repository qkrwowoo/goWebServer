package db

import (
	"context"
	"database/sql"
	"fmt"
	c "local/common"
	"sync"

	_ "github.com/denisenkom/go-mssqldb"
)

func mssql_Open(db DBinfo, ctx *context.Context) (interface{}, error) {
	db.connectQuery = fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;", db.ip, db.port, db.ID, db.PW, db.SID)
	c.Logging.Write(c.LogDEBUG, "connectQuery [%.]", db.connectQuery)
	conn, err := sql.Open(db.dbtype, db.connectQuery)
	if err != nil {
		c.Logging.Write(c.LogERROR, "MsSQL Connect Failed [%.][%s]", db.connectQuery, err.Error())
		return nil, err
	}
	c.Logging.Write(c.LogTRACE, "MySQL Connect Success")

	if err = conn.PingContext(*ctx); err != nil {
		return nil, err
	}
	return conn, nil
}

func mssql_AllOpen(in interface{}) error {
	r := in.(*rdb)

	var wg sync.WaitGroup
	wg.Add(r.Conninfo.Thread)
	for i := 0; i < r.Conninfo.Thread; i++ {
		index := i
		go func(*sync.WaitGroup, *rdb, int) {
			ctx, cancel := context.WithTimeout(context.Background(), r.Conninfo.duration)
			defer cancel()
			conn, err := mssql_Open(r.DBInfo, &ctx)
			if err != nil {
				r.Conninfo.connQueue.PushQ(sql.DB{})
			} else {
				r.Conn[index] = conn.(*sql.DB)
				r.Conninfo.connQueue.PushQ(r.Conn[index])
			}
			wg.Done()
		}(&wg, r, index)
	}
	wg.Wait()
	return nil

}
