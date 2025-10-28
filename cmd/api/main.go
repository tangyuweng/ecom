package main

import (
	"log"
	"time"

	"github.com/tangyuweng/ecom/conf"
	_ "github.com/tangyuweng/ecom/docs"
	"github.com/tangyuweng/ecom/internal/application/usecase"
	"github.com/tangyuweng/ecom/internal/infrastructure/jwt"
	"github.com/tangyuweng/ecom/internal/infrastructure/mysql"
	"github.com/tangyuweng/ecom/internal/presentation/http/router"
)

// @title           ECOM API
// @version         1.0
// @description     A e-commerce API server
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:3000
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @scheme bearer
// @bearerFormat JWT
func main() {
	cfg, err := conf.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := mysql.NewDatabase(mysql.DatabaseConfig{
		Username: cfg.MySQL.Username,
		Password: cfg.MySQL.Password,
		Host:     cfg.MySQL.Host,
		Port:     cfg.MySQL.Port,
		DBName:   cfg.MySQL.DBName,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := mysql.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	userRepo := mysql.NewUserRepository(db)

	jwtService := jwt.NewJWTService(cfg.JWT.Secret, time.Duration(cfg.JWT.Duration))

	authUseCase := usecase.NewAuthUseCase(userRepo, jwtService)

	r := router.SetupRouter(authUseCase, jwtService)

	log.Printf("Starting server on %s", cfg.Server.Port)
	log.Printf("Swagger UI: http://localhost%s/swagger/index.html", cfg.Server.Port)

	if err := r.Run(cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
