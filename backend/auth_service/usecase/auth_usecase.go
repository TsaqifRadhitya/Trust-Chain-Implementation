package usecase

import (
	"context"
	"errors"
	"os"
	"time"
	"backend/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type authUsecase struct {
	userRepo domain.UserRepository
}

func NewAuthUsecase(userRepo domain.UserRepository) domain.AuthUsecase {
	return &authUsecase{userRepo: userRepo}
}

func (u *authUsecase) Login(ctx context.Context, email string, password string) (string, string, *domain.User, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", nil, errors.New("email atau password salah")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", nil, errors.New("email atau password salah")
	}

	// Generate Access Token (expires in 1 hour)
	accessTokenClaims := jwt.MapClaims{
		"userId": user.ID,
		"exp":    time.Now().Add(time.Hour * 1).Unix(),
		"type":   "access",
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "my-super-secret-key-12345"
	}
	accessTokenString, err := accessToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", nil, err
	}

	// Generate Refresh Token (expires in 7 days)
	refreshTokenClaims := jwt.MapClaims{
		"userId": user.ID,
		"exp":    time.Now().Add(time.Hour * 24 * 7).Unix(),
		"type":   "refresh",
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)

	jwtRefreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if jwtRefreshSecret == "" {
		jwtRefreshSecret = "my-super-secret-refresh-key-67890"
	}
	refreshTokenString, err := refreshToken.SignedString([]byte(jwtRefreshSecret))
	if err != nil {
		return "", "", nil, err
	}

	return accessTokenString, refreshTokenString, user, nil
}

func (u *authUsecase) Refresh(ctx context.Context, refreshTokenString string) (string, string, error) {
	jwtRefreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if jwtRefreshSecret == "" {
		jwtRefreshSecret = "my-super-secret-refresh-key-67890"
	}

	token, err := jwt.Parse(refreshTokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtRefreshSecret), nil
	})

	if err != nil || !token.Valid {
		return "", "", errors.New("refresh token tidak valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("refresh token tidak valid")
	}

	// Validate token type
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return "", "", errors.New("token bukan refresh token")
	}

	userIdFloat, ok := claims["userId"].(float64)
	if !ok {
		return "", "", errors.New("refresh token tidak valid")
	}

	userId := uint(userIdFloat)

	// Verify user still exists in database
	user, err := u.userRepo.GetByID(ctx, userId)
	if err != nil {
		return "", "", errors.New("user tidak ditemukan")
	}

	// Generate new access token
	accessTokenClaims := jwt.MapClaims{
		"userId": user.ID,
		"exp":    time.Now().Add(time.Hour * 1).Unix(),
		"type":   "access",
	}
	newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "my-super-secret-key-12345"
	}
	newAccessTokenString, err := newAccessToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", err
	}

	// Generate new refresh token (token rotation)
	refreshTokenClaims := jwt.MapClaims{
		"userId": user.ID,
		"exp":    time.Now().Add(time.Hour * 24 * 7).Unix(),
		"type":   "refresh",
	}
	newRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	newRefreshTokenString, err := newRefreshToken.SignedString([]byte(jwtRefreshSecret))
	if err != nil {
		return "", "", err
	}

	return newAccessTokenString, newRefreshTokenString, nil
}

func (u *authUsecase) GetUserByID(ctx context.Context, id uint) (*domain.User, error) {
	return u.userRepo.GetByID(ctx, id)
}

func (u *authUsecase) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	return u.userRepo.GetAll(ctx)
}
