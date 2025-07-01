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
	hashed, _ := web.HashPassword(req.UserPw)

	dbinfo := setDB(req.DbType)
	sql := fmt.Sprintf("INSERT INTO users (UserID, UserPW) VALUES ('%s', '%s')", req.UserId, hashed)
	retn, err := dbinfo.Do(&ctx, sql)

	if err != nil {
		return &pb.RegisterResponse{Success: false, Message: err.Error()}, err
	} else if !retn.Status {
		return &pb.RegisterResponse{Success: false, Message: "User already exists"}, fmt.Errorf("user already exists")
	}

	return &pb.RegisterResponse{Success: true, Message: "Registered successfully"}, nil
}

// 로그인
func (s *UserServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	dbinfo := setDB(req.DbType)
	sql := fmt.Sprintf("SELECT UserPW FROM users WHERE UserID = '%s'", req.UserId)
	retn, err := dbinfo.Do(&ctx, sql)

	if err != nil {
		return &pb.LoginResponse{Message: err.Error()}, err
	} else if !retn.Status {
		return &pb.LoginResponse{Message: "User not found"}, fmt.Errorf("user not found")
	}

	if retn.Tuples[0]["UserPW"] == "" {
		return &pb.LoginResponse{Message: "User not found"}, fmt.Errorf("user not found")
	} else if !web.CheckPasswordHash(retn.Tuples[0]["UserPW"], req.UserPw) {
		return &pb.LoginResponse{Message: "Invalid password"}, fmt.Errorf("invalid password")
	}

	token, _ := web.GenerateToken(req.UserId)
	query := fmt.Sprintf("SET %s %s", token, common.GetDateTime17())
	_, err = db.Redis.RedisDo(&ctx, query)
	if err != nil {
		return &pb.LoginResponse{Message: err.Error()}, err
	}

	sql = fmt.Sprintf("UPDATE users SET LastLogin = '%s' WHERE UserID = '%s'", time.Now().Format("2006-01-02 15:04:05"), req.UserId)
	retn, err = dbinfo.Do(&ctx, sql)
	if err != nil {
		return &pb.LoginResponse{Message: err.Error()}, err
	}

	return &pb.LoginResponse{Token: token, Message: "Login success"}, nil
}

// 사용자 정보 조회
func (s *UserServer) GetUserInfo(ctx context.Context, req *pb.UserInfoRequest) (*pb.UserInfoResponse, error) {
	dbinfo := setDB(req.DbType)
	sql := fmt.Sprintf("SELECT Status, LastLogin FROM users WHERE UserID = '%s'", req.UserId)
	retn, err := dbinfo.Do(&ctx, sql)

	if err != nil {
		return &pb.UserInfoResponse{Status: false}, err
	} else if !retn.Status {
		return nil, fmt.Errorf("user not found")
	}
	fmt.Println(retn.Tuples)
	fmt.Println("LastLogin", retn.Tuples[0]["LastLogin"])

	return &pb.UserInfoResponse{Status: true, LastLogin: retn.Tuples[0]["LastLogin"]}, nil
	//return &pb.UserInfoResponse{Status: true, LastLogin: lastLogin}, nil
}

// 사용자 정보 수정
func (s *UserServer) UpdateUserInfo(ctx context.Context, req *pb.UpdateUserInfoRequest) (*pb.UpdateUserInfoResponse, error) {
	c.Logging.Write(c.LogTRACE, "[UpdateUserInfo] Start")
	dbinfo := setDB(req.DbType)
	sql := fmt.Sprintf("SELECT UserPW FROM users WHERE UserID = '%s'", req.UserId)
	retn, err := dbinfo.Do(&ctx, sql)
	if err != nil {
		return &pb.UpdateUserInfoResponse{Success: false, Message: err.Error()}, err
	} else if !web.CheckPasswordHash(retn.Tuples[0]["UserPW"], req.UserPw) {
		c.Logging.Write(c.LogERROR, "[UpdateUserInfo] Invalid password")
		return &pb.UpdateUserInfoResponse{Success: false, Message: "Invalid password"}, fmt.Errorf("invalid password")
	}

	hashed, _ := web.HashPassword(req.NewUserPw)
	sql = fmt.Sprintf("UPDATE users SET UserPW = '%s' WHERE UserID = '%s'", hashed, req.UserId)
	retn, err = dbinfo.Do(&ctx, sql)
	if err != nil {
		return &pb.UpdateUserInfoResponse{Success: false, Message: err.Error()}, err
	} else if !retn.Status {
		return &pb.UpdateUserInfoResponse{Success: false, Message: "User not found"}, fmt.Errorf("user not found")
	}

	return &pb.UpdateUserInfoResponse{Success: true, Message: "Password updated"}, nil
}

func setDB(dbtype string) *db.DBinfo {
	switch strings.ToLower(dbtype) {
	case "mysql", "mariadb":
		return &db.MySQL
	case "mssql", "sqlserver":
		return &db.MsSQL
	case "oracle":
		return &db.Oracle
	default:
		return nil
	}
}
