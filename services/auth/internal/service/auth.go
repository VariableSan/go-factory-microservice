package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/VariableSan/go-factory-microservice/pkg/common/redis"
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
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Roles     []string  `json:"roles" db:"roles"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Active    bool      `json:"active" db:"active"`
}

type AuthService struct {
	users       map[string]*User // In-memory store for now
	jwtSecret   string
	redisClient *redis.Client
	logger      *slog.Logger
}

type JWTClaims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

func NewAuthService(jwtSecret string, redisClient *redis.Client) *AuthService {
	// Initialize with some test users for development
	users := make(map[string]*User)

	// Create a test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := &User{
		ID:        uuid.New().String(),
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		FirstName: "Test",
		LastName:  "User",
		Roles:     []string{"user"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Active:    true,
	}
	users[testUser.Email] = testUser

	return &AuthService{
		users:       users,
		jwtSecret:   jwtSecret,
		redisClient: redisClient,
		logger:      slog.Default(),
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, firstName, lastName string) (*User, error) {
	// Check if user already exists
	if _, exists := s.users[email]; exists {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &User{
		ID:        uuid.New().String(),
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: firstName,
		LastName:  lastName,
		Roles:     []string{"user"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Active:    true,
	}

	// Store user
	s.users[email] = user

	s.logger.Info("User registered successfully", "email", email, "user_id", user.ID)
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*User, string, string, error) {
	// Find user
	user, exists := s.users[email]
	if !exists {
		return nil, "", "", ErrInvalidCredentials
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	// Check if user is active
	if !user.Active {
		return nil, "", "", errors.New("user account is deactivated")
	}

	// Generate tokens
	token, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in Redis
	refreshKey := fmt.Sprintf("refresh_token:%s", user.ID)
	if s.redisClient != nil {
		s.redisClient.SetWithExpiry(ctx, refreshKey, refreshToken, 7*24*time.Hour)
	}

	s.logger.Info("User logged in successfully", "email", email, "user_id", user.ID)
	return user, token, refreshToken, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*User, []string, error) {
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, nil, ErrInvalidToken
	}

	// Find user
	user, exists := s.users[claims.Email]
	if !exists {
		return nil, nil, ErrUserNotFound
	}

	return user, claims.Roles, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return "", "", ErrInvalidToken
	}

	if !token.Valid {
		return "", "", ErrInvalidToken
	}

	// Find user
	user, exists := s.users[claims.Email]
	if !exists {
		return "", "", ErrUserNotFound
	}

	// Validate refresh token from Redis
	if s.redisClient != nil {
		refreshKey := fmt.Sprintf("refresh_token:%s", user.ID)
		storedToken, err := s.redisClient.GetString(ctx, refreshKey)
		if err != nil || storedToken != refreshToken {
			return "", "", ErrInvalidToken
		}
	}

	// Generate new tokens
	newAccessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update refresh token in Redis
	if s.redisClient != nil {
		refreshKey := fmt.Sprintf("refresh_token:%s", user.ID)
		s.redisClient.SetWithExpiry(ctx, refreshKey, newRefreshToken, 7*24*time.Hour)
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *AuthService) GetUserProfile(ctx context.Context, userID string) (*User, error) {
	// Find user by ID
	for _, user := range s.users {
		if user.ID == userID {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (s *AuthService) generateAccessToken(user *User) (string, error) {
	claims := &JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Roles:  user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
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
		Roles:  user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
