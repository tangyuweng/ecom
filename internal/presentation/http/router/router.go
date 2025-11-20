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
	userUseCase *usecase.UserUseCase,
	categoryUseCase *usecase.CategoryUseCase,
	productUseCase *usecase.ProductUseCase,
	cartUseCase *usecase.CartUseCase,
	orderUseCase *usecase.OrderUseCase,
	jwtService usecase.JWTService,
	cfg *conf.Config,
) *gin.Engine {
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authHandler := handler.NewAuthHandler(authUseCase, cfg)
	userHandler := handler.NewUserHandler(userUseCase)
	categoryHandler := handler.NewCategoryHandler(categoryUseCase)
	productHandler := handler.NewProductHandler(productUseCase)
	cartHandler := handler.NewCartHandler(cartUseCase)
	orderHandler := handler.NewOrderHandler(orderUseCase)

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
		users := v1.Group("/users")
		{
			users.Use(authMiddleware)
			users.GET("/me", userHandler.GetUser)
			users.PUT("/me", userHandler.UpdateUser)
			users.PUT("/me/password", userHandler.UpdateUserPassword)
		}
		categories := v1.Group("/categories")
		{
			categories.GET("", categoryHandler.GetCategories)
			categories.GET("/:id", categoryHandler.GetCategory)
		}
		products := v1.Group("/products")
		{
			products.GET("", productHandler.GetProducts)
			products.GET("/:id", productHandler.GetProduct)
			products.GET("/category/:id", productHandler.GetProductsByCategory)
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
		orders := v1.Group("/orders")
		{
			orders.Use(authMiddleware)
			orders.POST("", orderHandler.CreateOrderFromCart)
			orders.GET("", orderHandler.GetUserOrders)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.POST("/:id/cancel", orderHandler.CancelOrder)
		}
		admin := v1.Group("/admin")
		{
			admin.Use(authMiddleware)
			admin.GET("/users", userHandler.GetUsers)
			admin.PATCH("/users/:id/role", userHandler.UpdateUserRole)
			admin.POST("/categories", categoryHandler.CreateCategory)
			admin.PUT("/categories/:id", categoryHandler.UpdateCategory)
			admin.DELETE("/categories/:id", categoryHandler.DeleteCategory)
			admin.POST("/products", productHandler.CreateProduct)
			admin.PUT("/products/:id", productHandler.UpdateProduct)
			admin.DELETE("/products/:id", productHandler.DeleteProduct)
			admin.PATCH("/products/:id/stock", productHandler.UpdateProductStock)
			admin.PATCH("/orders/:id/status", orderHandler.UpdateOrderStatus)
		}
	}

	return r
}
