package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tangyuweng/ecom/internal/application/dto"
	"github.com/tangyuweng/ecom/internal/domain/service"
)

// WebSocketAuthMiddleware 專門用於 WebSocket 的認證中間件
// 支援從 query string 中讀取 token
func WebSocketAuthMiddleware(jwtService service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Query("token")

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse("token required in query string"))
			c.Abort()
			return
		}

		tokenDetails, err := jwtService.ValidateTokenBasic(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse(err.Error()))
			c.Abort()
			return
		}

		c.Set("userID", tokenDetails.UserID)
		c.Next()
	}
}
