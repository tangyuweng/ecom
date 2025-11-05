package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/tangyuweng/ecom/conf"
	"github.com/tangyuweng/ecom/internal/application/usecase"
	"github.com/tangyuweng/ecom/internal/presentation/http/handler"
	"github.com/tangyuweng/ecom/internal/presentation/http/middleware"
)

func SetupRouter(
	authUseCase *usecase.AuthUseCase,
	categoryUseCase *usecase.CategoryUseCase,
	productUseCase *usecase.ProductUseCase,
	cartUseCase *usecase.CartUseCase,
	jwtService usecase.JWTService,
	cfg *conf.Config,
) *gin.Engine {
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authHandler := handler.NewAuthHandler(authUseCase, cfg)
	categoryHandler := handler.NewCategoryHandler(categoryUseCase)
	productHandler := handler.NewProductHandler(productUseCase)
	cartHandler := handler.NewCartHandler(cartUseCase)

	authMiddleware := middleware.AuthMiddleware(jwtService)

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
		}
		categories := v1.Group("/categories")
		{
			categories.GET("", categoryHandler.GetCategoryList)
			categories.GET("/:id", categoryHandler.GetCategory)
			categories.POST("", authMiddleware, categoryHandler.CreateCategory)
			categories.PUT("/:id", authMiddleware, categoryHandler.UpdateCategory)
			categories.DELETE("/:id", authMiddleware, categoryHandler.DeleteCategory)
		}
		products := v1.Group("/products")
		{
			products.GET("", productHandler.GetProducts)
			products.GET("/:id", productHandler.GetProduct)
			products.GET("/category/:id", productHandler.GetProductsByCategory)
			products.POST("", authMiddleware, productHandler.CreateProduct)
			products.PUT("/:id", authMiddleware, productHandler.UpdateProduct)
			products.DELETE("/:id", authMiddleware, productHandler.DeleteProduct)
			products.PATCH("/:id/stock", authMiddleware, productHandler.UpdateProductStock)
		}
		cart := v1.Group("/cart")
		{
			cart.Use(authMiddleware)
			cart.GET("", cartHandler.GetCart)
			cart.POST("/items", cartHandler.AddToCart)
			cart.PUT("/items/:id", cartHandler.UpdateCartItem)
			cart.DELETE("/items/:id", cartHandler.RemoveFromCart)
			cart.DELETE("", cartHandler.ClearCart)
		}
	}

	return r
}
