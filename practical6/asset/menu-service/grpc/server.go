package grpc

import (
	"context"
	"time"

	menuv1 "github.com/douglasswm/student-cafe-protos/gen/go/menu/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"menu-service/database"
	"menu-service/models"
)

// MenuServer implements the gRPC MenuService
type MenuServer struct {
	menuv1.UnimplementedMenuServiceServer
}

// NewMenuServer creates a new gRPC menu server
func NewMenuServer() *MenuServer {
	return &MenuServer{}
}

// GetMenuItem retrieves a menu item by ID
func (s *MenuServer) GetMenuItem(ctx context.Context, req *menuv1.GetMenuItemRequest) (*menuv1.GetMenuItemResponse, error) {
	var menuItem models.MenuItem
	if err := database.DB.First(&menuItem, req.Id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "menu item not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get menu item: %v", err)
	}

	return &menuv1.GetMenuItemResponse{
		MenuItem: modelToProto(&menuItem),
	}, nil
}

// GetMenu retrieves all menu items
func (s *MenuServer) GetMenu(ctx context.Context, req *menuv1.GetMenuRequest) (*menuv1.GetMenuResponse, error) {
	var menuItems []models.MenuItem
	if err := database.DB.Find(&menuItems).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get menu: %v", err)
	}

	protoItems := make([]*menuv1.MenuItem, len(menuItems))
	for i, item := range menuItems {
		protoItems[i] = modelToProto(&item)
	}

	return &menuv1.GetMenuResponse{
		MenuItems: protoItems,
	}, nil
}

// CreateMenuItem creates a new menu item
func (s *MenuServer) CreateMenuItem(ctx context.Context, req *menuv1.CreateMenuItemRequest) (*menuv1.CreateMenuItemResponse, error) {
	menuItem := models.MenuItem{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
	}

	if err := database.DB.Create(&menuItem).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create menu item: %v", err)
	}

	return &menuv1.CreateMenuItemResponse{
		MenuItem: modelToProto(&menuItem),
	}, nil
}

// modelToProto converts a GORM MenuItem model to proto MenuItem message
func modelToProto(item *models.MenuItem) *menuv1.MenuItem {
	return &menuv1.MenuItem{
		Id:          uint32(item.ID),
		Name:        item.Name,
		Description: item.Description,
		Price:       item.Price,
		CreatedAt:   item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.Format(time.RFC3339),
	}
}
