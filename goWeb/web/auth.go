package web

import (
	"github.com/gin-gonic/gin"
)

var jwtSecret = []byte("mysecretkey") // 암호화 키

/*
*******************************************************************************************
  - function	: AuthMiddleware
  - Description	: 개인정보 조회.
  - Argument	: [ (string)JWT 토큰 ]
  - Return		: [ (string)JWT 토큰, (error)에러메세지 ]

*******************************************************************************************
*/
// 개인정보 조회할 때 사용자 인증용
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			c.JSON(401, gin.H{"error": "Authorization header missing or malformed"})
			c.Abort()
			return
		}
		/*
			tokenString := authHeader[7:]

			claims, err := ParseToken(tokenString)
			if err != nil {
				c.JSON(401, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}

			// 사용자 정보 context에 저장
			c.Set("username", claims["username"])
			c.Next()
		*/
	}
}
