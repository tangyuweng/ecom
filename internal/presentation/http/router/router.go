package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/tangyuweng/ecom/conf"
	"github.com/tangyuweng/ecom/internal/application/usecase"
	"github.com/tangyuweng/ecom/internal/presentation/http/handler"
)

func SetupRouter(
	authUseCase *usecase.AuthUseCase,
	jwtService usecase.JWTService,
	cfg *conf.Config,
) *gin.Engine {
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authHandler := handler.NewAuthHandler(authUseCase, cfg)

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
		}
	}

	return r
}
