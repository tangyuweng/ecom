package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/tangyuweng/ecom/conf"
	_ "github.com/tangyuweng/ecom/docs"
	"github.com/tangyuweng/ecom/internal/application/usecase"
	"github.com/tangyuweng/ecom/internal/infrastructure/jwt"
	"github.com/tangyuweng/ecom/internal/infrastructure/mysql"
	"github.com/tangyuweng/ecom/internal/infrastructure/seed"
	"github.com/tangyuweng/ecom/internal/infrastructure/websocket"
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
	shouldSeed := flag.Bool("seed", false, "Run database seeding")
	shouldMigrateUp := flag.Bool("migrate-up", false, "Run migrations up")
	shouldMigrateDown := flag.Bool("migrate-down", false, "Rollback last migration")
	migrateSteps := flag.Int("migrate-steps", 0, "Run specific number of migration steps (positive=up, negative=down)")
	shouldCheckMigrationVersion := flag.Bool("migrate-status", false, "Check migration version")
	migrateForce := flag.Int("migrate-force", -1, "Force migration to specific version")
	flag.Parse()

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

	migrator, err := mysql.NewMigrator(db, "internal/infrastructure/mysql/migrations")
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}

	if *shouldMigrateUp {
		log.Println("Running migrations up...")
		if err := migrator.Up(); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Migrations completed successfully")
		return
	}

	if *shouldMigrateDown {
		log.Println("Rolling back migration...")
		if err := migrator.Down(); err != nil {
			log.Fatalf("Failed to rollback migration: %v", err)
		}
		log.Println("Rollback completed successfully")
		return
	}

	if *migrateSteps != 0 {
		log.Printf("Running %d migration steps...", *migrateSteps)
		if err := migrator.Steps(*migrateSteps); err != nil {
			log.Fatalf("Failed to run migration steps: %v", err)
		}
		log.Println("Migration steps completed successfully")
		return
	}

	if *migrateForce >= 0 {
		log.Printf("Forcing migration to version %d...", *migrateForce)
		if err := migrator.Force(*migrateForce); err != nil {
			log.Fatalf("Failed to force migration: %v", err)
		}
		log.Println("Migration forced successfully")
		return
	}

	if *shouldCheckMigrationVersion {
		version, dirty, err := migrator.Version()
		if err != nil {
			log.Printf("Failed to get migration version: %v", err)
			return
		}

		if dirty {
			log.Printf("Migration status: DIRTY (version %d)", version)
			log.Println("The database is in an inconsistent state.")
			log.Println("Run 'make migrate-force' to fix it.")
		} else {
			log.Printf("Current migration version: %d", version)
			log.Println("Database is up to date.")
		}
		return
	}

	userRepo := mysql.NewUserRepository(db)
	categoryRepo := mysql.NewMysqlCategoryRepository(db)
	productRepo := mysql.NewMysqlProductRepository(db)
	cartRepo := mysql.NewMysqlCartRepository(db)
	orderRepo := mysql.NewMysqlOrderRepository(db)

	if *shouldSeed {
		seeder := seed.NewSeeder(userRepo, categoryRepo, productRepo)
		if err := seeder.SeedAll(context.Background()); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
		log.Println("Seeding completed, exiting...")
		return
	}

	// 創建 TransactionManager
	txManager := mysql.NewGormTransactionManager(db)

	wsHub := websocket.NewHub()
	go wsHub.Run()

	notificationSvc := websocket.NewWsNotificationSvc(wsHub)

	jwtService := jwt.NewJWT(
		cfg.JWT.Secret,
		time.Duration(cfg.JWT.AccessTokenExpiry)*time.Hour,
		time.Duration(cfg.JWT.RefreshTokenExpiry)*time.Hour,
	)

	authUseCase := usecase.NewAuthUseCase(userRepo, jwtService)
	userUseCase := usecase.NewUserUseCase(userRepo)
	categoryUseCase := usecase.NewCategoryUseCase(categoryRepo, userRepo, productRepo)
	productUseCase := usecase.NewProductUseCase(productRepo, categoryRepo, userRepo)
	cartUseCase := usecase.NewCartUseCase(cartRepo, productRepo)
	orderUseCase := usecase.NewOrderUseCase(orderRepo, cartRepo, productRepo, userRepo, notificationSvc, txManager)

	r := router.SetupRouter(authUseCase, userUseCase, categoryUseCase, productUseCase, cartUseCase, orderUseCase, jwtService, cfg, wsHub)

	log.Printf("Starting server on %s", cfg.Server.Port)
	log.Printf("Swagger UI: http://localhost%s/swagger/index.html", cfg.Server.Port)

	if err := r.Run(cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
