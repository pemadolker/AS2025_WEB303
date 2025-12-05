package grpc

import (
	"context"
	"time"

	userv1 "github.com/douglasswm/student-cafe-protos/gen/go/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"user-service/database"
	"user-service/models"
)

// UserServer implements the gRPC UserService
type UserServer struct {
	userv1.UnimplementedUserServiceServer
}

// NewUserServer creates a new gRPC user server
func NewUserServer() *UserServer {
	return &UserServer{}
}

// CreateUser creates a new user
func (s *UserServer) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	user := models.User{
		Name:        req.Name,
		Email:       req.Email,
		IsCafeOwner: req.IsCafeOwner,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &userv1.CreateUserResponse{
		User: modelToProto(&user),
	}, nil
}

// GetUser retrieves a user by ID
func (s *UserServer) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	var user models.User
	if err := database.DB.First(&user, req.Id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return &userv1.GetUserResponse{
		User: modelToProto(&user),
	}, nil
}

// GetUsers retrieves all users
func (s *UserServer) GetUsers(ctx context.Context, req *userv1.GetUsersRequest) (*userv1.GetUsersResponse, error) {
	var users []models.User
	if err := database.DB.Find(&users).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get users: %v", err)
	}

	protoUsers := make([]*userv1.User, len(users))
	for i, user := range users {
		protoUsers[i] = modelToProto(&user)
	}

	return &userv1.GetUsersResponse{
		Users: protoUsers,
	}, nil
}

// modelToProto converts a GORM User model to proto User message
func modelToProto(user *models.User) *userv1.User {
	return &userv1.User{
		Id:          uint32(user.ID),
		Name:        user.Name,
		Email:       user.Email,
		IsCafeOwner: user.IsCafeOwner,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
	}
}
