package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/VariableSan/go-factory-microservice/pkg/common/logger"
	"github.com/VariableSan/go-factory-microservice/pkg/common/response"
	"github.com/VariableSan/go-factory-microservice/pkg/common/tracing"
	authMiddleware "github.com/VariableSan/go-factory-microservice/services/auth/internal/middleware"
	"github.com/VariableSan/go-factory-microservice/services/auth/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type HTTPServer struct {
	server         *http.Server
	authService    *service.AuthService
	logger         *logger.Logger
	jwtSecret      string
	tracingManager *tracing.TracingManager
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Active    bool      `json:"active"`
}

func NewHTTPServer(authService *service.AuthService, port, jwtSecret string, logger *logger.Logger, tracingManager *tracing.TracingManager) *HTTPServer {	
	r := chi.NewRouter()
	
	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	
	// Add tracing middleware if tracing is available
	if tracingManager != nil {
		r.Use(tracingManager.HTTPMiddleware())
	}

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure properly for production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	httpServer := &HTTPServer{
		authService:    authService,
		logger:         logger.WithComponent("http-server"),
		jwtSecret:      jwtSecret,
		tracingManager: tracingManager,
		server: &http.Server{
			Addr:         ":" + port,
			Handler:      r,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	// Routes
	httpServer.setupRoutes(r)

	return httpServer
}

func (s *HTTPServer) setupRoutes(r chi.Router) {
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/login", s.login)
		r.Post("/register", s.register)
		r.Post("/refresh", s.refreshToken)
		
		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.AuthMiddleware(s.jwtSecret))
			r.Get("/profile", s.getProfile)
			r.Get("/validate", s.validateToken)
		})
	})

	// Health check
	r.Get("/health", s.health)
}

func (s *HTTPServer) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	user, token, refreshToken, err := s.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		s.logger.Error("Login failed", "error", err, "email", req.Email)
		response.Unauthorized(w, err.Error())
		return
	}

	response.SuccessWithMessage(w, map[string]interface{}{
		"token":         token,
		"refresh_token": refreshToken,
		"user":          convertToUserResponse(user),
	}, "Login successful")
}

func (s *HTTPServer) register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	user, err := s.authService.Register(r.Context(), req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		s.logger.Error("Registration failed", "error", err, "email", req.Email)
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, map[string]interface{}{
		"user": convertToUserResponse(user),
	})
}

func (s *HTTPServer) refreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	newToken, err := s.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		s.logger.Error("Token refresh failed", "error", err)
		response.Unauthorized(w, err.Error())
		return
	}

	response.SuccessWithMessage(w, map[string]interface{}{
		"token":         newToken,
		"refresh_token": req.RefreshToken, // Keep the same refresh token
	}, "Token refreshed successfully")
}

func (s *HTTPServer) getProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	
	user, err := s.authService.GetProfile(r.Context(), userID)
	if err != nil {
		s.logger.Error("Get profile failed", "error", err, "userID", userID)
		response.Error(w, err)
		return
	}

	response.SuccessWithMessage(w, map[string]interface{}{
		"user": convertToUserResponse(user),
	}, "User profile retrieved successfully")
}

func (s *HTTPServer) validateToken(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		response.Error(w, errors.New("authorization header required"))
		return
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	user, err := s.authService.ValidateToken(r.Context(), token)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.SuccessWithMessage(w, map[string]interface{}{
		"user":  convertToUserResponse(user),
		"roles": []string{}, // No roles in simplified model
	}, "Token is valid")
}

func (s *HTTPServer) health(w http.ResponseWriter, r *http.Request) {
	response.SuccessWithMessage(w, map[string]interface{}{
		"service":   "auth",
		"timestamp": time.Now().Unix(),
	}, "Service is healthy")
}

func (s *HTTPServer) Start() error {
	s.logger.Info("Starting HTTP server", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server")
	return s.server.Shutdown(ctx)
}

func convertToUserResponse(user *service.User) *UserResponse {
	if user == nil {
		return nil
	}
	
	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Roles:     []string{}, // No roles in simplified model
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Active:    user.Active,
	}
}
