package db

import (
	"context"
	"database/sql"
	"fmt"
	c "local/common"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

func mysql_Open(db DBinfo, ctx *context.Context) (interface{}, error) {
	db.Info.connectQuery = fmt.Sprintf("%s:%s@tcp(%s)/%s", db.Info.ID, db.Info.PW, db.Info.Ipaddr, db.Info.SID)
	c.Logging.Write(c.LogDEBUG, "connectQuery [%.][%.]", db.Info.Ipaddr, db.Info.connectQuery)
	conn, err := sql.Open(db.Info.dbtype, db.Info.connectQuery)
	if err != nil {
		c.Logging.Write(c.LogERROR, "MySQL Open Failed [%.][%s]", db.Info.connectQuery, err.Error())
		return nil, err
	}
	if err = conn.PingContext(*ctx); err != nil {
		c.Logging.Write(c.LogERROR, "MySQL Connect Failed [%.][%s]", db.Info.connectQuery, err.Error())
		return nil, err
	}
	c.Logging.Write(c.LogTRACE, "MySQL Connect Success")

	return conn, nil
}

func mysql_AllOpen(in interface{}) error {
	r := in.(*rdb)

	var wg sync.WaitGroup
	wg.Add(r.DB.Info.Thread)
	for i := 0; i < r.DB.Info.Thread; i++ {
		index := i
		go func(*sync.WaitGroup, *rdb, int) {
			ctx, cancel := context.WithTimeout(context.Background(), r.DB.Info.duration)
			defer cancel()
			conn, err := mysql_Open(r.DB, &ctx)
			if err != nil {
				r.connQueue.PushQ(sql.DB{})
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
