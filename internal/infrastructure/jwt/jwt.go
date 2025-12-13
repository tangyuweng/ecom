package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tangyuweng/ecom/internal/domain/repository"
	"github.com/tangyuweng/ecom/internal/domain/service"
)

type JWTSvc struct {
	secretKey          string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
	tokenRepo          repository.TokenRepository
}

type Claims struct {
	UserID  string `json:"user_id"`
	Version int    `json:"version"`
	jwt.RegisteredClaims
}

func NewJWT(secretKey string, accessTokenExpiry, refreshTokenExpiry time.Duration, tokenRepo repository.TokenRepository) service.JWTService {
	return &JWTSvc{
		secretKey:          secretKey,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
		tokenRepo:          tokenRepo,
	}
}

func (s *JWTSvc) GenerateAccessToken(userID string, version int) (string, error) {
	expirationTime := time.Now().Add(s.accessTokenExpiry)
	jti := uuid.New().String()

	claims := &Claims{
		UserID:  userID,
		Version: version,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ecom-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *JWTSvc) GenerateRefreshToken(userID string, version int) (string, error) {
	expirationTime := time.Now().Add(s.refreshTokenExpiry)
	jti := uuid.New().String()

	claims := &Claims{
		UserID:  userID,
		Version: version,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ecom-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *JWTSvc) ValidateTokenBasic(tokenString string) (*service.TokenDetails, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return &service.TokenDetails{
		UserID:    claims.UserID,
		JTI:       claims.ID,
		Version:   claims.Version,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

func (s *JWTSvc) ValidateToken(ctx context.Context, tokenString string) (*service.TokenDetails, error) {
	tokenDetails, err := s.ValidateTokenBasic(tokenString)
	if err != nil {
		return nil, err
	}

	isBlacklisted, err := s.tokenRepo.IsBlacklisted(ctx, tokenDetails.JTI)
	if err != nil {
		return nil, errors.New("failed to verify token")
	}
	if isBlacklisted {
		return nil, errors.New("token has been revoked")
	}

	currentVersion, err := s.tokenRepo.GetUserVersion(ctx, tokenDetails.UserID)
	if err != nil {
		return nil, errors.New("failed to verify token")
	}
	if tokenDetails.Version != currentVersion {
		return nil, errors.New("token has been invalidated")
	}

	return tokenDetails, nil
}
