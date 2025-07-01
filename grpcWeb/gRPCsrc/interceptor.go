package gRPCsrc

import (
	"context"
	"fmt"
	"grpcWeb/db"
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

// UnaryInterceptor: 인증 + 로깅 + 트레이싱
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
			log.Printf("🛑 Panic recovered: %v", r)
			log.Printf("📄 Stack trace:\n%s", string(debug.Stack()))

			// gRPC 내부 에러로 반환
			err = status.Errorf(codes.Internal, "서버 내부 오류가 발생했습니다.")
			log.Printf("[recover][%s]", err)
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
				log.Printf("[UNAUTHORIZED] [%s] 메서드: %s", traceID, info.FullMethod)
				return nil, status.Error(codes.Unauthenticated, "missing or invalid authorization token")
			}
			token := strings.TrimPrefix(authHeader[0], "Bearer ")

			redisCmd := fmt.Sprintf("GET %s", token)
			_, err = db.REDIS.RedisDo(&ctx, redisCmd)
			if err != nil {
				log.Printf("[UNAUTHORIZED] [%s] 인증실패: %s", traceID, info.FullMethod)
				return nil, status.Error(codes.Unauthenticated, err.Error()+" | "+"invalid token")
			}
		}
	}

	// --- 요청 처리 ---
	resp, err := handler(ctx, req)
	if err != nil {
		log.Printf("[gRPC][%s] %s | %v | 에러: %v", traceID, info.FullMethod, time.Since(start), err)
	} else {
		log.Printf("[gRPC][%s] %s | %v", traceID, info.FullMethod, time.Since(start))
	}

	return resp, err
}
