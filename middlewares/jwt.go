package middlewares

import (
	"time"

	"errors"

	"github.com/Liv1020/move-car/components"
	"github.com/Liv1020/move-car/models"
	"github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
)

// JwtMiddleware JwtMiddleware
var JwtMiddleware *jwt.GinJWTMiddleware

func init() {
	// the jwt middleware
	JwtMiddleware = &jwt.GinJWTMiddleware{
		Realm:      "test zone",
		Key:        []byte("secret key"),
		Timeout:    time.Hour,
		MaxRefresh: time.Hour,
		Authorizator: func(userID string, c *gin.Context) bool {
			db := components.App.DB()
			count := 0
			if err := db.Where("id = ?", userID).Model(&models.User{}).Count(&count).Error; err != nil {
				return false
			}

			if count == 0 {
				return false
			}

			return true
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			components.ResponseError(c, code, errors.New(message))
		},
		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	}
}
