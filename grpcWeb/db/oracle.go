package db

//##################################################################
// 오라클은 so 가 필요함. 지금은 테스트용이니 일단 주석.
//##################################################################
//import (
//	"context"
//	"database/sql"
//	"fmt"
//	c "local/common"
//	"strings"
//	"sync"
//
//	_ "github.com/godror/godror"
//)
//
//func oracle_Open(db DBinfo, ctx *context.Context) (interface{}, error) {
//	if strings.ToLower(db.Info.dbtype) == "oracle" {
//		db.Info.dbtype = "godror"
//	}
//	db.Info.connectQuery = fmt.Sprintf("user=\"%s\" password=\"%s\" connectString=\"%s/%s\"", db.Info.ID, db.Info.PW, db.Info.Ipaddr, db.Info.SID)
//	c.Logging.Write(c.LogDEBUG, "connectQuery [%.]", db.Info.connectQuery)
//	conn, err := sql.Open(db.Info.dbtype, db.Info.connectQuery)
//	if err != nil {
//		c.Logging.Write(c.LogERROR, "Oracle Open Failed [%.][%s]", db.Info.connectQuery, err.Error())
//		return nil, err
//	}
//	if err = conn.PingContext(*ctx); err != nil {
//		c.Logging.Write(c.LogERROR, "Oracle Connect Failed [%.][%s]", db.Info.connectQuery, err.Error())
//		return nil, err
//	}
//	c.Logging.Write(c.LogTRACE, "Oracle Connect Success")
//	return conn, nil
//}
//func oracle_AllOpen(in interface{}) error {
//	r := in.(*rdb)
//	var wg sync.WaitGroup
//	wg.Add(r.DB.Info.Thread)
//	for i := 0; i < r.DB.Info.Thread; i++ {
//		index := i
//		go func(*sync.WaitGroup, *rdb, int) {
//			ctx, cancel := context.WithTimeout(context.Background(), r.DB.Info.duration)
//			defer cancel()
//			conn, err := oracle_Open(r.DB, &ctx)
//			if err != nil {
//				r.connQueue.PushQ(&sql.DB{})
//			} else {
//				r.Conn[index] = conn.(*sql.DB)
//				r.connQueue.PushQ(r.Conn[index])
//			}
//			wg.Done()
//		}(&wg, r, index)
//	}
//	wg.Wait()
//	return nil
//}
