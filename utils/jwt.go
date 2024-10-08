package utils

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

var jwtKey = []byte(viper.GetString("jwt.secret"))

type Claims struct {
	Id       uint   `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateToken(id uint, username string) (string, error) {
	nowTime := time.Now()

	// 计算过期时间
	expireHours := viper.GetInt("jwt.expire_hours")
	expireTime := nowTime.Add(time.Duration(expireHours) * time.Hour)

	claims := &Claims{
		id,
		username,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			Issuer:    "go-tiny",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func GetClaims(c *gin.Context) (*Claims, error) {
	userInfo, ok := c.Get("claims")
	if !ok {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return nil, errors.New("Unauthorized")
	}
	return userInfo.(*Claims), nil
}
