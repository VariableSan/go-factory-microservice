package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/VariableSan/go-factory-microservice/pkg/common/database"
	"github.com/google/uuid"
)

type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Active    bool      `json:"active" db:"active"`
}

type UserRepository struct {
	DB *database.DB // Expose for health checks
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (id, email, password, first_name, last_name, created_at, updated_at, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.DB.ExecContext(ctx, query,
		user.ID, user.Email, user.Password, user.FirstName, user.LastName,
		user.CreatedAt, user.UpdatedAt, user.Active,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, email, password, first_name, last_name, created_at, updated_at, active
		FROM users 
		WHERE email = $1 AND active = true
	`

	err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.CreatedAt, &user.UpdatedAt, &user.Active,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, email, password, first_name, last_name, created_at, updated_at, active
		FROM users 
		WHERE id = $1 AND active = true
	`

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.CreatedAt, &user.UpdatedAt, &user.Active,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users 
		SET first_name = $2, last_name = $3, updated_at = $4
		WHERE id = $1
	`

	result, err := r.DB.ExecContext(ctx, query, user.ID, user.FirstName, user.LastName, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Delete soft deletes a user (sets active = false)
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE users SET active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// EmailExists checks if email already exists
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	err := r.DB.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}
