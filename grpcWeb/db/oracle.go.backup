package db

import (
	"context"
	"database/sql"
	"fmt"
	c "local/common"
	"strings"
	"time"

	_ "github.com/godror/godror"
)

var Oracle DBinfo

func (db *DBinfo) Oracle_default() {
	db.ip = "127.0.0.1"
	db.port = "1521"
	db.Ipaddr = db.ip + ":" + db.port
	db.ID = "qkrwo"
	db.PW = "123qwe"
	db.SID = "test"
	db.dbtype = "godror"
	db.Timeout = 10000
	db.duration = time.Duration(db.Timeout) * time.Millisecond
	db.Thread = 10
	db.Conn = make([]*sql.DB, db.Thread)
	db.connQueue.CreateQ()

	db.AllOpen = func(db *DBinfo) error {
		for i := 0; i < db.Thread; i++ {
			go func(*DBinfo, int) {
				if strings.ToLower(db.dbtype) == "oracle" {
					db.dbtype = "godror"
				}
				var err error
				db.connectQuery = fmt.Sprintf("user=\"%s\" password=\"%s\" connectString=\"%s/%s\"", db.ID, db.PW, db.Ipaddr, db.SID)
				c.Logging.Write(c.LogDEBUG, "connectQuery [%.]", db.connectQuery)
				conn, err := sql.Open(db.dbtype, db.connectQuery)
				if err != nil {
					c.Logging.Write(c.LogERROR, "Oracle Open Failed [%.][%s]", db.connectQuery, err.Error())
					db.connQueue.PushQ(sql.DB{})
					return
				}
				ctx, cancel := context.WithTimeout(context.Background(), db.duration)
				defer cancel()

				if err = conn.PingContext(ctx); err != nil {
					c.Logging.Write(c.LogERROR, "Oracle Connect Failed [%.][%s]", db.connectQuery, err.Error())
					db.connQueue.PushQ(sql.DB{})
					return
				}
				c.Logging.Write(c.LogTRACE, "Oracle Connect Success")
				db.Conn[i] = conn
				db.connQueue.PushQ(db.Conn[i])
			}(db, i)
		}
		return nil
	}
	db.Open = func(db *DBinfo) (interface{}, error) {
		if strings.ToLower(db.dbtype) == "oracle" {
			db.dbtype = "godror"
		}
		var err error
		db.connectQuery = fmt.Sprintf("user=\"%s\" password=\"%s\" connectString=\"%s/%s\"", db.ID, db.PW, db.Ipaddr, db.SID)
		c.Logging.Write(c.LogDEBUG, "connectQuery [%.]", db.connectQuery)
		conn, err := sql.Open(db.dbtype, db.connectQuery)
		if err != nil {
			c.Logging.Write(c.LogERROR, "Oracle Connect Failed [%.][%s]", db.connectQuery, err.Error())
			return nil, err
		}
		c.Logging.Write(c.LogTRACE, "Oracle Connect Success")
		ctx, cancel := context.WithTimeout(context.Background(), db.duration)
		defer cancel()

		err = conn.PingContext(ctx)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}
