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

type rdb struct {
	Mutex    sync.Mutex
	Conn     []*sql.DB
	Conninfo ConnInfo
	DBInfo   DBinfo
}

func (r *rdb) LoadConfig() {
	r.AllClose()
	r.Conninfo.init()
	r.DBInfo.init(c.CFG["DB"]["TYPE"].(string))
	r.Conn = make([]*sql.DB, r.Conninfo.Thread)
	r.DBInfo.AllOpen(r)
}

func (r *rdb) Default(dbType string) {
	r.DBInfo.dbtype = strings.ToLower(dbType)
	switch r.DBInfo.dbtype {
	case "oracle", "godror":
		r.DBInfo.port = "1521"
		r.DBInfo.dbtype = "godror"
		r.DBInfo.Open = oracle_Open
		r.DBInfo.AllOpen = oracle_AllOpen
	case "mssql":
		r.DBInfo.port = "1433"
		r.DBInfo.dbtype = "sqlserver"
		r.DBInfo.Open = mssql_Open
		r.DBInfo.AllOpen = mssql_AllOpen
	case "mysql":
		r.DBInfo.port = "3306"
		r.DBInfo.dbtype = "mysql"
		r.DBInfo.Open = mysql_Open
		r.DBInfo.AllOpen = mysql_AllOpen
	default:
		r.DBInfo.port = "3306"
		r.DBInfo.dbtype = "mysql"
		r.DBInfo.Open = mysql_Open
		r.DBInfo.AllOpen = mysql_AllOpen
	}

	r.DBInfo.ip = "127.0.0.1"
	r.DBInfo.Ipaddr = r.DBInfo.ip + ":" + r.DBInfo.port
	r.DBInfo.ID = "qkrwo"
	r.DBInfo.PW = "123qwe"
	r.DBInfo.SID = "test"

	r.Conninfo.Timeout = 10000
	r.Conninfo.duration = time.Duration(r.Conninfo.Timeout) * time.Millisecond
	r.Conninfo.Thread = 10
	r.Conninfo.connQueue.CreateQ()

	r.Conn = make([]*sql.DB, r.Conninfo.Thread)
	r.DBInfo.AllOpen(r)
}

func (r *rdb) AllClose() {
	for i := 0; i < r.Conninfo.Thread; i++ {
		if len(r.Conn) > i && r.Conn[i] != nil {
			r.Conn[i].Close()
			r.Conn[i] = nil
		}
	}
}

func (r *rdb) Do(ctx *context.Context, query string) (table, error) {
	var conn interface{}
	var err error
	if conn, err = r.Conninfo.GetDBConn(ctx); conn == nil || err != nil {
		return table{}, fmt.Errorf("no connection idle [%s]", err.Error())
	}
	defer r.Conninfo.connQueue.PushQ(conn)
	switch query[0] {
	case 's', 'S':
		return r.DBInfo.Select(ctx, query, conn.(*sql.DB))
	case 'i', 'I', 'u', 'U', 'd', 'D':
		return r.DBInfo.Query(ctx, query, conn.(*sql.DB))
	case 'e', 'E', 'c', 'C':
		return r.DBInfo.Procedure(ctx, query, conn.(*sql.DB))
	default:
		cmd := strings.Split(query, " ")
		return table{}, fmt.Errorf("not support query [%s]", cmd[0])
	}
}

func (db *DBinfo) Select(ctx *context.Context, query string, conn *sql.DB) (table, error) {
	var retnTables table
	var err error
	// var temp interface{}
	// if !CheckConnection(ctx, conn) {
	// 	temp, err = db.Open(db)
	// 	if err != nil {
	// 		return table{}, err
	// 	} else {
	// 		conn = temp.(*sql.DB)
	// 	}
	// }

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
	var err error
	// var temp interface{}
	// if !CheckConnection(ctx, conn) {
	// 	temp, err = db.Open(db)
	// 	if err != nil {
	// 		return table{}, err
	// 	} else {
	// 		conn = temp.(*sql.DB)
	// 	}
	// }

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
