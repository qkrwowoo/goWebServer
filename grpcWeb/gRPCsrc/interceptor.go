package gRPCsrc

import (
	"context"
	"fmt"
	"grpcWeb/db"
	c "local/common"
	"log"
	"runtime/debug"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var noAuthMethods = map[string]bool{
	"/auth.UserService/Register": true,
	"/auth.UserService/Login":    true,
}

/*
*******************************************************************************************
  - function	: UnaryInterceptor
  - Description	: gRPC 이벤트 처리 미들웨어 (auth, logging, error, trace)
  - Argument	: [ ]
  - Return		: [ ]
  - Etc         :

*******************************************************************************************
*/
func UnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			// 패닉 메시지 출력
			err = status.Errorf(codes.Internal, "서버 내부 오류가 발생했습니다.") // gRPC 내부 에러로 반환
			c.Logging.Write(c.LogALL, "[PANIC] Method[%s] [%v]", info.FullMethod, r)
			c.Logging.Write(c.LogERROR, "[PANIC] Method[%s] error[%s] [%s]", info.FullMethod, err, string(debug.Stack()))
		}
	}()

	start := time.Now()

	// --- 트레이싱 ID 추적 ---
	traceID := ""
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if val, found := md["trace-id"]; found && len(val) > 0 {
			traceID = val[0]
		} else {
			traceID = "trace-" + time.Now().Format("150405.000")
		}
	} else {
		traceID = "trace-" + time.Now().Format("150405.000")
	}
	ctx = context.WithValue(ctx, "trace-id", traceID)

	// --- 인증 확인 ---
	if !noAuthMethods[info.FullMethod] {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			authHeader := md["authorization"]
			if len(authHeader) == 0 || !strings.HasPrefix(authHeader[0], "Bearer ") {
				c.Logging.Write(c.LogERROR, "[UNAUTHORIZED][%s] Method[%s] missing authorization token", traceID, info.FullMethod)
				return nil, status.Error(codes.Unauthenticated, "missing authorization token")
			}
			token := strings.TrimPrefix(authHeader[0], "Bearer ")

			redisCmd := fmt.Sprintf("GET %s", token)
			_, err = db.REDIS.RedisDo(&ctx, redisCmd)
			if err != nil {
				c.Logging.Write(c.LogERROR, "[UNAUTHORIZED][%s] Method[%s] invalid authorization token", traceID, info.FullMethod)
				return nil, status.Error(codes.Unauthenticated, err.Error()+" | "+"invalid token")
			}
		}
	}

	// --- 요청 처리 ---
	resp, err := handler(ctx, req)
	if err != nil {
		log.Printf("[gRPC][%s] Method[%s] | [%v] | error: [%v]", traceID, info.FullMethod, time.Since(start), err)
	} else {
		log.Printf("[gRPC][%s] Method[%s] | [%v]", traceID, info.FullMethod, time.Since(start))
	}

	return resp, err
}
