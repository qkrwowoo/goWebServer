package web

//
//import (
//	"errors"
//	"fmt"
//	"grpcWeb/db"
//)
//
///*
//*******************************************************************************************
//  - function	: Login
//  - Description	: 접속
//  - Argument	: [ (string)사용자 ID, (string)사용자 PW ]
//  - Return		: [ (string)JWT 토큰, (error)에러메세지 ]
//
//*******************************************************************************************
//*/
//func Login(userID, userPW string) (string, error) {
//	sql := fmt.Sprintf("SELECT status FROM %s WHERE user_id = '%s' AND password = '%s'",
//		db.MySQL.ID, userID, userPW)
//	retnMsg, err := db.MySQL.Query(sql)
//	if err != nil {
//		return err.Error(), err
//	} else if len(retnMsg) == 0 {
//		return "empty data", errors.New("empty data")
//	} else if retnMsg != "ok" {
//		return "status error", fmt.Errorf("status error")
//	}
//	return GenerateToken(userID)
//}
//
///*
//*******************************************************************************************
//  - function	: Register
//  - Description	: 회원가입
//  - Argument	: [ (string)사용자 ID, (string)사용자 PW ]
//  - Return		: [ (string)JWT 토큰, (error)에러메세지 ]
//
//*******************************************************************************************
//*/
//func Register(userID, userPW string) (string, error) {
//	sql := fmt.Sprintf("INSERT INTO users (user_id, password, status) VALUES ('%s', '%s', 'ok')", userID, userPW)
//	retnMsg, err := db.MySQL.Query(sql)
//	if err != nil {
//		return err.Error(), err
//	} else if len(retnMsg) == 0 {
//		return "empty data", errors.New("empty data")
//	}
//	return GenerateToken(userID)
//}
//
///*
//*******************************************************************************************
//  - function	: GetUserInfo
//  - Description	: 회원정보 조회
//  - Argument	: [ (string)사용자 ID ]
//  - Return		: [ (string)조회 데이터, (error)에러메세지 ]
//
//*******************************************************************************************
//*/
//func GetUserInfo(userID string) (string, error) {
//	sql := fmt.Sprintf("SELECT user_id, password, status FROM users WHERE user_id = '%s'", userID)
//	retnMsg, err := db.MySQL.Query(sql)
//	if err != nil {
//		return err.Error(), err
//	} else if len(retnMsg) == 0 {
//		return "empty data", errors.New("empty data")
//	}
//	return retnMsg, nil
//}
//
///*
//*******************************************************************************************
//  - function	: UpdateUserInfo
//  - Description	: 회원정보 갱신
//  - Argument	: [ (string)사용자 ID, (string)사용자 PW ]
//  - Return		: [ (string)갱신 여부 , (error)에러메세지 ]
//
//*******************************************************************************************
//*/
//func UpdateUserInfo(userID, userPW string) (string, error) {
//	sql := fmt.Sprintf("UPDATE users SET password = '%s' WHERE user_id = '%s'", userPW, userID)
//	retnMsg, err := db.MySQL.Query(sql)
//	if err != nil {
//		return err.Error(), err
//	} else if len(retnMsg) == 0 {
//		return "empty data", errors.New("empty data")
//	}
//	return retnMsg, nil
//}
