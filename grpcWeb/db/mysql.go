package db

import (
	"context"
	"database/sql"
	"fmt"
	c "local/common"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var MySQL DBinfo

func (db *DBinfo) MySQL_default() {
	db.ip = "127.0.0.1"
	db.port = "3306"
	db.Ipaddr = db.ip + ":" + db.port
	db.ID = "root"
	db.PW = "123qwe"
	db.SID = "test"
	db.dbtype = "mysql"
	db.Timeout = 10000
	db.duration = time.Duration(db.Timeout) * time.Millisecond
	db.Thread = 10
	db.Conn = make([]*sql.DB, db.Thread)
	db.connQueue.CreateQ()

	db.AllOpen = func(db *DBinfo) error {
		for i := 0; i < db.Thread; i++ {
			go func(*DBinfo, int) {
				db.connectQuery = fmt.Sprintf("%s:%s@tcp(%s)/%s", db.ID, db.PW, db.Ipaddr, db.SID)
				c.Logging.Write(c.LogDEBUG, "connectQuery [%.][%.]", db.Ipaddr, db.connectQuery)
				conn, err := sql.Open(db.dbtype, db.connectQuery)
				if err != nil {
					c.Logging.Write(c.LogERROR, "MySQL Open Failed [%.][%s]", db.connectQuery, err.Error())
					db.connQueue.PushQ(sql.DB{})
					return
				}

				ctx, cancel := context.WithTimeout(context.Background(), db.duration)
				defer cancel()

				if err = conn.PingContext(ctx); err != nil {
					c.Logging.Write(c.LogERROR, "MySQL Connect Failed [%.][%s]", db.connectQuery, err.Error())
					db.connQueue.PushQ(sql.DB{})
					return
				}
				c.Logging.Write(c.LogTRACE, "MySQL Connect Success")
				db.Conn[i] = conn
				db.connQueue.PushQ(db.Conn[i])
			}(db, i)
		}
		return nil
	}
	db.Open = func(db *DBinfo) (interface{}, error) {
		db.connectQuery = fmt.Sprintf("%s:%s@tcp(%s)/%s", db.ID, db.PW, db.Ipaddr, db.SID)
		c.Logging.Write(c.LogDEBUG, "connectQuery [%.][%.]", db.Ipaddr, db.connectQuery)
		conn, err := sql.Open(db.dbtype, db.connectQuery)
		if err != nil {
			c.Logging.Write(c.LogERROR, "MySQL Connect Failed [%.][%s]", db.connectQuery, err.Error())
			return nil, err
		}
		c.Logging.Write(c.LogTRACE, "MySQL Connect Success")

		ctx, cancel := context.WithTimeout(context.Background(), db.duration)
		defer cancel()

		err = conn.PingContext(ctx)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}
