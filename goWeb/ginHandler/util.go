package ginHandler

import (
	"bytes"
	"context"
	"fmt"
	"goWeb/grpcHandler"
	pb "goWeb/proto"
	"io"
	c "local/common"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

/*
*******************************************************************************************
  - function	: returnResponse
  - Description	: Client 응답
  - Argument	: [ (*gin.Engine) gin Router 정보, (int) HTTP 응답코드, (map[string]any) 응답데이터 ]
  - Return		: [ ]
  - Etc         :

*******************************************************************************************
*/
func returnResponse(g *gin.Context, code int, response map[string]any) {
	c.Logging.Write(c.LogTRACE, "%s", c.Json2String(response))
	g.JSON(code, response)
}

/*
*******************************************************************************************
  - function	: GetTokenFromHeader
  - Description	: JWT 조회
  - Argument	: [ (*gin.Engine) gin Router 정보 ]
  - Return		: [ (string) JWT ]
  - Etc         :

*******************************************************************************************
*/
func GetTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// 예: "Bearer abc.def.ghi" 라면, "abc.def.ghi"만 추출
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

/*
*******************************************************************************************
  - function	: AddJwtTokenToContext
  - Description	: gRPC 요청 헤더에 JWT 추가
  - Argument	: [ (context.Context) Timeout 정보, (*gin.Engine) gin Router 정보 ]
  - Return		: [ (context.Context) Timeout 정보, (error) 오류 ]
  - Etc         :

*******************************************************************************************
*/
func AddJwtTokenToContext(ctx context.Context, c *gin.Context) (context.Context, error) {
	token := GetTokenFromHeader(c)
	if token == "" {
		return nil, fmt.Errorf("token is empty")
	} else {
		md := metadata.Pairs("authorization", "Bearer "+token)
		return metadata.NewOutgoingContext(ctx, md), nil
	}
}

/*
*******************************************************************************************
  - function	: getClient
  - Description	: grpcHandler.Conn 정상여부 확인 & JWT 추가
  - Argument	: [ (bool) JWT 확인 API 여부, (*gin.Engine) gin Router 정보, (context.Context) Timeout 정보 ]
  - Return		: [ (pb.UserServiceClient) gRPC 서비스, (context.Context) Timeout 정보 , (error) 오류]
  - Etc         : gRPC 연결 상태가 SHUTDOWN 이라면 재연결 시도.

*******************************************************************************************
*/
func getClient(chk_token bool, g *gin.Context, ctx context.Context) (pb.UserServiceClient, context.Context, error) {
	var err error
	c.Logging.Write(c.LogDEBUG, "getClient [%s]", grpcHandler.Conn.GetState().String())
	if grpcHandler.Conn != nil && grpcHandler.Conn.GetState().String() == "SHUTDOWN" {
		grpcHandler.Conn.Close()
		if err := grpcHandler.Open_gRPC_Session(); err != nil {
			return nil, nil, err
		}
	}

	if chk_token {
		ctx, err = AddJwtTokenToContext(ctx, g)
		if err != nil {
			return nil, nil, err
		}
	}
	return pb.NewUserServiceClient(grpcHandler.Conn), ctx, nil
}

type loggingResponseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)                  // 응답 본문 저장
	return w.ResponseWriter.Write(b) // 실제 클라이언트에 전송
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

/*
*******************************************************************************************
  - function	: LoggerMiddleware
  - Description	: gin 트랜잭션 Logging
  - Argument	: [ ]
  - Return		: [ (gin.HandlerFunc) gin 미들웨어 ]
  - Etc         :

*******************************************************************************************
*/
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		var reqBody []byte
		if c.Request.Body != nil {
			reqBody, _ = io.ReadAll(c.Request.Body)
			// Body는 한 번만 읽을 수 있으므로 복원해줘야 함
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		lrw := &loggingResponseWriter{
			ResponseWriter: c.Writer,
			body:           new(bytes.Buffer),
			statusCode:     http.StatusOK,
		}
		c.Writer = lrw

		// 요청 시작
		log.Printf(
			"[REQUEST]  Method[%s] URL[%s] IP[%s] Body[%s]",
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
			strings.TrimSpace(string(reqBody)),
		)
		c.Next()

		// 응답 시작
		log.Printf(
			"[RESPONSE] Method[%s] URL[%s] Time[%s] (%d)[%s]",
			c.Request.Method,
			c.Request.URL.Path,
			time.Since(start),
			lrw.statusCode,
			strings.TrimSpace(lrw.body.String()),
		)
	}
}

// gRPC 에서 JWT 검사하는걸로
//func AuthMiddleware() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		// 요청 헤더에서 API 키를 가져옵니다.
//		apiKey := c.GetHeader("X-API-KEY")
//
//		if apiKey == "" {
//			// API 키가 없으면 요청을 중단하고 401 Unauthorized 에러를 반환합니다.
//			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key is required"})
//			return
//		}
//
//		username, ok := userDatabase[apiKey]
//		if !ok {
//			// 유효하지 않은 API 키이면 요청을 중단합니다.
//			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
//			return
//		}
//
//		// 컨텍스트(Context)에 사용자 정보를 저장하여 다음 핸들러에서 사용할 수 있도록 합니다.
//		c.Set("username", username)
//
//		// 인증에 성공했으므로 다음 핸들러로 제어를 넘깁니다.
//		c.Next()
//	}
//}
