package gRPCsrc

import (
	"context"
	"fmt"
	"grpcWeb/db"
	pb "grpcWeb/proto"
	"grpcWeb/web"
	c "local/common"
	"time"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
}

/*
*******************************************************************************************
  - function	: Register
  - Description	: 회원가입
  - Argument	: [ (context.Conetxt) TIMEOUT설정, (*pb.RegisterRequest) gRPC 요청 ]
  - Return		: [ (*pb.RegisterResponse) gRPC 응답, (error) 오류]
  - Etc         :

*******************************************************************************************
*/
func (s *UserServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	c.Logging.Write(c.LogTRACE, "[Register] Start [%s]", req.UserId)

	// hash 비밀번호 생성 (암호화)
	hashed, _ := web.HashPassword(req.UserPw)
	sql := fmt.Sprintf("INSERT INTO users (UserID, UserPW) VALUES ('%s', '%s')", req.UserId, hashed)

	// 회원가입 DB 등록
	retn, err := db.RDB.Do(&ctx, sql)
	if err != nil {
		c.Logging.Write(c.LogERROR, "[Register] db.Do Error  [%v]", err.Error())
		return &pb.RegisterResponse{Success: false, Message: err.Error()}, err
	} else if !retn.Status {
		c.Logging.Write(c.LogERROR, "[Register] db.Do Failed [%v]", retn.Status)
		return &pb.RegisterResponse{Success: false, Message: "User already exists"}, fmt.Errorf("user already exists")
	}

	// log...
	c.Logging.Write(c.LogDEBUG, "[Register] db.Do Return [%v]", retn.Tuples)
	c.Logging.Write(c.LogTRACE, "[Register] db.Do Success")
	c.Logging.Write(c.LogTRACE, "[Register] End")
	msg := fmt.Sprintf("%s Registered successfully", req.UserId)
	return &pb.RegisterResponse{Success: true, Message: msg}, nil
}

/*
*******************************************************************************************
  - function	: Login
  - Description	: 로그인
  - Argument	: [ (context.Conetxt) TIMEOUT설정, (*pb.RegisterRequest) gRPC 요청 ]
  - Return		: [ (*pb.RegisterResponse) gRPC 응답, (error) 오류]
  - Etc         :

*******************************************************************************************
*/
func (s *UserServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	c.Logging.Write(c.LogTRACE, "[Login] Start [%s]", req.UserId)

	// 비밀번호 조회
	sql := fmt.Sprintf("SELECT UserPW FROM users WHERE UserID = '%s'", req.UserId)
	retn, err := db.RDB.Do(&ctx, sql)
	if err != nil {
		c.Logging.Write(c.LogERROR, "[Login] db.Do Error  [%v]", err.Error())
		return &pb.LoginResponse{Message: err.Error()}, err
	} else if !retn.Status {
		c.Logging.Write(c.LogERROR, "[Login] db.Do Failed [%v]", retn.Status)
		return &pb.LoginResponse{Message: "User not found"}, fmt.Errorf("user not found")
	}
	c.Logging.Write(c.LogDEBUG, "[Register] db.Do Return [%v]", retn.Tuples)
	c.Logging.Write(c.LogTRACE, "[Login] db.Do Success")

	// 비밀번호 hash 일치여부 확인
	if retn.Tuples[0]["UserPW"] == "" {
		c.Logging.Write(c.LogERROR, "[Login] Can't Found User")
		return &pb.LoginResponse{Message: "User not found"}, fmt.Errorf("user not found")
	} else if !web.CheckPasswordHash(retn.Tuples[0]["UserPW"], req.UserPw) {
		c.Logging.Write(c.LogERROR, "[Login] Wrong Password")
		return &pb.LoginResponse{Message: "Invalid password"}, fmt.Errorf("invalid password")
	}
	c.Logging.Write(c.LogTRACE, "[Login] Password Check ok")

	// JWT 토큰 발급
	token, _ := web.GenerateToken(req.UserId)
	c.Logging.Write(c.LogDEBUG, "[Login] Make JWT Token [%.]", token)
	query := fmt.Sprintf("SET %s %s", token, c.GetDateTime17())
	_, err = db.REDIS.RedisDo(&ctx, query)
	if err != nil {
		c.Logging.Write(c.LogERROR, "[Login] Store JWT Token Error (Redis Err:%s)", err.Error())
		return &pb.LoginResponse{Message: err.Error()}, err
	}
	c.Logging.Write(c.LogTRACE, "[Login] Store JWT Token Success ")

	// 마지막 접속시간 갱신
	sql = fmt.Sprintf("UPDATE users SET LastLogin = '%s' WHERE UserID = '%s'", time.Now().Format("2006-01-02 15:04:05"), req.UserId)
	retn, err = db.RDB.Do(&ctx, sql)
	if err != nil { // 갱신용이라 오류는 무시
		//return &pb.LoginResponse{Message: err.Error()}, err
	}
	return &pb.LoginResponse{Token: token, Message: "Login success"}, nil
}

/*
*******************************************************************************************
  - function	: GetUserInfo
  - Description	: 사용자 조회
  - Argument	: [ (context.Conetxt) TIMEOUT설정, (*pb.RegisterRequest) gRPC 요청 ]
  - Return		: [ (*pb.RegisterResponse) gRPC 응답, (error) 오류]
  - Etc         :

*******************************************************************************************
*/
func (s *UserServer) GetUserInfo(ctx context.Context, req *pb.UserInfoRequest) (*pb.UserInfoResponse, error) {
	c.Logging.Write(c.LogTRACE, "[GetUserInfo] Start [%s]", req.UserId)

	// 사용자 조회
	sql := fmt.Sprintf("SELECT Status, LastLogin FROM users WHERE UserID = '%s'", req.UserId)
	retn, err := db.RDB.Do(&ctx, sql)
	if err != nil {
		c.Logging.Write(c.LogERROR, "[GetUserInfo] db.Do Error  [%v]", err.Error())
		return &pb.UserInfoResponse{Status: false}, err
	} else if !retn.Status {
		c.Logging.Write(c.LogERROR, "[GetUserInfo] db.Do Failed [%v]", retn.Status)
		return nil, fmt.Errorf("user not found")
	}

	// log...
	c.Logging.Write(c.LogDEBUG, "[GetUserInfo] db.Do Return [%v]", retn.Tuples)
	c.Logging.Write(c.LogTRACE, "[GetUserInfo] db.Do Success")
	c.Logging.Write(c.LogTRACE, "[GetUserInfo] End")
	return &pb.UserInfoResponse{Status: true, LastLogin: retn.Tuples[0]["LastLogin"]}, nil
	//return &pb.UserInfoResponse{Status: true, LastLogin: lastLogin}, nil
}

/*
*******************************************************************************************
  - function	: UpdateUserInfo
  - Description	: 사용자 갱신
  - Argument	: [ (context.Conetxt) TIMEOUT설정, (*pb.RegisterRequest) gRPC 요청 ]
  - Return		: [ (*pb.RegisterResponse) gRPC 응답, (error) 오류]
  - Etc         :

*******************************************************************************************
*/
func (s *UserServer) UpdateUserInfo(ctx context.Context, req *pb.UpdateUserInfoRequest) (*pb.UpdateUserInfoResponse, error) {
	c.Logging.Write(c.LogTRACE, "[UpdateUserInfo] Start [%s]", req.UserId)

	// 사용자 조회
	sql := fmt.Sprintf("SELECT UserPW FROM users WHERE UserID = '%s'", req.UserId)
	retn, err := db.RDB.Do(&ctx, sql)
	if err != nil {
		c.Logging.Write(c.LogERROR, "[UpdateUserInfo] db.Do Error  [%v]", err.Error())
		return &pb.UpdateUserInfoResponse{Success: false, Message: err.Error()}, err
	} else if !web.CheckPasswordHash(retn.Tuples[0]["UserPW"], req.UserPw) {
		c.Logging.Write(c.LogERROR, "[UpdateUserInfo] Wrong Password")
		return &pb.UpdateUserInfoResponse{Success: false, Message: "Invalid password"}, fmt.Errorf("invalid password")
	}
	c.Logging.Write(c.LogTRACE, "[UpdateUserInfo] Select Password Success")

	// 비밀번호 변경
	hashed, _ := web.HashPassword(req.NewUserPw)
	sql = fmt.Sprintf("UPDATE users SET UserPW = '%s' WHERE UserID = '%s'", hashed, req.UserId)
	retn, err = db.RDB.Do(&ctx, sql)
	if err != nil {
		c.Logging.Write(c.LogERROR, "[UpdateUserInfo] Update Password Error  [%v]", err.Error())
		return &pb.UpdateUserInfoResponse{Success: false, Message: err.Error()}, err
	}

	// log...
	c.Logging.Write(c.LogDEBUG, "[Register] db.Do Return [%v]", retn.Tuples)
	c.Logging.Write(c.LogTRACE, "[UpdateUserInfo] Update Password Success")
	c.Logging.Write(c.LogTRACE, "[UpdateUserInfo] End")
	return &pb.UpdateUserInfoResponse{Success: true, Message: "Password updated"}, nil
}
