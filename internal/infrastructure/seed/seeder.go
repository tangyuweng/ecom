package seed

import (
	"context"
	"log"

	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/domain/repository"
)

type Seeder struct {
	userRepo     repository.UserRepository
	categoryRepo repository.CategoryRepository
	productRepo  repository.ProductRepository
}

func NewSeeder(
	userRepo repository.UserRepository,
	categoryRepo repository.CategoryRepository,
	productRepo repository.ProductRepository,
) *Seeder {
	return &Seeder{
		userRepo:     userRepo,
		categoryRepo: categoryRepo,
		productRepo:  productRepo,
	}
}

func (s *Seeder) SeedAll(ctx context.Context) error {
	if err := s.seedUsers(ctx); err != nil {
		return err
	}

	if err := s.seedCategories(ctx); err != nil {
		return err
	}

	if err := s.seedProducts(ctx); err != nil {
		return err
	}

	log.Println("Database seeding completed successfully")
	return nil
}

func (s *Seeder) seedUsers(ctx context.Context) error {
	exists, _ := s.userRepo.ExistsByEmail(ctx, "admin@example.com")
	if exists {
		log.Println("Admin user already exists, skipping...")
		return nil
	}

	admin, err := entity.NewAdmin(
		"admin@example.com",
		"admin123",
		"Admin User",
		"0912345678",
	)
	if err != nil {
		return err
	}

	if err := s.userRepo.Create(ctx, admin); err != nil {
		return err
	}

	user, err := entity.NewUser(
		"user@example.com",
		"user1234",
		"Test User",
		"0987654321",
	)
	if err != nil {
		return err
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}

	log.Println("Users seeded successfully")
	return nil
}

func (s *Seeder) seedCategories(ctx context.Context) error {
	categories := []struct {
		name string
		slug string
	}{
		{"電子產品", "electronics"},
		{"服飾配件", "fashion"},
		{"家居用品", "home"},
		{"運動健身", "sports"},
		{"書籍文具", "books"},
	}

	for _, cat := range categories {
		exists, _ := s.categoryRepo.ExistsByName(ctx, cat.name)
		if exists {
			log.Printf("Category %s already exists, skipping...", cat.name)
			continue
		}

		category, err := entity.NewCategory(cat.name, cat.slug)
		if err != nil {
			return err
		}

		if err := s.categoryRepo.Create(ctx, category); err != nil {
			return err
		}
	}

	log.Println("Categories seeded successfully")
	return nil
}

func (s *Seeder) seedProducts(ctx context.Context) error {
	categories, err := s.categoryRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	if len(categories) == 0 {
		log.Println("No categories found, skipping products...")
		return nil
	}

	products := []struct {
		name        string
		description string
		price       float64
		stock       int
		categoryIdx int
		IsActive    bool
	}{
		{"iPhone 15 Pro", "最新款 iPhone，搭載 A17 Pro 晶片", 35900, 50, 0, true},
		{"MacBook Air M2", "輕薄筆記型電腦，效能卓越", 36900, 30, 0, true},
		{"AirPods Pro", "主動降噪無線耳機", 7990, 100, 0, true},
		{"運動T恤", "透氣排汗，適合運動穿著", 890, 200, 1, true},
		{"牛仔褲", "經典款牛仔褲，舒適耐穿", 1590, 150, 1, true},
		{"咖啡機", "全自動義式咖啡機", 12900, 20, 2, false},
		{"掃地機器人", "智能清潔，自動回充", 8900, 40, 2, true},
		{"瑜珈墊", "加厚防滑瑜珈墊", 690, 80, 3, false},
		{"啞鈴組", "可調式啞鈴 5-25kg", 3500, 60, 3, false},
		{"Python 程式設計", "Python 入門到精通", 580, 100, 4, true},
	}

	for _, p := range products {
		if p.categoryIdx >= len(categories) {
			continue
		}

		product, err := entity.NewProduct(
			categories[p.categoryIdx].ID,
			p.name,
			p.description,
			p.price,
			p.stock,
			p.IsActive,
		)
		if err != nil {
			log.Printf("Failed to create product %s: %v", p.name, err)
			continue
		}

		if err := s.productRepo.Create(ctx, product); err != nil {
			log.Printf("Failed to seed product %s: %v", p.name, err)
			continue
		}
	}

	log.Println("Products seeded successfully")
	return nil
}
