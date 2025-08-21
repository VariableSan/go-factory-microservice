package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/VariableSan/go-factory-microservice/pkg/common/database"
	"github.com/VariableSan/go-factory-microservice/pkg/common/redis"
	"github.com/VariableSan/go-factory-microservice/services/auth/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid token")
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Active    bool      `json:"active"`
}

type AuthService struct {
	userRepo      *repository.UserRepository
	jwtSecret     string
	redisClient   *redis.Client
	logger        *slog.Logger
	tokenExpiry   time.Duration
	refreshExpiry time.Duration
}

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func NewAuthService(db *database.DB, jwtSecret string, redisClient *redis.Client, tokenExpiry, refreshExpiry time.Duration) *AuthService {
	userRepo := repository.NewUserRepository(db)
	
	service := &AuthService{
		userRepo:      userRepo,
		jwtSecret:     jwtSecret,
		redisClient:   redisClient,
		logger:        slog.Default(),
		tokenExpiry:   tokenExpiry,
		refreshExpiry: refreshExpiry,
	}

	// Initialize database tables
	ctx := context.Background()
	if err := userRepo.CreateTable(ctx); err != nil {
		service.logger.Error("Failed to create user table", "error", err)
	}

	return service
}

func (s *AuthService) Register(ctx context.Context, email, password, firstName, lastName string) (*User, error) {
	// Check if user already exists
	exists, err := s.userRepo.EmailExists(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	repoUser := &repository.User{
		ID:        uuid.New().String(),
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: firstName,
		LastName:  lastName,
		Active:    true,
	}

	// Store user in database
	if err := s.userRepo.Create(ctx, repoUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Convert to service user (without password)
	user := &User{
		ID:        repoUser.ID,
		Email:     repoUser.Email,
		FirstName: repoUser.FirstName,
		LastName:  repoUser.LastName,
		CreatedAt: repoUser.CreatedAt,
		UpdatedAt: repoUser.UpdatedAt,
		Active:    repoUser.Active,
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*User, string, string, error) {
	// Get user from database
	repoUser, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(repoUser.Password), []byte(password)); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	// Convert to service user
	user := &User{
		ID:        repoUser.ID,
		Email:     repoUser.Email,
		FirstName: repoUser.FirstName,
		LastName:  repoUser.LastName,
		CreatedAt: repoUser.CreatedAt,
		UpdatedAt: repoUser.UpdatedAt,
		Active:    repoUser.Active,
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in Redis
	if s.redisClient != nil {
		refreshKey := fmt.Sprintf("refresh_token:%s", user.ID)
		if err := s.redisClient.SetWithExpiry(ctx, refreshKey, refreshToken, s.refreshExpiry); err != nil {
			s.logger.Warn("Failed to store refresh token in Redis", "error", err)
		}
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID string) (*User, error) {
	repoUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	user := &User{
		ID:        repoUser.ID,
		Email:     repoUser.Email,
		FirstName: repoUser.FirstName,
		LastName:  repoUser.LastName,
		CreatedAt: repoUser.CreatedAt,
		UpdatedAt: repoUser.UpdatedAt,
		Active:    repoUser.Active,
	}

	return user, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*User, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Get user from database to ensure they still exist and are active
	user, err := s.GetProfile(ctx, claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// Parse refresh token
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return "", ErrInvalidToken
	}

	// Check if refresh token exists in Redis
	if s.redisClient != nil {
		refreshKey := fmt.Sprintf("refresh_token:%s", claims.UserID)
		storedToken, err := s.redisClient.GetString(ctx, refreshKey)
		if err != nil || storedToken != refreshToken {
			return "", ErrInvalidToken
		}
	}

	// Get user from database
	user, err := s.GetProfile(ctx, claims.UserID)
	if err != nil {
		return "", ErrUserNotFound
	}

	// Generate new access token
	newAccessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	return newAccessToken, nil
}

func (s *AuthService) Health() error {
	// Check database connection
	if err := s.userRepo.DB.Health(); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	
	// Check Redis connection if available
	if s.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.redisClient.Ping(ctx); err != nil {
			s.logger.Warn("Redis health check failed", "error", err)
		}
	}
	
	return nil
}

func (s *AuthService) generateAccessToken(user *User) (string, error) {
	claims := &JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) generateRefreshToken(user *User) (string, error) {
	claims := &JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) parseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
