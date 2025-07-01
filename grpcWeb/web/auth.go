package web

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("mysecretkey") // 암호화 키

/*
*******************************************************************************************
  - function	: GenerateToken
  - Description	: 토큰키 생성
  - Argument	: [ (string)사용자 ID ]
  - Return		: [ (string)JWT 토큰, (error)에러메세지 ]

*******************************************************************************************
*/
func GenerateToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 1).Unix(), // 만료시간 1시간
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

/*
*******************************************************************************************
  - function	: ParseToken
  - Description	: 토큰키 검사
  - Argument	: [ (string)JWT 토큰 ]
  - Return		: [ (string)JWT 토큰, (error)에러메세지 ]

*******************************************************************************************
*/
func ParseToken(tokenString string) (jwt.MapClaims, error) {
	// 해쉬 암호화 한 게 다시 들어오고, 복호화 된 token 이 jwtSecret 이 되는지 검증
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return token.Claims.(jwt.MapClaims), nil
}

/*
*******************************************************************************************
  - function	: HashPassword
  - Description	: 비밀번호 Hash 생성
  - Argument	: [ (string)비밀번호 ]
  - Return		: [ (string)Hash, (error)에러메세지 ]

*******************************************************************************************
*/
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

/*
*******************************************************************************************
  - function	: CheckPasswordHash
  - Description	: 비밀번호 Hash 체크
  - Argument	: [ (string)비밀번호, (string)Hash ]
  - Return		: [ (bool)정상여부 ]

*******************************************************************************************
*/
func CheckPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
