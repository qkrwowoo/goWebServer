package grpcHandler

import (
	"context"
	pb "goWeb/proto"
	c "local/common"

	"github.com/gin-gonic/gin"
)

func getClient(chk_token bool, g *gin.Context, ctx context.Context) (pb.UserServiceClient, context.Context, error) {
	var err error
	c.Logging.Write(c.LogDEBUG, "getClient [%s]", gRPC_Conn.GetState().String())
	//if gRPC_Conn != nil && (gRPC_Conn.GetState().String() == "SHUTDOWN" || gRPC_Conn.GetState().String() == "TRANSIENT_FAILURE") {
	if gRPC_Conn != nil && gRPC_Conn.GetState().String() == "SHUTDOWN" {
		gRPC_Conn.Close()
		if err := Open_gRPC_Session(); err != nil {
			return nil, nil, err
		}
	}

	if chk_token {
		ctx, err = AddJwtTokenToContext(ctx, g)
		if err != nil {
			return nil, nil, err
		}
	}
	return pb.NewUserServiceClient(gRPC_Conn), ctx, nil
}

func returnResponse(g *gin.Context, code int, response map[string]any) {
	c.Logging.Write(c.LogTRACE, "%s", c.Json2String(response))
	g.JSON(code, response)
}
