package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	c "local/common"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type my_redis struct {
	Mutex     sync.Mutex
	connQueue c.Queue
	Conn      []*redis.Client
	DB        DBinfo
}

/*
*******************************************************************************************
  - function	: LoadConfig
  - Description	: 환경파일 내부 Redis 연동 설정
  - Argument	: [ ]
  - Return		: [ ]
  - Etc         :

*******************************************************************************************
*/
func (r *my_redis) LoadConfig() {
	r.AllClose()
	r.DB.init("redis")
	r.Conn = make([]*redis.Client, r.DB.Info.Thread)
	if r.connQueue.V != nil && r.connQueue.V.Len() > 0 {
		r.connQueue.Clear()
	}
	r.connQueue.CreateQ()
	r.DB.AllOpen(r)
}

/*
*******************************************************************************************
  - function	: Default
  - Description	: 기본(하드코딩) Redis 연동 설정
  - Argument	: [ ]
  - Return		: [ ]
  - Etc         :

*******************************************************************************************
*/
func (r *my_redis) Default() {
	r.DB.Info.dbtype = "redis"

	r.DB.Open = redis_Open
	r.DB.AllOpen = redis_AllOpen

	r.DB.Info.ip = "127.0.0.1"
	r.DB.Info.port = "6379"
	r.DB.Info.Ipaddr = r.DB.Info.ip + ":" + r.DB.Info.port

	r.DB.Info.ID = ""
	r.DB.Info.PW = ""
	r.DB.Info.SID = "0"

	r.DB.Info.Timeout = 10000
	r.DB.Info.duration = time.Duration(r.DB.Info.Timeout) * time.Millisecond
	r.DB.Info.Thread = 10

	r.connQueue.CreateQ()
	r.Conn = make([]*redis.Client, r.DB.Info.Thread)
	r.DB.AllOpen(r)
}

/*
*******************************************************************************************
  - function	: redis_Open
  - Description	: Redis 단일 접속
  - Argument	: [ (DBinfo) DB연동정보, (*context.Context) TIMEOUT 설정 ]
  - Return		: [ (interface{}) Redis Connection, (error) 오류 ]
  - Etc         :

*******************************************************************************************
*/
func redis_Open(db DBinfo, ctx *context.Context) (interface{}, error) {
	var err error
	redisNum, _ := strconv.Atoi(db.Info.SID)
	redisConn := redis.NewClient(&redis.Options{
		Addr:     db.Info.Ipaddr,
		Username: db.Info.ID,
		Password: db.Info.PW,
		DB:       redisNum,
	})
	if _, err = redisConn.Ping(*ctx).Result(); err != nil {
		c.Logging.Write(c.LogERROR, "Redis Connect Failed [%s]", err.Error())
		return nil, err
	}
	c.Logging.Write(c.LogTRACE, "Redis Connect Success")
	return redisConn, nil
}

/*
*******************************************************************************************
  - function	: redis_AllOpen
  - Description	: Redis 전체 접속
  - Argument	: [ (interface{}) DB연동정보 ]
  - Return		: [ (error) 오류 ]
  - Etc         : Thread 개수 만큼 동시 접속할 수 있는 Connection 생성

*******************************************************************************************
*/
func redis_AllOpen(in interface{}) error {
	r := in.(*my_redis)

	var wg sync.WaitGroup
	wg.Add(r.DB.Info.Thread)
	for i := 0; i < r.DB.Info.Thread; i++ {
		index := i
		go func(*sync.WaitGroup, *my_redis, int) {
			ctx, cancel := context.WithTimeout(context.Background(), r.DB.Info.duration)
			defer cancel()
			conn, err := redis_Open(r.DB, &ctx)
			if err != nil {
				r.connQueue.PushQ(&redis.Client{})
				return
			} else {
				r.Conn[index] = conn.(*redis.Client)
				r.connQueue.PushQ(r.Conn[index])
			}
			wg.Done()
		}(&wg, r, index)
	}
	return nil
}

/*
*******************************************************************************************
  - function	: AllClose
  - Description	: Redis 전체 접속 해제
  - Argument	: [ ]
  - Return		: [ ]
  - Etc         :

*******************************************************************************************
*/
func (r *my_redis) AllClose() {
	for i := 0; i < r.DB.Info.Thread; i++ {
		if len(r.Conn) > i && r.Conn[i] != nil {
			r.Conn[i].Close()
			r.Conn[i] = nil
		}
	}
}

/*
*******************************************************************************************
  - function	: GetDBConn
  - Description	: Connection 세션 조회
  - Argument	: [ (*context.Context) TIMEOUT 설정 ]
  - Return		: [ (interface{}) Redis Connection, (error) 오류 ]
  - Etc         : Multi-Thread 기반일 때, 아직 회수되지 않은 Connection 사용 방지를 위하여
    #             Queue 사용으로 유효한 Connection 사용 보장.
    #             Connection 이 끊긴 경우, 재연결 시도.

*******************************************************************************************
*/
func (r *my_redis) GetDBConn(ctx *context.Context) (interface{}, error) {
	var temp interface{}
	for {
		if temp = r.connQueue.PopQ(); temp != nil {
			break
		}
		if (*ctx).Err() != nil {
			return nil, (*ctx).Err()
		}
		time.Sleep(10 * time.Millisecond)
	}
	conn := temp.(*redis.Client)
	if err := conn.Ping(*ctx).Err(); err != nil {
		c.Logging.Write(c.LogWARN, "DB Connection Broken[%s]... Try ReConnect ", err.Error())
		temp, err = r.DB.Open(r.DB, ctx)
		if err != nil {
			r.connQueue.PushQ(&redis.Client{})
			return nil, err
		} else {
			return temp, nil
		}
	} else {
		return temp, nil
	}
}

/*
*******************************************************************************************
  - function	: RedisDo
  - Description	: Redis Query 수행
  - Argument	: [ (*context.Context) TIMEOUT 설정, (string) Query문 ]
  - Return		: [ ([][]byte) Redis 응답, (error) 오류 ]
  - Etc         : 응답값의 커스터마이징 상태.

*******************************************************************************************
*/
func (r *my_redis) RedisDo(ctx *context.Context, redis_query string) ([][]byte, error) {
	var expire time.Duration
	var err error

	var boolFlag bool = false
	var intFlag bool = false
	var sliceFlag bool = false

	var readStr string = ""
	var readInt int64 = 0
	var readBool bool = false
	var readStrSlice []string
	var readInterface []interface{}
	var readMap map[string]string
	var readtime time.Duration

	var keys []string
	var vals []string

	var idx64 int64

	var command string
	var key string
	var value []string
	var hashKey string
	var hashValue []string

	var temp interface{}
	var conn *redis.Client

	if temp, err = r.GetDBConn(ctx); temp == nil || err != nil {
		return nil, fmt.Errorf("no connection idle [%s]", err.Error())
	}
	conn = temp.(*redis.Client)
	defer r.connQueue.PushQ(conn)

	query := strings.Split(redis_query, " ")

	command = query[0]
	key = query[1]
	value = append(value, query[2:]...)

	hashKey = query[1]
	hashValue = append(hashValue, query[2:]...)

	keys = make([]string, 0)
	vals = make([]string, 0)
	expire = 60 * 60 * time.Second // 60분 후 삭제
	// expire = 20 * time.Second // REDIS 보관시간제한 설정 !!!지원하는 것만!!!

	switch strings.ToUpper(command) {
	///////////////////////////////////////////////////////////////////////////////////////
	// String Command
	///////////////////////////////////////////////////////////////////////////////////////
	case "SET": // 문자열 저장
		keys = append(keys, "RESULT")
		readStr, err = conn.Set(*ctx, key, value[0], expire).Result()
	case "SETEX": // 문자열 저장 (시간 초 후 삭제)
		keys = append(keys, "RESULT")
		i, _ := strconv.Atoi(value[0])
		expire = time.Duration(i) * time.Second
		readStr, err = conn.Set(*ctx, key, value[1], expire).Result()
	case "GET": // 문자열 조회
		keys = append(keys, key)
		readStr, err = conn.Get(*ctx, key).Result()
	case "DEL": // 문자열 삭제
		intFlag = true
		readInt, err = conn.Del(*ctx, key).Result()
	case "MSET": // 문자열 다중 저장
		keys = append(keys, "RESULT")
		readStr, err = conn.MSet(*ctx, value).Result()
	case "MGET": // 문자열 다중 조회
		//_, keys = Common.Map2Strings(keyValue)
		keys = append(keys, value...)
		readInterface, err = conn.MGet(*ctx, value...).Result()
	case "SETNX": // 문자열 저장 (KEY 값 있으면 무시)
		boolFlag = true
		readBool, err = conn.SetNX(*ctx, key, value[0], expire).Result()
	case "MSETNX": // 문자열 저장 (KEY 값 있으면 무시)
		boolFlag = true
		readBool, err = conn.MSetNX(*ctx, value).Result()
	case "GETEX": // 문자열 조회/삭제시간 지정 (REDIS 6.2 이상)
		keys = append(keys, key)
		readStr, err = conn.GetEx(*ctx, key, expire).Result()
	case "GETDEL": // 문자열 조회/삭제 (REDIS 6.2 이상)
		keys = append(keys, key)
		readStr, err = conn.GetDel(*ctx, key).Result()
	case "APPEND": // 문자열 추가
		intFlag = true
		readInt, err = conn.Append(*ctx, key, value[0]).Result()
	case "STRLEN":
		intFlag = true
		readInt, err = conn.StrLen(*ctx, key).Result()

	///////////////////////////////////////////////////////////////////////////////////////
	// Hash Command
	///////////////////////////////////////////////////////////////////////////////////////
	case "HSET": // 단일 HASH 저장
		intFlag = true
		readInt, err = conn.HSet(*ctx, hashKey, hashValue).Result()
	case "HMSET": // 다중 HASH 저장
		boolFlag = true
		readBool, err = conn.HMSet(*ctx, hashKey, hashValue).Result()
	case "HGET": // 단일 HASH 조회
		keys = hashValue
		readStr, err = conn.HGet(*ctx, hashKey, value[0]).Result()
	case "HDEL": // HASH 삭제
		intFlag = true
		readInt, err = conn.HDel(*ctx, hashKey, hashValue...).Result()
	case "HLEN": // HASH 개수 조회
		intFlag = true
		readInt, err = conn.HLen(*ctx, hashKey).Result()
	case "HMGET": // 다중 HASH 조회
		//_, keys = Common.Map2Strings(keyValue)
		keys = hashValue
		readInterface, err = conn.HMGet(*ctx, hashKey, hashValue...).Result()
	case "HGETALL": // 전체 HASH 조회
		readMap, err = conn.HGetAll(*ctx, hashKey).Result()
	case "HKEYS": // 전체 HASH KEY 조회
		sliceFlag = true
		readStrSlice, err = conn.HKeys(*ctx, hashKey).Result()
	case "HVALS": // 전체 HASH VALUE 조회
		sliceFlag = true
		readStrSlice, err = conn.HVals(*ctx, hashKey).Result()
	case "HEXISTS": // HASH FIELD 조회
		boolFlag = true
		readBool, err = conn.HExists(*ctx, hashKey, hashValue[0]).Result()
	case "HSETNX": // 단일 HASH 저장 (중복 저장 방지)
		boolFlag = true
		readBool, err = conn.HSetNX(*ctx, hashKey, hashValue[0], hashValue[1]).Result()
	///////////////////////////////////////////////////////////////////////////////////////
	// LIST
	///////////////////////////////////////////////////////////////////////////////////////
	case "LPUSH":
		intFlag = true
		//_, vals = Common.Map2Strings(keyValue)
		readInt, err = conn.LPush(*ctx, key, value[0:]).Result()
	case "RPUSH":
		intFlag = true
		//_, vals = Common.Map2Strings(keyValue)
		readInt, err = conn.RPush(*ctx, key, value[0:]).Result()
	case "LPOP":
		keys = append(keys, "RESULT")
		readStr, err = conn.LPop(*ctx, key).Result()
	case "RPOP":
		keys = append(keys, "RESULT")
		readStr, err = conn.RPop(*ctx, key).Result()
	case "LLEN":
		intFlag = true
		readInt, err = conn.LLen(*ctx, key).Result()

	case "LINDEX": // 지정 인덱스 조회
		keys = append(keys, "RESULT")
		idx64, err = strconv.ParseInt(value[0], 10, 64)
		if err == nil {
			readStr, err = conn.LIndex(*ctx, key, idx64).Result()
		}
	case "LRANGE": // 지정 인덱스 범위 조회
		sliceFlag = true
		startIdx, err2 := strconv.Atoi(value[0])
		endIdx, err3 := strconv.Atoi(value[1])
		if err2 != nil {
			err = err2
		} else if err3 != nil {
			err = err3
		} else {
			readStrSlice, err = conn.LRange(*ctx, key, int64(startIdx), int64(endIdx)).Result()
		}
	case "LSET": // 리스트 index 위치 데이터 변경
		keys = append(keys, "RESULT")
		idx64, err = strconv.ParseInt(value[0], 10, 64)
		if err == nil {
			readStr, err = conn.LSet(*ctx, key, idx64, value[1]).Result()
		}
	case "LREM": // 리스트에서 int64 개수만큼 value 삭제
		intFlag = true
		idx64, err = strconv.ParseInt(value[0], 10, 64)
		if err == nil {
			readInt, err = conn.LRem(*ctx, key, idx64, value[1]).Result()
		}
	case "RPOPLPUSH": // (1)리스트에서 RPOP 해서 (2)리스트 에 LPUSH
		keys = append(keys, "RESULT")
		readStr, err = conn.RPopLPush(*ctx, key, value[0]).Result()
	case "LPUSHX": // 이미 리스트가 존재해야지만 PUSH
		intFlag = true
		readInt, err = conn.LPushX(*ctx, key, value[0:]).Result()
	case "RPUSHX": // 이미 리스트가 존재해야지만 PUSH
		intFlag = true
		readInt, err = conn.RPushX(*ctx, key, value[0:]).Result()
	case "BLPOP": // 이미 값이 있으면 LPOP. 없으면 올때까지 대기 (poptime / timeout)
		var popTimeout time.Duration
		sliceFlag = true
		poptime, err2 := strconv.Atoi(value[0])
		if err2 == nil {
			popTimeout = time.Duration(poptime) * time.Second
		} else {
			popTimeout = r.DB.Info.duration
		}

		readStrSlice, err = conn.BLPop(*ctx, popTimeout, key).Result()
	case "BRPOP": // 이미 값이 있으면 RPOP. 없으면 올때까지 대기 (poptime / timeout)
		var popTimeout time.Duration
		sliceFlag = true
		poptime, err2 := strconv.Atoi(value[0])
		if err2 == nil {
			popTimeout = time.Duration(poptime) * time.Second
		} else {
			popTimeout = r.DB.Info.duration
		}
		readStrSlice, err = conn.BRPop(*ctx, popTimeout, key).Result()
	case "BRPOPLPUSH": // 이미 값이 있으면 RPOP. 없으면 올때까지 대기 (poptime / timeout) + LPUSH 까지
		keys = append(keys, "RESULT")
		var popTimeout time.Duration
		var poptime int
		poptime, err = strconv.Atoi(value[1])
		if err == nil {
			popTimeout = time.Duration(poptime) * time.Second
			readStr, err = conn.BRPopLPush(*ctx, key, value[0], popTimeout).Result()
		}

	///////////////////////////////////////////////////////////////////////////////////////
	// SET
	///////////////////////////////////////////////////////////////////////////////////////
	case "SADD": // 멤버 추가
		intFlag = true
		readInt, err = conn.SAdd(*ctx, key, value[0:]).Result()
	case "SREM": // 멤버 삭제
		intFlag = true
		readInt, err = conn.SRem(*ctx, key, value[0:]).Result()
	case "SMEMBERS": // 전체 조회
		sliceFlag = true
		readStrSlice, err = conn.SMembers(*ctx, key).Result()
	case "SCARD": // 개수 조회
		intFlag = true
		readInt, err = conn.SCard(*ctx, key).Result()
	case "SPOP": // 랜덤 뽑기
		keys = append(keys, "RESULT")
		readStr, err = conn.SPop(*ctx, key).Result()
	case "SRANDMEMBER": // 랜덤 조회
		keys = append(keys, "RESULT")
		readStr, err = conn.SRandMember(*ctx, key).Result()
	case "SISMEMBER": // 멤버 유무 확인
		boolFlag = true
		readBool, err = conn.SIsMember(*ctx, key, value[0]).Result()
		/*
			case "SMOVE": // 다른 집합(key)으로 멤버 이동
				boolFlag = true
				readBool, err = conn.SMove(*ctx, key, value[0], value[1:]).Result()
		*/
	///////////////////////////////////////////////////////////////////////////////////////
	// ZSET
	///////////////////////////////////////////////////////////////////////////////////////
	case "ZADD": // 저장
		var i int
		var err2 error
		var f64 float64
		var members []*redis.Z
		var tempKeys []float64
		var tempVals []string

		intFlag = true
		for i = 0; i < len(value); i += 2 {
			f64, err2 = strconv.ParseFloat(value[i], 64)
			if err2 != nil {
				break
			}
			tempKeys = append(tempKeys, f64)
			tempVals = append(tempVals, value[i+1])
		}
		if i < len(value) {
			err = err2
		} else {
			members = make([]*redis.Z, len(tempKeys))
			for i := 0; i < len(tempKeys); i++ {
				members[i] = &redis.Z{
					Score:  tempKeys[i],
					Member: tempVals[i],
				}
			}
			readInt, err = conn.ZAdd(*ctx, key, members...).Result()
		}

	case "ZSCORE": // 단건 조회
		var f float64
		intFlag = true
		f, err = conn.ZScore(*ctx, key, value[0]).Result()
		readInt = int64(f)
	case "ZRANK": // 인덱스 순위 조회
		intFlag = true
		readInt, err = conn.ZRank(*ctx, key, value[0]).Result()
	case "ZREVRANK": // 인덱스 순위 역순 조회
		intFlag = true
		readInt, err = conn.ZRevRank(*ctx, key, value[0]).Result()
	case "ZRANGE": // 범위 조회
		sliceFlag = true
		startIdx, err2 := strconv.ParseInt(value[0], 10, 64)
		endIdx, err3 := strconv.ParseInt(value[1], 10, 64)
		if err2 != nil {
			err = err2
		} else if err3 != nil {
			err = err3
		} else {
			readStrSlice, err = conn.ZRange(*ctx, key, startIdx, endIdx).Result()
		}
	case "ZREVRANGE": // 범위 역순 조회
		sliceFlag = true
		startIdx, err2 := strconv.ParseInt(value[0], 10, 64)
		endIdx, err3 := strconv.ParseInt(value[1], 10, 64)
		if err2 != nil {
			err = err2
		} else if err3 != nil {
			err = err3
		} else {
			readStrSlice, err = conn.ZRevRange(*ctx, key, startIdx, endIdx).Result()
		}
	case "ZCARD": // 개수 조회
		intFlag = true
		readInt, err = conn.ZCard(*ctx, key).Result()
	case "ZREM": // 멤버 삭제
		intFlag = true
		readInt, err = conn.ZRem(*ctx, key, value[0:]).Result()
	case "ZINCRBY": // 멤버 스코어 증감
		var f float64
		intFlag = true
		f64, err2 := strconv.ParseFloat(value[0], 64)
		if err2 != nil {
			err = err2
		} else {
			f, err = conn.ZIncrBy(*ctx, key, f64, value[1]).Result()
			readInt = int64(f)
		}

	case "ZRANGEBYSCORE": // 멤버 스코어 증감
		var zrangeBy *redis.ZRangeBy
		zrangeBy = &redis.ZRangeBy{
			Max: value[0],
			Min: value[1],
		}
		if len(value[2]) > 0 && value[2] == "withscores" || value[2] == "WITHSCORES" {
			conn.ZRangeByScoreWithScores(*ctx, key, zrangeBy).Result()
		} else {
			conn.ZRangeByScore(*ctx, key, zrangeBy).Result()
			//readStrSlice
		}
	///////////////////////////////////////////////////////////////////////////////////////
	// COMMON
	///////////////////////////////////////////////////////////////////////////////////////
	case "EXPIRE": // 키 삭제 시간 지정
		var mytime int
		var exTime time.Duration
		mytime, err = strconv.Atoi(value[0])
		if err == nil {
			boolFlag = true
			exTime = time.Duration(mytime) * time.Second
			readBool, err = conn.Expire(*ctx, key, exTime).Result()
		}
	case "TTL": // 키 삭제 시간 조회
		keys = append(keys, "RESULT")
		readtime, err = conn.TTL(*ctx, key).Result()
		if err == nil {
			readStr = readtime.String()
		}
	case "KEYS": // 키 패턴 조회
		sliceFlag = true
		readStrSlice, err = conn.Keys(*ctx, key).Result()
	case "EXISTS": // 키 존재 여부 확인
		intFlag = true
		readInt, err = conn.Exists(*ctx, value...).Result()
	case "RENAME":
		keys = append(keys, "RESULT")
		readStr, err = conn.Rename(*ctx, key, value[0]).Result()
	case "RENAMENX":
		boolFlag = true
		readBool, err = conn.RenameNX(*ctx, key, value[0]).Result()
	case "PERSIST":
		boolFlag = true
		readBool, err = conn.Persist(*ctx, key).Result()
	case "OBJECT_ENCODE":
		keys = append(keys, "RESULT")
		readStr, err = conn.ObjectEncoding(*ctx, key).Result()
	case "OBJECT_IDLE":
		keys = append(keys, "RESULT")
		readtime, err = conn.ObjectIdleTime(*ctx, key).Result()
		if err == nil {
			readStr = readtime.String()
		}
	case "OBJECT_COUNT":
		intFlag = true
		readInt, err = conn.ObjectRefCount(*ctx, key).Result()
	case "TYPE":
		keys = append(keys, "RESULT")
		readStr, err = conn.Type(*ctx, key).Result()
	///////////////////////////////////////////////////////////////////////////////////////
	default:
		return nil, errors.New("NOT SUPPORTED REDIS COMMAND")
	}

	if expire > 0 && command != "EXPIRE" {
		conn.Expire(*ctx, key, expire).Result()
	}

	if err != nil { // 에러발생
		return nil, err
	} else if readInterface != nil { // REDIS interface 리턴
		if readInterface[0] == nil {
			return nil, errors.New("data is nil")
		}

		i := 0
		for _, val := range readInterface { // keys 는 위에서 설정하고 옴
			if readInterface[i] == nil {
				return nil, errors.New("data is nil")
			}
			vals = append(vals, val.(string))
			i++
		}
		//keys = value
	} else if len(readMap) != 0 { // REDIS []string 리턴
		for key, val := range readMap {
			keys = append(keys, key)
			vals = append(vals, val)
		}
	} else if sliceFlag { // REDIS []string 리턴
		if len(readStrSlice) <= 0 {
			return nil, errors.New("data is nil")
		}
		for i, val := range readStrSlice {
			keys = append(keys, "RESULT"+strconv.Itoa(i+1))
			vals = append(vals, val)
		}
	} else if boolFlag { // REDIS bool 리턴
		keys = append(keys, "RESULT")
		if readBool {
			vals = append(vals, "OK")
		} else {
			vals = append(vals, "NO")
		}
	} else if intFlag { // REDIS int 리턴
		keys = append(keys, "RESULT")
		sInt := fmt.Sprintf("%d", readInt)
		vals = make([]string, 0) // int 리턴은 어차피 응답값만 하니까 한번 초기화하자
		vals = append(vals, sInt)
		//keys = append(keys, "RESULT")
	} else if len(readStr) == 0 { // REDIS err 아니지만, 리턴값이 없을 때
		keys = append(keys, "RESULT")
		vals = append(vals, "OK")
		//keys = append(keys, "RESULT")
	} else {
		vals = append(vals, readStr)
		//keys = append(keys, "RESULT")
	}

	return PacketSetResponse(query, keys, vals)
}

/*
*******************************************************************************************
  - function	: PacketSetResponse
  - Description	: Redis 응답값 커스터마이징 (command|key|key|key|value|value|value|value|value|value...)
  - Argument	: [ ]
  - Return		: [ ]
  - Etc         :

*******************************************************************************************
*/
func PacketSetResponse(query []string, keys []string, vals []string) ([][]byte, error) {
	var lineBuff string
	var retnBuff [][]byte
	var keyValue map[string]interface{}

	for i := 0; i < len(vals); i++ {
		err := json.Unmarshal([]byte(vals[i]), &keyValue)
		if err == nil {
			for key, value := range keyValue {
				keys = append(keys, key)

				switch v := value.(type) {
				case string:
					vals = append(vals, value.(string))
				case int:
					vals = append(vals, fmt.Sprintf("%d", value.(int)))
				case float64:
					_, _float := math.Modf(value.(float64))
					if _float == 0 {
						vals = append(vals, fmt.Sprintf("%.0f", value.(float64)))
					} else {
						vals = append(vals, fmt.Sprintf("%f", value.(float64)))
					}
				default:
					vals = append(vals, v.(string))
				}
				//vals = append(vals, value.(string))
			}
		}
	}

	// 한줄씩 뽑아서 리턴
	retnBuff = make([][]byte, len(vals)/len(keys))
	queryString := strings.Join(query, " ") + "|"
	for i := 0; i < len(vals)/len(keys); i++ {
		retnBuff[i] = append(retnBuff[i], []byte(queryString)...)
		for _, key := range keys {
			lineBuff += key + "|"
		}
		for _, val := range vals {
			lineBuff += val + "|"
		}
		retnBuff[i] = append(retnBuff[i], []byte(lineBuff[:len(lineBuff)-1])...)
	}

	return retnBuff, nil
}
