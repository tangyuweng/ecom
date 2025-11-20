package entity

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID        string
	Email     string
	Password  string
	Name      string
	Phone     string
	Role      UserRole
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters")
	ErrInvalidName        = errors.New("name cannot be empty")
	ErrInvalidPhone       = errors.New("phone cannot be empty")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserUnauthorized   = errors.New("unauthorized access to or modification of the User is forbidden")
	ErrInvalidUserRole    = errors.New("invalid role")
	ErrInvalidCredentials = errors.New("invalid email or password")
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
		Role:      RoleUser,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func NewAdmin(email, password, name, phone string) (*User, error) {
	user, err := NewUser(email, password, name, phone)
	if err != nil {
		return nil, err
	}
	user.Role = RoleAdmin
	return user, nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) UpdateUser(name, phone string) error {
	if err := validateName(name); err != nil {
		return err
	}

	if err := validatePhone(phone); err != nil {
		return err
	}

	u.Name = name
	u.Phone = phone
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) UpdateRole(role UserRole) error {
	if role != RoleUser && role != RoleAdmin {
		return ErrInvalidUserRole
	}
	u.Role = role
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) UpdatePassword(newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	hashPwd, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	u.Password = hashPwd
	u.UpdatedAt = time.Now()
	return nil
}

func validateEmail(email string) error {
	if email == "" || !strings.Contains(email, "@") {
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
