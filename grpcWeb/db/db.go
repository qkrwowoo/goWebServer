package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	c "local/common"

	"github.com/go-redis/redis/v8"
)

type DBinfo struct {
	Mutex        sync.Mutex
	Open         func(db *DBinfo) (interface{}, error)
	AllOpen      func(db *DBinfo) error
	Conn         []*sql.DB
	RedisConn    []*redis.Client
	connQueue    c.Queue
	Thread       int
	Timeout      int
	duration     time.Duration
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

func init() {
	MySQL.MySQL_default()
	MsSQL.MsSQL_default()
	//Oracle.Oracle_default()
	Redis.Redis_default()
}

func (db *DBinfo) Init(dbType string) {
	if c.CFG[dbType] == nil {
		c.Logging.Write(c.LogWARN, "Not Found DB Section [%s] in configure file", dbType)
		return
	}
	db.AllClose()
	db.connQueue.Clear()
	db.connQueue.CreateQ()
	db.dbtype = strings.ToLower(dbType)
	if c.CFG[dbType]["CONNECT"] != nil && len(c.CFG[dbType]["CONNECT"].(string)) > 0 {
		db.connectQuery = c.CFG[dbType]["CONNECT"].(string)
	}
	db.ID = c.CFG[dbType]["ID"].(string)
	db.PW = c.CFG[dbType]["PW"].(string)
	db.SID = c.CFG[dbType]["SID"].(string)
	db.ip = c.CFG[dbType]["IP"].(string)
	db.port = c.CFG[dbType]["PORT"].(string)
	db.Ipaddr = db.ip + ":" + db.port
	db.Timeout = c.S_Atoi(c.CFG[dbType]["TIMEOUT"].(string))
	db.duration = time.Duration(db.Timeout) * time.Millisecond
	db.Thread = c.S_Atoi(c.CFG[dbType]["THREAD"].(string))
	db.Conn = make([]*sql.DB, db.Thread)
	db.RedisConn = make([]*redis.Client, db.Thread)
	db.AllOpen(db)
}

func (db *DBinfo) AllClose() {
	for i := 0; i < db.Thread; i++ {
		if len(db.Conn) >= i && db.Conn[i] != nil {
			db.Conn[i].Close()
			db.Conn[i] = nil
		} else if len(db.RedisConn) > i && db.RedisConn[i] != nil {
			db.RedisConn[i].Close()
			db.RedisConn[i] = nil
		}
	}
}

func (db *DBinfo) GetDBConn(ctx *context.Context) interface{} {
	for {
		if temp := db.connQueue.PopQ(); temp != nil {
			return temp.(*sql.DB)
		}
		select {
		case <-(*ctx).Done():
			return nil
		}
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

func (db *DBinfo) Do(ctx *context.Context, query string) (table, error) {
	var conn interface{}
	if conn = db.GetDBConn(ctx); conn == nil {
		return table{}, fmt.Errorf("no connection idle")
	}
	defer db.connQueue.PushQ(conn)
	switch query[0] {
	case 's', 'S':
		return db.Select(ctx, query, conn.(*sql.DB))
	case 'i', 'I', 'u', 'U', 'd', 'D':
		return db.Query(ctx, query, conn.(*sql.DB))
	case 'e', 'E', 'c', 'C':
		return db.Procedure(ctx, query, conn.(*sql.DB))
	default:
		cmd := strings.Split(query, " ")
		return table{}, fmt.Errorf("not support query [%s]", cmd[0])
	}
}

func (db *DBinfo) Select(ctx *context.Context, query string, conn *sql.DB) (table, error) {
	var retnTables table
	var temp interface{}
	var err error
	if !CheckConnection(ctx, conn) {
		temp, err = db.Open(db)
		if err != nil {
			return table{}, err
		} else {
			conn = temp.(*sql.DB)
		}
	}

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

func (db *DBinfo) Query(ctx *context.Context, query string, conn *sql.DB) (table, error) {
	var retnTables table
	var temp interface{}
	var err error
	if !CheckConnection(ctx, conn) {
		temp, err = db.Open(db)
		if err != nil {
			return table{}, err
		} else {
			conn = temp.(*sql.DB)
		}
	}

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

func (db *DBinfo) Procedure(ctx *context.Context, query string, conn *sql.DB) (table, error) {
	return table{}, nil
}
