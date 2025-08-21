package server

import (
	"context"
	"log/slog"
	"net"

	authpb "github.com/VariableSan/go-factory-microservice/pkg/proto/auth"
	"github.com/VariableSan/go-factory-microservice/services/auth/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	server      *grpc.Server
	listener    net.Listener
	authService *service.AuthService
	logger      *slog.Logger
}

func NewGRPCServer(authService *service.AuthService, port string) (*GRPCServer, error) {
	logger := slog.Default()
	
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}

	// Create gRPC server with middleware
	server := grpc.NewServer()

	// Register auth service
	authServer := &AuthGRPCServer{
		authService: authService,
		logger:      logger,
	}
	authpb.RegisterAuthServiceServer(server, authServer)

	// Register health check service
	healthServer := health.NewServer()
	healthServer.SetServingStatus("auth.v1.AuthService", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	// Register reflection service for development
	reflection.Register(server)

	return &GRPCServer{
		server:      server,
		listener:    lis,
		authService: authService,
		logger:      logger,
	}, nil
}

func (s *GRPCServer) Start() error {
	s.logger.Info("Starting gRPC server", "addr", s.listener.Addr())
	return s.server.Serve(s.listener)
}

func (s *GRPCServer) Stop() {
	s.logger.Info("Stopping gRPC server")
	s.server.GracefulStop()
}

// AuthGRPCServer implements the gRPC auth service
type AuthGRPCServer struct {
	authpb.UnimplementedAuthServiceServer
	authService *service.AuthService
	logger      *slog.Logger
}

func (s *AuthGRPCServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	s.logger.Info("Login request", "email", req.Email)
	
	user, token, refreshToken, err := s.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		s.logger.Error("Login failed", "error", err)
		return &authpb.LoginResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &authpb.LoginResponse{
		Success:      true,
		Token:        token,
		RefreshToken: refreshToken,
		User:         convertToProtoUser(user),
		Message:      "Login successful",
	}, nil
}

func (s *AuthGRPCServer) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	s.logger.Info("Register request", "email", req.Email)
	
	user, err := s.authService.Register(ctx, req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		s.logger.Error("Registration failed", "error", err)
		return &authpb.RegisterResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &authpb.RegisterResponse{
		Success: true,
		User:    convertToProtoUser(user),
		Message: "User registered successfully",
	}, nil
}

func (s *AuthGRPCServer) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	user, roles, err := s.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		return &authpb.ValidateTokenResponse{
			Valid:   false,
			Message: err.Error(),
		}, nil
	}

	return &authpb.ValidateTokenResponse{
		Valid:   true,
		User:    convertToProtoUser(user),
		Roles:   roles,
		Message: "Token is valid",
	}, nil
}

func (s *AuthGRPCServer) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
	newToken, newRefreshToken, err := s.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return &authpb.RefreshTokenResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &authpb.RefreshTokenResponse{
		Success:      true,
		Token:        newToken,
		RefreshToken: newRefreshToken,
		Message:      "Token refreshed successfully",
	}, nil
}

func (s *AuthGRPCServer) GetUserProfile(ctx context.Context, req *authpb.GetUserProfileRequest) (*authpb.GetUserProfileResponse, error) {
	user, err := s.authService.GetUserProfile(ctx, req.UserId)
	if err != nil {
		return &authpb.GetUserProfileResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &authpb.GetUserProfileResponse{
		Success: true,
		User:    convertToProtoUser(user),
		Message: "User profile retrieved successfully",
	}, nil
}

func convertToProtoUser(user *service.User) *authpb.User {
	if user == nil {
		return nil
	}
	
	return &authpb.User{
		Id:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Roles:     user.Roles,
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
		Active:    user.Active,
	}
}
