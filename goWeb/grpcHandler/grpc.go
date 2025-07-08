package grpcHandler

import (
	"context"
	c "local/common"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var Conn *grpc.ClientConn
var IPADDR string

func init() {
}

/*
*******************************************************************************************
  - function	: Open_gRPC_Session
  - Description	: gRPC 세션 접속
  - Argument	: [ ]
  - Return		: [ (error) 오류 ]
  - Etc         :

*******************************************************************************************
*/
func Open_gRPC_Session() error {
	var err error
	Conn = nil
	Conn, err = grpc.NewClient(IPADDR,
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`), // gRPC 로드밸런싱
		grpc.WithTransportCredentials(insecure.NewCredentials()),               // gRPC TLS 설정
		//grpc.WithUnaryInterceptor(AuthInterceptor("")),                         // gRPC JWT (단일요청)
		//grpc.WithTransportCredentials(insecure.NewCredentials()),               // (테스트용) 인증없이 연결
		//grpc.WithStreamInterceptor(),                                           // gRPC JWT (스트리밍(양방향)))
	)

	//Idle	(IDLE)초기 상태. 연결을 맺고 있지 않음. RPC 요청이 발생하면 연결 시도 시작.
	//Connecting	(CONNECTING)서버에 연결 중. 아직 연결이 완료되지 않음.
	//Ready	(READY)서버에 연결 성공. RPC 호출이 가능함.
	//TransientFailure	(TRANSIENT_FAILURE)일시적 오류로 연결이 끊어진 상태. gRPC가 백오프(backoff)를 두고 재연결을 시도 중.
	//Shutdown	(SHUTDOWN)Close()가 호출되어 연결이 종료됨. 이후 상태 변화 없음.

	if Conn.GetState().String() != "READY" && Conn.GetState().String() != "IDLE" {
		c.Logging.Write(c.LogWARN, "gRPC Connect Failed [%s]", Conn.GetState().String())
	} else {
		c.Logging.Write(c.LogTRACE, "gRPC Connect Success [%s]", Conn.GetState().String())
	}
	return err
}

/*
*******************************************************************************************
  - function	: Close_gRPC_Session
  - Description	: gRPC 세션 해제
  - Argument	: [ ]
  - Return		: [ (error) 오류 ]
  - Etc         :

*******************************************************************************************
*/
func Close_gRPC_Session() {
	Conn.Close()
}

// client 가 http 헤더에 jwt 담아주기로
func AuthInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		md := metadata.Pairs("authorization", "Bearer "+token)
		ctx = metadata.NewOutgoingContext(ctx, md)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
