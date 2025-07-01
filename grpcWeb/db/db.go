package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	c "local/common"

	"github.com/go-redis/redis/v8"
)

type ConnInfo struct {
	connQueue c.Queue
	Thread    int
	Timeout   int
	duration  time.Duration
	Ctx       context.Context
}

type DBinfo struct {
	Open         func(DBinfo, *context.Context) (interface{}, error)
	AllOpen      func(interface{}) error
	dbtype       string
	connectQuery string
	ID           string
	PW           string
	SID          string
	ip           string
	port         string
	Ipaddr       string
}

type table struct {
	Name   string
	Status bool
	Retn   int
	Cols   []column
	Tuples []map[string]string
}

type column struct {
	Name string
	Type string
}

var RDB rdb
var REDIS my_redis

func init() {
}

func (conninfo *ConnInfo) init() {
	conninfo.Timeout = c.S_Atoi(c.CFG["DB"]["TIMEOUT"].(string))
	conninfo.duration = time.Duration(conninfo.Timeout) * time.Millisecond
	conninfo.Thread = c.S_Atoi(c.CFG["DB"]["THREAD"].(string))
	conninfo.connQueue.Clear()
	conninfo.connQueue.CreateQ()
}

func (db *DBinfo) init(dbType string) {
	if dbType == "REDIS" {
		db.Open = redis_Open
		db.AllOpen = redis_AllOpen
		db.ID = c.CFG["REDIS"]["ID"].(string)
		db.PW = c.CFG["REDIS"]["PW"].(string)
		db.SID = c.CFG["REDIS"]["SID"].(string)
		db.ip = c.CFG["REDIS"]["IP"].(string)
		db.port = c.CFG["REDIS"]["PORT"].(string)
		db.Ipaddr = db.ip + ":" + db.port
	} else {
		db.dbtype = strings.ToLower(dbType)
		switch db.dbtype {
		case "mysql":
			db.Open = mysql_Open
			db.AllOpen = mysql_AllOpen
		case "mssql":
			db.Open = mssql_Open
			db.AllOpen = mssql_AllOpen
		case "oracle", "godror":
			db.Open = oracle_Open
			db.AllOpen = oracle_AllOpen
		}
		db.ID = c.CFG["DB"]["ID"].(string)
		db.PW = c.CFG["DB"]["PW"].(string)
		db.SID = c.CFG["DB"]["SID"].(string)
		db.ip = c.CFG["DB"]["IP"].(string)
		db.port = c.CFG["DB"]["PORT"].(string)
		db.Ipaddr = db.ip + ":" + db.port
	}

}

func (conninfo *ConnInfo) GetDBConn(ctx *context.Context) (interface{}, error) {
	var temp interface{}
	for {
		if temp := conninfo.connQueue.PopQ(); temp != nil {
			break
		}
		if (*ctx).Err() != nil {
			return nil, (*ctx).Err()
		}
		time.Sleep(10 * time.Millisecond)
	}

	switch conn := temp.(type) {
	case *sql.DB:
		if err := conn.PingContext(*ctx); err != nil {
			return nil, err
		} else {
			return temp.(*sql.DB), nil
		}
	case *redis.Client:
		if err := conn.Ping(*ctx).Err(); err != nil {
			return nil, err
		} else {
			return temp.(*redis.Client), nil
		}
	default:
		return nil, fmt.Errorf("Invalid Connection DB Type [%v]", conn)
	}
}

func CheckConnection(ctx *context.Context, conn *sql.DB) bool {
	if err := conn.PingContext(*ctx); err != nil {
		return false
	}
	return true
}
func CheckRedisConnection(ctx *context.Context, conn *redis.Client) bool {
	if err := conn.Ping(*ctx).Err(); err != nil {
		return false
	}
	return true
}
