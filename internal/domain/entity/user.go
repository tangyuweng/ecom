package entity

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        string
	Email     string
	Password  string
	Name      string
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidPassword = errors.New("password must be at least 8 characters")
	ErrInvalidName     = errors.New("name cannot be empty")
	ErrInvalidPhone    = errors.New("phone cannot be empty")
)

func NewUser(email, password, name, phone string) (*User, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	if err := validatePassword(password); err != nil {
		return nil, err
	}

	if err := validateName(name); err != nil {
		return nil, err
	}

	if err := validatePhone(phone); err != nil {
		return nil, err
	}

	hashPwd, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return &User{
		Email:     email,
		Password:  hashPwd,
		Name:      name,
		Phone:     phone,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func validateEmail(email string) error {
	if email == "" || !strings.Contains(email, "@") || !strings.Contains(email, " ") {
		return ErrInvalidEmail
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrInvalidPassword
	}
	return nil
}

func validateName(name string) error {
	if name == "" {
		return ErrInvalidName
	}
	return nil
}

func validatePhone(phone string) error {
	if phone == "" {
		return ErrInvalidPhone
	}
	return nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
