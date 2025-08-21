package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/VariableSan/go-factory-microservice/pkg/common/response"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware validates JWT tokens and adds user context
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				response.Unauthorized(w, "Authorization header required")
				return
			}

			// Remove "Bearer " prefix if present
			if len(token) > 7 && strings.ToLower(token[:7]) == "bearer " {
				token = token[7:]
			}

			claims := &JWTClaims{}
			parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})

			if err != nil || !parsedToken.Valid {
				response.Unauthorized(w, "Invalid token")
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), "userID", claims.UserID)
			ctx = context.WithValue(ctx, "email", claims.Email)
			ctx = context.WithValue(ctx, "roles", claims.Roles)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type JWTClaims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}
