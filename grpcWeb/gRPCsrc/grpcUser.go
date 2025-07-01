package gRPCsrc

import (
	"context"
	"fmt"
	"grpcWeb/db"
	pb "grpcWeb/proto"
	"grpcWeb/web"
	"local/common"
	c "local/common"
	"strings"
	"time"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
}

// 회원가입
func (s *UserServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	c.Logging.Write(c.LogTRACE, "[Register] Start [%s]", req.UserId)

	dbinfo := setDB(req.DbType)
	hashed, _ := web.HashPassword(req.UserPw)
	sql := fmt.Sprintf("INSERT INTO users (UserID, UserPW) VALUES ('%s', '%s')", req.UserId, hashed)
	retn, err := dbinfo.Do(&ctx, sql)
	if err != nil {
		c.Logging.Write(c.LogERROR, "[Register] db.Do Error  [%v]", err.Error())
		return &pb.RegisterResponse{Success: false, Message: err.Error()}, err
	} else if !retn.Status {
		c.Logging.Write(c.LogERROR, "[Register] db.Do Failed [%v]", retn.Status)
		return &pb.RegisterResponse{Success: false, Message: "User already exists"}, fmt.Errorf("user already exists")
	}
	c.Logging.Write(c.LogDEBUG, "[Register] db.Do Return [%v]", retn.Tuples)
	c.Logging.Write(c.LogTRACE, "[Register] db.Do Success")
	c.Logging.Write(c.LogTRACE, "[Register] End")

	msg := fmt.Sprintf("%s Registered successfully", req.UserId)
	return &pb.RegisterResponse{Success: true, Message: msg}, nil
}

// 로그인
func (s *UserServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	c.Logging.Write(c.LogTRACE, "[Login] Start [%s]", req.UserId)

	dbinfo := setDB(req.DbType)
	sql := fmt.Sprintf("SELECT UserPW FROM users WHERE UserID = '%s'", req.UserId)
	retn, err := dbinfo.Do(&ctx, sql)
	if err != nil {
		c.Logging.Write(c.LogERROR, "[Login] db.Do Error  [%v]", err.Error())
		return &pb.LoginResponse{Message: err.Error()}, err
	} else if !retn.Status {
		c.Logging.Write(c.LogERROR, "[Login] db.Do Failed [%v]", retn.Status)
		return &pb.LoginResponse{Message: "User not found"}, fmt.Errorf("user not found")
	}
	c.Logging.Write(c.LogDEBUG, "[Register] db.Do Return [%v]", retn.Tuples)
	c.Logging.Write(c.LogTRACE, "[Login] db.Do Success")

	if retn.Tuples[0]["UserPW"] == "" {
		c.Logging.Write(c.LogERROR, "[Login] Can't Found User")
		return &pb.LoginResponse{Message: "User not found"}, fmt.Errorf("user not found")
	} else if !web.CheckPasswordHash(retn.Tuples[0]["UserPW"], req.UserPw) {
		c.Logging.Write(c.LogERROR, "[Login] Wrong Password")
		return &pb.LoginResponse{Message: "Invalid password"}, fmt.Errorf("invalid password")
	}
	c.Logging.Write(c.LogTRACE, "[Login] Password Check ok")

	token, _ := web.GenerateToken(req.UserId)
	c.Logging.Write(c.LogDEBUG, "[Login] Make JWT Token [%.]", token)
	query := fmt.Sprintf("SET %s %s", token, common.GetDateTime17())
	_, err = db.Redis.RedisDo(&ctx, query)
	if err != nil {
		c.Logging.Write(c.LogERROR, "[Login] Store JWT Token Error (Redis Err:%s)", err.Error())
		return &pb.LoginResponse{Message: err.Error()}, err
	}
	c.Logging.Write(c.LogTRACE, "[Login] Store JWT Token Success ")

	sql = fmt.Sprintf("UPDATE users SET LastLogin = '%s' WHERE UserID = '%s'", time.Now().Format("2006-01-02 15:04:05"), req.UserId)
	retn, err = dbinfo.Do(&ctx, sql)
	if err != nil {
		//return &pb.LoginResponse{Message: err.Error()}, err
	}

	return &pb.LoginResponse{Token: token, Message: "Login success"}, nil
}

// 사용자 정보 조회
func (s *UserServer) GetUserInfo(ctx context.Context, req *pb.UserInfoRequest) (*pb.UserInfoResponse, error) {
	c.Logging.Write(c.LogTRACE, "[GetUserInfo] Start [%s]", req.UserId)

	dbinfo := setDB(req.DbType)
	sql := fmt.Sprintf("SELECT Status, LastLogin FROM users WHERE UserID = '%s'", req.UserId)
	retn, err := dbinfo.Do(&ctx, sql)

	if err != nil {
		c.Logging.Write(c.LogERROR, "[GetUserInfo] db.Do Error  [%v]", err.Error())
		return &pb.UserInfoResponse{Status: false}, err
	} else if !retn.Status {
		c.Logging.Write(c.LogERROR, "[GetUserInfo] db.Do Failed [%v]", retn.Status)
		return nil, fmt.Errorf("user not found")
	}
	c.Logging.Write(c.LogDEBUG, "[GetUserInfo] db.Do Return [%v]", retn.Tuples)
	c.Logging.Write(c.LogTRACE, "[GetUserInfo] db.Do Success")
	c.Logging.Write(c.LogTRACE, "[GetUserInfo] End")

	return &pb.UserInfoResponse{Status: true, LastLogin: retn.Tuples[0]["LastLogin"]}, nil
	//return &pb.UserInfoResponse{Status: true, LastLogin: lastLogin}, nil
}

// 사용자 정보 수정
func (s *UserServer) UpdateUserInfo(ctx context.Context, req *pb.UpdateUserInfoRequest) (*pb.UpdateUserInfoResponse, error) {
	c.Logging.Write(c.LogTRACE, "[UpdateUserInfo] Start [%s]", req.UserId)

	dbinfo := setDB(req.DbType)
	sql := fmt.Sprintf("SELECT UserPW FROM users WHERE UserID = '%s'", req.UserId)
	retn, err := dbinfo.Do(&ctx, sql)
	if err != nil {
		c.Logging.Write(c.LogERROR, "[UpdateUserInfo] db.Do Error  [%v]", err.Error())
		return &pb.UpdateUserInfoResponse{Success: false, Message: err.Error()}, err
	} else if !web.CheckPasswordHash(retn.Tuples[0]["UserPW"], req.UserPw) {
		c.Logging.Write(c.LogERROR, "[UpdateUserInfo] Wrong Password")
		return &pb.UpdateUserInfoResponse{Success: false, Message: "Invalid password"}, fmt.Errorf("invalid password")
	}
	c.Logging.Write(c.LogTRACE, "[UpdateUserInfo] Select Password Success")

	hashed, _ := web.HashPassword(req.NewUserPw)
	sql = fmt.Sprintf("UPDATE users SET UserPW = '%s' WHERE UserID = '%s'", hashed, req.UserId)
	retn, err = dbinfo.Do(&ctx, sql)
	if err != nil {
		c.Logging.Write(c.LogERROR, "[UpdateUserInfo] Update Password Error  [%v]", err.Error())
		return &pb.UpdateUserInfoResponse{Success: false, Message: err.Error()}, err
	}
	c.Logging.Write(c.LogDEBUG, "[Register] db.Do Return [%v]", retn.Tuples)
	c.Logging.Write(c.LogTRACE, "[UpdateUserInfo] Update Password Success")
	c.Logging.Write(c.LogTRACE, "[UpdateUserInfo] End")

	return &pb.UpdateUserInfoResponse{Success: true, Message: "Password updated"}, nil
}

func setDB(dbtype string) *db.DBinfo {
	switch strings.ToLower(dbtype) {
	case "mysql", "mariadb":
		return &db.MySQL
	case "mssql", "sqlserver":
		return &db.MsSQL
	/*
		case "oracle":
			return &db.Oracle
	*/
	default:
		return nil
	}
}
