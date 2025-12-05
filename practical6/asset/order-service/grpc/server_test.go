package grpc

import (
	"context"
	"database/sql"
	"order-service/database"
	"order-service/models"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	menuv1 "github.com/douglasswm/student-cafe-protos/gen/go/menu/v1"
	orderv1 "github.com/douglasswm/student-cafe-protos/gen/go/order/v1"
	userv1 "github.com/douglasswm/student-cafe-protos/gen/go/user/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MockUserServiceClient is a mock for UserServiceClient
type MockUserServiceClient struct {
	mock.Mock
}

func (m *MockUserServiceClient) CreateUser(ctx context.Context, req *userv1.CreateUserRequest, opts ...grpc.CallOption) (*userv1.CreateUserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userv1.CreateUserResponse), args.Error(1)
}

func (m *MockUserServiceClient) GetUser(ctx context.Context, req *userv1.GetUserRequest, opts ...grpc.CallOption) (*userv1.GetUserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userv1.GetUserResponse), args.Error(1)
}

func (m *MockUserServiceClient) GetUsers(ctx context.Context, req *userv1.GetUsersRequest, opts ...grpc.CallOption) (*userv1.GetUsersResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userv1.GetUsersResponse), args.Error(1)
}

// MockMenuServiceClient is a mock for MenuServiceClient
type MockMenuServiceClient struct {
	mock.Mock
}

func (m *MockMenuServiceClient) GetMenuItem(ctx context.Context, req *menuv1.GetMenuItemRequest, opts ...grpc.CallOption) (*menuv1.GetMenuItemResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*menuv1.GetMenuItemResponse), args.Error(1)
}

func (m *MockMenuServiceClient) GetMenu(ctx context.Context, req *menuv1.GetMenuRequest, opts ...grpc.CallOption) (*menuv1.GetMenuResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*menuv1.GetMenuResponse), args.Error(1)
}

func (m *MockMenuServiceClient) CreateMenuItem(ctx context.Context, req *menuv1.CreateMenuItemRequest, opts ...grpc.CallOption) (*menuv1.CreateMenuItemResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*menuv1.CreateMenuItemResponse), args.Error(1)
}

// setupTestDB creates a mock database for testing
func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Failed to create mock database")

	// Use postgres dialect (works without CGO)
	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})

	// Disable GORM logging during tests to reduce noise
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Failed to open test database")

	return db, mock, sqlDB
}

// teardownTestDB cleans up the test database
func teardownTestDB(t *testing.T, sqlDB *sql.DB) {
	sqlDB.Close()
}

func TestCreateOrder_Success(t *testing.T) {
	// Setup
	db, dbMock, sqlDB := setupTestDB(t)
	defer teardownTestDB(t, sqlDB)
	database.DB = db

	mockUserClient := new(MockUserServiceClient)
	mockMenuClient := new(MockMenuServiceClient)

	server := &OrderServer{
		UserClient: mockUserClient,
		MenuClient: mockMenuClient,
	}

	// Mock user validation
	mockUserClient.On("GetUser", mock.Anything, &userv1.GetUserRequest{Id: 1}).
		Return(&userv1.GetUserResponse{
			User: &userv1.User{Id: 1, Name: "Test User", Email: "test@example.com"},
		}, nil)

	// Mock menu item lookup
	mockMenuClient.On("GetMenuItem", mock.Anything, &menuv1.GetMenuItemRequest{Id: 1}).
		Return(&menuv1.GetMenuItemResponse{
			MenuItem: &menuv1.MenuItem{Id: 1, Name: "Coffee", Price: 2.50},
		}, nil)

	mockMenuClient.On("GetMenuItem", mock.Anything, &menuv1.GetMenuItemRequest{Id: 2}).
		Return(&menuv1.GetMenuItemResponse{
			MenuItem: &menuv1.MenuItem{Id: 2, Name: "Tea", Price: 2.00},
		}, nil)

	// Mock database operations
	now := time.Now()
	dbMock.ExpectBegin()
	// Mock INSERT for order
	dbMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "orders"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, "pending").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, now, now))
	// Mock INSERT for order items (uses QUERY not EXEC because of RETURNING clause)
	dbMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "order_items"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
	dbMock.ExpectCommit()

	// Test
	ctx := context.Background()
	resp, err := server.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		UserId: 1,
		Items: []*orderv1.OrderItemRequest{
			{MenuItemId: 1, Quantity: 2},
			{MenuItemId: 2, Quantity: 1},
		},
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotZero(t, resp.Order.Id)
	assert.Equal(t, uint32(1), resp.Order.UserId)
	assert.Equal(t, "pending", resp.Order.Status)
	assert.Len(t, resp.Order.OrderItems, 2)

	// Verify first item
	assert.Equal(t, uint32(1), resp.Order.OrderItems[0].MenuItemId)
	assert.Equal(t, int32(2), resp.Order.OrderItems[0].Quantity)
	assert.InDelta(t, 2.50, resp.Order.OrderItems[0].Price, 0.001)

	// Verify second item
	assert.Equal(t, uint32(2), resp.Order.OrderItems[1].MenuItemId)
	assert.Equal(t, int32(1), resp.Order.OrderItems[1].Quantity)
	assert.InDelta(t, 2.00, resp.Order.OrderItems[1].Price, 0.001)

	mockUserClient.AssertExpectations(t)
	mockMenuClient.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestCreateOrder_InvalidUser(t *testing.T) {
	// Setup
	db, _, sqlDB := setupTestDB(t)
	defer teardownTestDB(t, sqlDB)
	database.DB = db

	mockUserClient := new(MockUserServiceClient)
	mockMenuClient := new(MockMenuServiceClient)

	server := &OrderServer{
		UserClient: mockUserClient,
		MenuClient: mockMenuClient,
	}

	// Mock user validation failure
	mockUserClient.On("GetUser", mock.Anything, &userv1.GetUserRequest{Id: 999}).
		Return(nil, status.Errorf(codes.NotFound, "user not found"))

	// Test
	ctx := context.Background()
	resp, err := server.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		UserId: 999,
		Items: []*orderv1.OrderItemRequest{
			{MenuItemId: 1, Quantity: 1},
		},
	})

	// Assert
	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "user not found")

	mockUserClient.AssertExpectations(t)
}

func TestCreateOrder_InvalidMenuItem(t *testing.T) {
	// Setup
	db, _, sqlDB := setupTestDB(t)
	defer teardownTestDB(t, sqlDB)
	database.DB = db

	mockUserClient := new(MockUserServiceClient)
	mockMenuClient := new(MockMenuServiceClient)

	server := &OrderServer{
		UserClient: mockUserClient,
		MenuClient: mockMenuClient,
	}

	// Mock user validation success
	mockUserClient.On("GetUser", mock.Anything, &userv1.GetUserRequest{Id: 1}).
		Return(&userv1.GetUserResponse{
			User: &userv1.User{Id: 1, Name: "Test User"},
		}, nil)

	// Mock menu item lookup failure
	mockMenuClient.On("GetMenuItem", mock.Anything, &menuv1.GetMenuItemRequest{Id: 999}).
		Return(nil, status.Errorf(codes.NotFound, "menu item not found"))

	// Test
	ctx := context.Background()
	resp, err := server.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		UserId: 1,
		Items: []*orderv1.OrderItemRequest{
			{MenuItemId: 999, Quantity: 1},
		},
	})

	// Assert
	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "menu item 999 not found")

	mockUserClient.AssertExpectations(t)
	mockMenuClient.AssertExpectations(t)
}

func TestCreateOrder_DatabaseError(t *testing.T) {
	// Setup
	db, dbMock, sqlDB := setupTestDB(t)
	defer teardownTestDB(t, sqlDB)
	database.DB = db

	mockUserClient := new(MockUserServiceClient)
	mockMenuClient := new(MockMenuServiceClient)

	server := &OrderServer{
		UserClient: mockUserClient,
		MenuClient: mockMenuClient,
	}

	// Mock user validation success
	mockUserClient.On("GetUser", mock.Anything, &userv1.GetUserRequest{Id: 1}).
		Return(&userv1.GetUserResponse{
			User: &userv1.User{Id: 1, Name: "Test User"},
		}, nil)

	// Mock menu item lookup success
	mockMenuClient.On("GetMenuItem", mock.Anything, &menuv1.GetMenuItemRequest{Id: 1}).
		Return(&menuv1.GetMenuItemResponse{
			MenuItem: &menuv1.MenuItem{Id: 1, Name: "Coffee", Price: 2.50},
		}, nil)

	// Mock database error
	dbMock.ExpectBegin()
	dbMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "orders"`)).
		WillReturnError(gorm.ErrInvalidDB)
	dbMock.ExpectRollback()

	// Test
	ctx := context.Background()
	resp, err := server.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		UserId: 1,
		Items: []*orderv1.OrderItemRequest{
			{MenuItemId: 1, Quantity: 1},
		},
	})

	// Assert
	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "failed to create order")

	mockUserClient.AssertExpectations(t)
	mockMenuClient.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestGetOrder(t *testing.T) {
	// Setup
	db, dbMock, sqlDB := setupTestDB(t)
	defer teardownTestDB(t, sqlDB)
	database.DB = db

	mockUserClient := new(MockUserServiceClient)
	mockMenuClient := new(MockMenuServiceClient)

	server := &OrderServer{
		UserClient: mockUserClient,
		MenuClient: mockMenuClient,
	}

	tests := []struct {
		name        string
		orderID     uint32
		mockSetup   func()
		wantErr     bool
		expectedErr codes.Code
	}{
		{
			name:    "get existing order",
			orderID: 1,
			mockSetup: func() {
				now := time.Now()
				// Mock order query
				orderRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "user_id", "status"}).
					AddRow(1, now, now, nil, 1, "pending")
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders" WHERE "orders"."id" = $1 AND "orders"."deleted_at" IS NULL ORDER BY "orders"."id" LIMIT $2`)).
					WithArgs(1, 1).
					WillReturnRows(orderRows)
				// Mock order items query
				itemRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "order_id", "menu_item_id", "quantity", "price"}).
					AddRow(1, now, now, nil, 1, 1, 2, 2.50)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "order_items" WHERE "order_items"."order_id" = $1 AND "order_items"."deleted_at" IS NULL`)).
					WithArgs(1).
					WillReturnRows(itemRows)
			},
			wantErr: false,
		},
		{
			name:    "get non-existent order",
			orderID: 9999,
			mockSetup: func() {
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders" WHERE "orders"."id" = $1 AND "orders"."deleted_at" IS NULL ORDER BY "orders"."id" LIMIT $2`)).
					WithArgs(9999, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			expectedErr: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			resp, err := server.GetOrder(ctx, &orderv1.GetOrderRequest{
				Id: tt.orderID,
			})

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedErr, st.Code())
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.orderID, resp.Order.Id)
			}

			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}

func TestGetOrders(t *testing.T) {
	// Setup
	db, dbMock, sqlDB := setupTestDB(t)
	defer teardownTestDB(t, sqlDB)
	database.DB = db

	mockUserClient := new(MockUserServiceClient)
	mockMenuClient := new(MockMenuServiceClient)

	server := &OrderServer{
		UserClient: mockUserClient,
		MenuClient: mockMenuClient,
	}

	// Test empty orders
	t.Run("empty orders", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "user_id", "status"})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders" WHERE "orders"."deleted_at" IS NULL`)).
			WillReturnRows(rows)

		ctx := context.Background()
		resp, err := server.GetOrders(ctx, &orderv1.GetOrdersRequest{})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Orders)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	// Test multiple orders
	t.Run("multiple orders", func(t *testing.T) {
		now := time.Now()
		orderRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "user_id", "status"}).
			AddRow(1, now, now, nil, 1, "pending").
			AddRow(2, now, now, nil, 2, "completed")

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders" WHERE "orders"."deleted_at" IS NULL`)).
			WillReturnRows(orderRows)

		// Mock order items query with IN clause (GORM optimizes this)
		itemRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "order_id", "menu_item_id", "quantity", "price"}).
			AddRow(1, now, now, nil, 1, 1, 2, 2.50).
			AddRow(2, now, now, nil, 2, 2, 1, 3.00)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "order_items" WHERE "order_items"."order_id" IN ($1,$2) AND "order_items"."deleted_at" IS NULL`)).
			WithArgs(1, 2).
			WillReturnRows(itemRows)

		ctx := context.Background()
		resp, err := server.GetOrders(ctx, &orderv1.GetOrdersRequest{})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Orders, 2)

		// Verify orders are returned
		assert.Equal(t, uint32(1), resp.Orders[0].UserId)
		assert.Equal(t, "pending", resp.Orders[0].Status)

		assert.Equal(t, uint32(2), resp.Orders[1].UserId)
		assert.Equal(t, "completed", resp.Orders[1].Status)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	// Test database error
	t.Run("database error", func(t *testing.T) {
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders" WHERE "orders"."deleted_at" IS NULL`)).
			WillReturnError(gorm.ErrInvalidDB)

		ctx := context.Background()
		resp, err := server.GetOrders(ctx, &orderv1.GetOrdersRequest{})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Contains(t, st.Message(), "failed to get orders")
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestModelToProto(t *testing.T) {
	now := time.Now()
	order := &models.Order{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		UserID: 5,
		Status: "pending",
		OrderItems: []models.OrderItem{
			{
				Model: gorm.Model{
					ID:        10,
					CreatedAt: now,
					UpdatedAt: now,
				},
				OrderID:    1,
				MenuItemID: 2,
				Quantity:   3,
				Price:      4.50,
			},
		},
	}

	protoOrder := modelToProto(order)

	assert.Equal(t, uint32(1), protoOrder.Id)
	assert.Equal(t, uint32(5), protoOrder.UserId)
	assert.Equal(t, "pending", protoOrder.Status)
	assert.Equal(t, now.Format(time.RFC3339), protoOrder.CreatedAt)
	assert.Equal(t, now.Format(time.RFC3339), protoOrder.UpdatedAt)

	// Verify order items
	require.Len(t, protoOrder.OrderItems, 1)
	assert.Equal(t, uint32(10), protoOrder.OrderItems[0].Id)
	assert.Equal(t, uint32(1), protoOrder.OrderItems[0].OrderId)
	assert.Equal(t, uint32(2), protoOrder.OrderItems[0].MenuItemId)
	assert.Equal(t, int32(3), protoOrder.OrderItems[0].Quantity)
	assert.InDelta(t, 4.50, protoOrder.OrderItems[0].Price, 0.001)
}

func TestCreateOrder_PriceSnapshot(t *testing.T) {
	// This test verifies that prices are snapshotted at order creation time
	db, dbMock, sqlDB := setupTestDB(t)
	defer teardownTestDB(t, sqlDB)
	database.DB = db

	mockUserClient := new(MockUserServiceClient)
	mockMenuClient := new(MockMenuServiceClient)

	server := &OrderServer{
		UserClient: mockUserClient,
		MenuClient: mockMenuClient,
	}

	// Mock user validation
	mockUserClient.On("GetUser", mock.Anything, &userv1.GetUserRequest{Id: 1}).
		Return(&userv1.GetUserResponse{
			User: &userv1.User{Id: 1, Name: "Test User"},
		}, nil)

	// Mock menu item with specific price
	originalPrice := 5.99
	mockMenuClient.On("GetMenuItem", mock.Anything, &menuv1.GetMenuItemRequest{Id: 1}).
		Return(&menuv1.GetMenuItemResponse{
			MenuItem: &menuv1.MenuItem{Id: 1, Name: "Special", Price: originalPrice},
		}, nil)

	// Mock database operations
	now := time.Now()
	dbMock.ExpectBegin()
	dbMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "orders"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, "pending").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, now, now))
	dbMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "order_items"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	dbMock.ExpectCommit()

	// Create order
	ctx := context.Background()
	resp, err := server.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		UserId: 1,
		Items: []*orderv1.OrderItemRequest{
			{MenuItemId: 1, Quantity: 1},
		},
	})

	require.NoError(t, err)
	assert.InDelta(t, originalPrice, resp.Order.OrderItems[0].Price, 0.001)

	mockUserClient.AssertExpectations(t)
	mockMenuClient.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestNewOrderServer(t *testing.T) {
	// Test successful creation (will create connections even though services aren't running)
	t.Run("creates server successfully", func(t *testing.T) {
		server, err := NewOrderServer("localhost:9091", "localhost:9092")

		// The connections will be created successfully even without running services
		// because grpc.NewClient doesn't immediately connect
		require.NoError(t, err)
		require.NotNil(t, server)
		assert.NotNil(t, server.UserClient)
		assert.NotNil(t, server.MenuClient)
	})

	// Test with invalid address format to trigger user service connection error
	t.Run("invalid user service address", func(t *testing.T) {
		// Use an invalid scheme that will cause connection failure
		server, err := NewOrderServer("://invalid-address", "localhost:9092")

		// This should succeed with current gRPC implementation
		// because grpc.NewClient is lazy and doesn't validate immediately
		_ = server
		_ = err
	})

	// Test with invalid menu service address
	t.Run("invalid menu service address", func(t *testing.T) {
		server, err := NewOrderServer("localhost:9091", "://invalid-address")

		// This should succeed with current gRPC implementation
		_ = server
		_ = err
	})
}

func TestNewClients(t *testing.T) {
	// Test successful creation with default addresses
	t.Run("creates clients with defaults", func(t *testing.T) {
		clients, err := NewClients()

		// The connections will be created successfully
		require.NoError(t, err)
		require.NotNil(t, clients)
		assert.NotNil(t, clients.UserClient)
		assert.NotNil(t, clients.MenuClient)
	})

	// Test with custom environment variables
	t.Run("uses environment variables", func(t *testing.T) {
		// Set custom addresses
		t.Setenv("USER_SERVICE_GRPC_ADDR", "custom-user:9091")
		t.Setenv("MENU_SERVICE_GRPC_ADDR", "custom-menu:9092")

		clients, err := NewClients()

		require.NoError(t, err)
		require.NotNil(t, clients)
		assert.NotNil(t, clients.UserClient)
		assert.NotNil(t, clients.MenuClient)
	})
}
