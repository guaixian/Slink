package middleware

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// jwtKey 由环境变量 SLINK_JWT_SECRET 提供；未设置时使用内置仅适用于本地/开发环境。
var jwtKey []byte

func init() {
	if s := os.Getenv("SLINK_JWT_SECRET"); s != "" {
		jwtKey = []byte(s)
	} else {
		jwtKey = []byte("your_secret_key")
	}
}

type Claims struct {
	UserID uint
	Email  string
	jwt.RegisteredClaims
}

func GenerateToken(userID uint, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供Token"})
			c.Abort()
			return
		}

		// 检查并提取 Bearer token
		const bearerPrefix = "Bearer "
		if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token格式错误，应为 Bearer token"})
			c.Abort()
			return
		}

		token := authHeader[len(bearerPrefix):]
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token为空"})
			c.Abort()
			return
		}

		claims, err := ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token无效"})
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email) // 兼容：与 JWT 字段同名，存的是登录账号
		c.Set("account", claims.Email)
		c.Next()
	}
}
