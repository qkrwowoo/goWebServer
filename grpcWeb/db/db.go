package db

import (
	"context"
	"strings"
	"time"

	c "local/common"
)

type DBinfo struct {
	Open    func(DBinfo, *context.Context) (interface{}, error)
	AllOpen func(interface{}) error
	Info    Conninfo
}

type Conninfo struct {
	Thread       int
	Timeout      int
	duration     time.Duration
	Ctx          context.Context
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

var RDB rdbms
var REDIS my_redis

func init() {
	RDB.Default("mysql")
	REDIS.Default()
}

/*
*******************************************************************************************
  - function	: (*DBinfo) init
  - Description	: 사용 DB에 맞는 open 함수포인터를 지정
  - Argument	: [ (string) DB타입 ]
  - Return		: [ ]
  - Etc         :

*******************************************************************************************
*/
func (db *DBinfo) init(dbType string) {
	db.Info.dbtype = strings.ToLower(dbType)
	switch db.Info.dbtype {
	case "redis":
		db.Open = redis_Open
		db.AllOpen = redis_AllOpen
	case "mysql":
		db.Open = mysql_Open
		db.AllOpen = mysql_AllOpen
	case "mssql":
		db.Open = mssql_Open
		db.AllOpen = mssql_AllOpen
		// case "oracle", "godror":
		// 	db.Open = oracle_Open
		// 	db.AllOpen = oracle_AllOpen
	}
	db.Info.LoadConfig()
}

/*
*******************************************************************************************
  - function	: (*Conninfo) LoadConfig
  - Description	: 환경파일의 DB접속정보를 조회/갱신
  - Argument	: [ ]
  - Return		: [ ]
  - Etc         :

*******************************************************************************************
*/
func (cinfo *Conninfo) LoadConfig() {
	if cinfo.dbtype == "redis" {
		cinfo.ID = c.CFG["REDIS"]["ID"].(string)
		cinfo.PW = c.CFG["REDIS"]["PW"].(string)
		cinfo.SID = c.CFG["REDIS"]["SID"].(string)
		cinfo.ip = c.CFG["REDIS"]["IP"].(string)
		cinfo.port = c.CFG["REDIS"]["PORT"].(string)
		cinfo.Ipaddr = cinfo.ip + ":" + cinfo.port
		cinfo.Timeout = c.S_Atoi(c.CFG["REDIS"]["TIMEOUT"].(string))
		cinfo.Thread = c.S_Atoi(c.CFG["REDIS"]["THREAD"].(string))
	} else {
		cinfo.ID = c.CFG["DB"]["ID"].(string)
		cinfo.PW = c.CFG["DB"]["PW"].(string)
		cinfo.SID = c.CFG["DB"]["SID"].(string)
		cinfo.ip = c.CFG["DB"]["IP"].(string)
		cinfo.port = c.CFG["DB"]["PORT"].(string)
		cinfo.Ipaddr = cinfo.ip + ":" + cinfo.port
		cinfo.Timeout = c.S_Atoi(c.CFG["DB"]["TIMEOUT"].(string))
		cinfo.Thread = c.S_Atoi(c.CFG["DB"]["THREAD"].(string))
	}
	cinfo.duration = time.Duration(cinfo.Timeout) * time.Millisecond
}
