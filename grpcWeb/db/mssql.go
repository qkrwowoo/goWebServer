package db

import (
	"context"
	"database/sql"
	"fmt"
	c "local/common"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

var MsSQL DBinfo

func (db *DBinfo) MsSQL_default() {
	db.ip = "127.0.0.1"
	db.port = "1433"
	db.Ipaddr = db.ip + ":" + db.port
	db.ID = "qkrwo"
	db.PW = "123qwe"
	db.SID = "test"
	db.dbtype = "sqlserver"
	db.Timeout = 10000
	db.duration = time.Duration(db.Timeout) * time.Millisecond
	db.Thread = 10
	db.Conn = make([]*sql.DB, db.Thread)
	db.connQueue.CreateQ()

	db.AllOpen = func(db *DBinfo) error {
		for i := 0; i < db.Thread; i++ {
			go func(*DBinfo, int) {
				db.connectQuery = fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;", db.ip, db.port, db.ID, db.PW, db.SID)
				c.Logging.Write(c.LogDEBUG, "connectQuery [%.]", db.connectQuery)
				conn, err := sql.Open(db.dbtype, db.connectQuery)
				if err != nil {
					c.Logging.Write(c.LogERROR, "MsSQL Open Failed [%.][%s]", db.connectQuery, err.Error())
					db.connQueue.PushQ(sql.DB{})
					return
				}

				ctx, cancel := context.WithTimeout(context.Background(), db.duration)
				defer cancel()

				if err = conn.PingContext(ctx); err != nil {
					c.Logging.Write(c.LogERROR, "MsSQL Connect Failed [%.][%s]", db.connectQuery, err.Error())
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
		db.connectQuery = fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;", db.ip, db.port, db.ID, db.PW, db.SID)
		c.Logging.Write(c.LogDEBUG, "connectQuery [%.]", db.connectQuery)
		conn, err := sql.Open(db.dbtype, db.connectQuery)
		if err != nil {
			c.Logging.Write(c.LogERROR, "MsSQL Connect Failed [%.][%s]", db.connectQuery, err.Error())
			return nil, err
		}
		c.Logging.Write(c.LogTRACE, "MySQL Connect Success")

		ctx, cancel := context.WithTimeout(context.Background(), db.duration)
		defer cancel()

		if conn.PingContext(ctx) != nil {
			return nil, err
		}
		return conn, nil
	}
}
