package db

import (
	"context"
	"database/sql"
	"fmt"
	c "local/common"
	"strings"
	"sync"

	_ "github.com/godror/godror"
)

func oracle_Open(db DBinfo, ctx *context.Context) (interface{}, error) {
	if strings.ToLower(db.dbtype) == "oracle" {
		db.dbtype = "godror"
	}

	db.connectQuery = fmt.Sprintf("user=\"%s\" password=\"%s\" connectString=\"%s/%s\"", db.ID, db.PW, db.Ipaddr, db.SID)
	c.Logging.Write(c.LogDEBUG, "connectQuery [%.]", db.connectQuery)
	conn, err := sql.Open(db.dbtype, db.connectQuery)
	if err != nil {
		c.Logging.Write(c.LogERROR, "Oracle Connect Failed [%.][%s]", db.connectQuery, err.Error())
		return nil, err
	}
	c.Logging.Write(c.LogTRACE, "Oracle Connect Success")

	if err = conn.PingContext(*ctx); err != nil {
		return nil, err
	}
	return conn, nil
}

func oracle_AllOpen(in interface{}) error {
	r := in.(*rdb)

	var wg sync.WaitGroup
	wg.Add(r.Conninfo.Thread)
	for i := 0; i < r.Conninfo.Thread; i++ {
		index := i
		go func(*sync.WaitGroup, *rdb, int) {
			ctx, cancel := context.WithTimeout(context.Background(), r.Conninfo.duration)
			defer cancel()
			conn, err := oracle_Open(r.DBInfo, &ctx)
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
