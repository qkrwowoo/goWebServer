package ginHandler

import (
	"context"
	pb "goWeb/proto"
	c "local/common"
	"time"

	"github.com/gin-gonic/gin"
)

/*
*******************************************************************************************
  - function	: UserInfo
  - Description	: /user/... 라우팅
  - Argument	: [ (*gin.Engine) gin Router 정보 ]
  - Return		: [ ]
  - Etc         :

*******************************************************************************************
*/
func UserInfo(r *gin.Engine) {
	user := r.Group("user")
	{
		user.POST("/get", func(g *gin.Context) { get_UserInfo(g) })
		user.POST("/update", func(g *gin.Context) { update_UserInfo(g) })
	}
}

/*
*******************************************************************************************
  - function	: Register
  - Description	: 회원가입 API 호출
  - Argument	: [ (*gin.Engine) gin Router 정보 ]
  - Return		: [ ]
  - Etc         :

*******************************************************************************************
*/
func Register(r *gin.Engine) {
	r.POST("/register", func(g *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		c.Logging.Write(c.LogTRACE, "[Register] start")

		var req map[string]interface{}
		if err := g.BindJSON(&req); err != nil {
			returnResponse(g, 400, gin.H{"error": "Invalid request"})
			return
		} else if _, ok := req["UserID"].(string); !ok {
			returnResponse(g, 400, gin.H{"error": "Invalid request UserID"})
			return
		} else if _, ok := req["UserPW"].(string); !ok {
			returnResponse(g, 400, gin.H{"error": "Invalid request UserPW"})
			return
		} else if _, ok := req["DbType"].(string); !ok {
			returnResponse(g, 400, gin.H{"error": "Invalid request DbType"})
			return
		}

		client, ctx, err := getClient(false, g, ctx)
		if err != nil {
			returnResponse(g, 500, gin.H{"error": err.Error()})
			return
		}

		res, err := client.Register(ctx, &pb.RegisterRequest{
			UserId: req["UserID"].(string),
			UserPw: req["UserPW"].(string),
			DbType: req["DbType"].(string),
		})
		if err != nil {
			returnResponse(g, 500, gin.H{"error": err.Error()})
		} else if !res.Success {
			returnResponse(g, 400, gin.H{"error": res.Message})
		} else {
			returnResponse(g, 200, gin.H{"message": res.Message})
		}
		c.Logging.Write(c.LogTRACE, "[Register] end")
	})
}

/*
*******************************************************************************************
  - function	: Login
  - Description	: 로그인 API 호출
  - Argument	: [ (*gin.Engine) gin Router 정보 ]
  - Return		: [ ]
  - Etc         :

*******************************************************************************************
*/
func Login(r *gin.Engine) {
	r.POST("/login", func(g *gin.Context) {
		c.Logging.Write(c.LogTRACE, "[Login] Start")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var req map[string]interface{}
		if err := g.BindJSON(&req); err != nil {
			returnResponse(g, 400, gin.H{"error": "Invalid request"})
			return
		} else if _, ok := req["UserID"].(string); !ok {
			returnResponse(g, 400, gin.H{"error": "Invalid request UserID"})
			return
		} else if _, ok := req["UserPW"].(string); !ok {
			returnResponse(g, 400, gin.H{"error": "Invalid request UserPW"})
			return
		} else if _, ok := req["DbType"].(string); !ok {
			returnResponse(g, 400, gin.H{"error": "Invalid request DbType"})
			return
		}

		client, ctx, err := getClient(false, g, ctx)
		if err != nil {
			returnResponse(g, 500, gin.H{"error": err.Error()})
			return
		}

		res, err := client.Login(ctx, &pb.LoginRequest{
			UserId: req["UserID"].(string),
			UserPw: req["UserPW"].(string),
			DbType: req["DbType"].(string),
		})
		if err != nil {
			returnResponse(g, 500, gin.H{"error": err.Error()})
		} else {
			returnResponse(g, 200, gin.H{"message": res.Token})
		}
		c.Logging.Write(c.LogTRACE, "[Login] End")
	})
}

/*
*******************************************************************************************
  - function	: get_UserInfo
  - Description	: 회원정보 조회 API 호출
  - Argument	: [ (*gin.Engine) gin Router 정보 ]
  - Return		: [ ]
  - Etc         :

*******************************************************************************************
*/
func get_UserInfo(g *gin.Context) {
	c.Logging.Write(c.LogTRACE, "[GetUserInfo] Start")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var req map[string]interface{}
	if err := g.BindJSON(&req); err != nil {
		returnResponse(g, 400, gin.H{"error": "Invalid request"})
		return
	} else if _, ok := req["UserID"].(string); !ok {
		returnResponse(g, 400, gin.H{"error": "Invalid request UserID"})
		return
	}

	client, ctx, err := getClient(true, g, ctx)
	if err != nil {
		g.JSON(500, gin.H{"error": err.Error()})
		returnResponse(g, 500, gin.H{"error": err.Error()})
		return
	}

	res, err := client.GetUserInfo(ctx, &pb.UserInfoRequest{
		UserId: req["UserID"].(string),
		DbType: req["DbType"].(string),
	})
	if err != nil {
		returnResponse(g, 500, gin.H{"error": err.Error()})
	} else if !res.Status {
		returnResponse(g, 400, gin.H{"error": "User not found"})
	} else {
		returnResponse(g, 200, gin.H{"Status": res.Status, "LastLogin": res.LastLogin})
	}
	c.Logging.Write(c.LogTRACE, "[GetUserInfo] End")
}

/*
*******************************************************************************************
  - function	: update_UserInfo
  - Description	: 회원정보 갱신 API 호출
  - Argument	: [ (*gin.Engine) gin Router 정보 ]
  - Return		: [ ]
  - Etc         :

*******************************************************************************************
*/
func update_UserInfo(g *gin.Context) {
	c.Logging.Write(c.LogTRACE, "[UpdateUserInfo] Start")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var req map[string]interface{}
	if err := g.BindJSON(&req); err != nil {
		returnResponse(g, 400, gin.H{"error": "Invalid request"})
		return
	} else if _, ok := req["UserID"].(string); !ok {
		returnResponse(g, 400, gin.H{"error": "Invalid request UserID"})
		return
	} else if _, ok := req["UserPW"].(string); !ok {
		returnResponse(g, 400, gin.H{"error": "Invalid request UserPW"})
		return
	} else if _, ok := req["NewUserPw"].(string); !ok {
		returnResponse(g, 400, gin.H{"error": "Invalid request NewUserPw"})
		return
	} else if _, ok := req["DbType"].(string); !ok {
		returnResponse(g, 400, gin.H{"error": "Invalid request DbType"})
		return
	}

	client, ctx, err := getClient(true, g, ctx)
	if err != nil {
		g.JSON(500, gin.H{"error": err.Error()})
		return
	}

	res, err := client.UpdateUserInfo(ctx, &pb.UpdateUserInfoRequest{
		UserId: req["UserID"].(string),
		UserPw: req["UserPW"].(string),
		DbType: req["DbType"].(string),
	})
	if err != nil {
		g.JSON(500, gin.H{"error": err.Error()})
	} else if !res.Success {
		g.JSON(400, gin.H{"error": res.Message})
	} else {
		g.JSON(200, gin.H{"message": res.Message})
	}
	c.Logging.Write(c.LogTRACE, "[UpdateUserInfo] End")
}
