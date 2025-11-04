package entity

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type Category struct {
	ID        string
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	ErrInvalidCategoryName  = errors.New("category name cannot be empty")
	ErrInvalidCategorySlug  = errors.New("category slug cannot be empty")
	ErrCategorySlugFormat   = errors.New("slug must contain only lowercase letters, numbers, and hyphens")
	ErrCategoryNotFound     = errors.New("category not found")
	ErrCategoryExists       = errors.New("category with this name or slug already exists")
	ErrCategoryUnauthorized = errors.New("unauthorized access to or modification of the Category is forbidden")
	ErrCategoryHasProducts  = errors.New("cannot delete category with existing products")
)

var slugRegex = regexp.MustCompile("^[a-z0-9-]+$")

func NewCategory(name, slug string) (*Category, error) {
	if err := validateCategoryName(name); err != nil {
		return nil, err
	}

	if err := validateCategorySlug(slug); err != nil {
		return nil, err
	}

	now := time.Now()

	return &Category{
		Name:      strings.ToLower(name),
		Slug:      slug,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (c *Category) Update(name, slug string) error {
	if err := validateCategoryName(name); err != nil {
		return err
	}

	if err := validateCategorySlug(slug); err != nil {
		return err
	}

	c.Name = strings.ToLower(name)
	c.Slug = slug
	c.UpdatedAt = time.Now()
	return nil
}

func validateCategoryName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidCategoryName
	}
	return nil
}

func validateCategorySlug(slug string) error {
	trimmedSlug := strings.TrimSpace(slug)

	if trimmedSlug == "" {
		return ErrInvalidCategorySlug
	}

	if !slugRegex.MatchString(trimmedSlug) {
		return ErrCategorySlugFormat
	}

	return nil
}
