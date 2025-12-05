package integration

import (
	"context"
	"net"
	"os"
	"testing"

	menuv1 "github.com/douglasswm/student-cafe-protos/gen/go/menu/v1"
	orderv1 "github.com/douglasswm/student-cafe-protos/gen/go/order/v1"
	userv1 "github.com/douglasswm/student-cafe-protos/gen/go/user/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	// Import actual service implementations
	menudatabase "menu-service/database"
	menugrpc "menu-service/grpc"
	menumodels "menu-service/models"

	orderdatabase "order-service/database"
	ordergrpc "order-service/grpc"
	ordermodels "order-service/models"

	userdatabase "user-service/database"
	usergrpc "user-service/grpc"
	usermodels "user-service/models"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const bufSize = 1024 * 1024

var (
	userListener  *bufconn.Listener
	menuListener  *bufconn.Listener
	orderListener *bufconn.Listener
)

// setupUserService creates and starts the user service
func setupUserService(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&usermodels.User{})
	require.NoError(t, err)

	userdatabase.DB = db

	// Create gRPC server with bufconn
	userListener = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	userv1.RegisterUserServiceServer(s, usergrpc.NewUserServer())

	go func() {
		_ = s.Serve(userListener)
	}()
}

// setupMenuService creates and starts the menu service
func setupMenuService(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&menumodels.MenuItem{})
	require.NoError(t, err)

	menudatabase.DB = db

	// Create gRPC server with bufconn
	menuListener = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	menuv1.RegisterMenuServiceServer(s, menugrpc.NewMenuServer())

	go func() {
		_ = s.Serve(menuListener)
	}()
}

// setupOrderService creates and starts the order service with clients to user and menu services
func setupOrderService(t *testing.T, userConn, menuConn *grpc.ClientConn) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&ordermodels.Order{}, &ordermodels.OrderItem{})
	require.NoError(t, err)

	orderdatabase.DB = db

	// Create order server with injected clients
	orderServer := &ordergrpc.OrderServer{
		UserClient: userv1.NewUserServiceClient(userConn),
		MenuClient: menuv1.NewMenuServiceClient(menuConn),
	}

	// Create gRPC server with bufconn
	orderListener = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	orderv1.RegisterOrderServiceServer(s, orderServer)

	go func() {
		_ = s.Serve(orderListener)
	}()
}

// bufDialer returns a dialer function for bufconn
func bufDialer(listener *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, url string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestMain(m *testing.M) {
	// Exit code
	code := m.Run()
	os.Exit(code)
}

func TestIntegration_CreateUser(t *testing.T) {
	setupUserService(t)

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer(userListener)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := userv1.NewUserServiceClient(conn)

	// Create user
	resp, err := client.CreateUser(ctx, &userv1.CreateUserRequest{
		Name:        "Integration Test User",
		Email:       "integration@test.com",
		IsCafeOwner: false,
	})

	require.NoError(t, err)
	assert.NotZero(t, resp.User.Id)
	assert.Equal(t, "Integration Test User", resp.User.Name)
	assert.Equal(t, "integration@test.com", resp.User.Email)
}

func TestIntegration_CreateAndGetMenuItem(t *testing.T) {
	setupMenuService(t)

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer(menuListener)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := menuv1.NewMenuServiceClient(conn)

	// Create menu item
	createResp, err := client.CreateMenuItem(ctx, &menuv1.CreateMenuItemRequest{
		Name:        "Integration Coffee",
		Description: "Test coffee",
		Price:       3.50,
	})

	require.NoError(t, err)
	assert.NotZero(t, createResp.MenuItem.Id)

	// Get menu item
	getResp, err := client.GetMenuItem(ctx, &menuv1.GetMenuItemRequest{
		Id: createResp.MenuItem.Id,
	})

	require.NoError(t, err)
	assert.Equal(t, createResp.MenuItem.Id, getResp.MenuItem.Id)
	assert.Equal(t, "Integration Coffee", getResp.MenuItem.Name)
	assert.InDelta(t, 3.50, getResp.MenuItem.Price, 0.001)
}

func TestIntegration_CompleteOrderFlow(t *testing.T) {
	// Setup all three services
	setupUserService(t)
	setupMenuService(t)

	ctx := context.Background()

	// Connect to user service
	userConn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer(userListener)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer userConn.Close()

	userClient := userv1.NewUserServiceClient(userConn)

	// Connect to menu service
	menuConn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer(menuListener)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer menuConn.Close()

	menuClient := menuv1.NewMenuServiceClient(menuConn)

	// Setup order service with connections to user and menu services
	setupOrderService(t, userConn, menuConn)

	// Connect to order service
	orderConn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer(orderListener)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer orderConn.Close()

	orderClient := orderv1.NewOrderServiceClient(orderConn)

	// Step 1: Create a user
	userResp, err := userClient.CreateUser(ctx, &userv1.CreateUserRequest{
		Name:        "Order Test User",
		Email:       "order@test.com",
		IsCafeOwner: false,
	})
	require.NoError(t, err)
	userID := userResp.User.Id

	// Step 2: Create menu items
	item1, err := menuClient.CreateMenuItem(ctx, &menuv1.CreateMenuItemRequest{
		Name:        "Coffee",
		Description: "Hot coffee",
		Price:       2.50,
	})
	require.NoError(t, err)

	item2, err := menuClient.CreateMenuItem(ctx, &menuv1.CreateMenuItemRequest{
		Name:        "Sandwich",
		Description: "Ham sandwich",
		Price:       5.00,
	})
	require.NoError(t, err)

	// Step 3: Create an order
	orderResp, err := orderClient.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		UserId: userID,
		Items: []*orderv1.OrderItemRequest{
			{MenuItemId: item1.MenuItem.Id, Quantity: 2},
			{MenuItemId: item2.MenuItem.Id, Quantity: 1},
		},
	})

	require.NoError(t, err)
	assert.NotZero(t, orderResp.Order.Id)
	assert.Equal(t, userID, orderResp.Order.UserId)
	assert.Equal(t, "pending", orderResp.Order.Status)
	assert.Len(t, orderResp.Order.OrderItems, 2)

	// Verify prices were snapshotted
	assert.InDelta(t, 2.50, orderResp.Order.OrderItems[0].Price, 0.001)
	assert.InDelta(t, 5.00, orderResp.Order.OrderItems[1].Price, 0.001)

	// Step 4: Retrieve the order
	getOrderResp, err := orderClient.GetOrder(ctx, &orderv1.GetOrderRequest{
		Id: orderResp.Order.Id,
	})

	require.NoError(t, err)
	assert.Equal(t, orderResp.Order.Id, getOrderResp.Order.Id)
	assert.Len(t, getOrderResp.Order.OrderItems, 2)
}

func TestIntegration_OrderValidation(t *testing.T) {
	// Setup all three services
	setupUserService(t)
	setupMenuService(t)

	ctx := context.Background()

	// Connect to services
	userConn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer(userListener)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer userConn.Close()

	menuConn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer(menuListener)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer menuConn.Close()

	setupOrderService(t, userConn, menuConn)

	orderConn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer(orderListener)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer orderConn.Close()

	orderClient := orderv1.NewOrderServiceClient(orderConn)

	// Try to create order with invalid user
	t.Run("invalid user", func(t *testing.T) {
		_, err := orderClient.CreateOrder(ctx, &orderv1.CreateOrderRequest{
			UserId: 9999,
			Items: []*orderv1.OrderItemRequest{
				{MenuItemId: 1, Quantity: 1},
			},
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})

	// Create a valid user
	userClient := userv1.NewUserServiceClient(userConn)
	userResp, err := userClient.CreateUser(ctx, &userv1.CreateUserRequest{
		Name:        "Valid User",
		Email:       "valid@test.com",
		IsCafeOwner: false,
	})
	require.NoError(t, err)

	// Try to create order with invalid menu item
	t.Run("invalid menu item", func(t *testing.T) {
		_, err := orderClient.CreateOrder(ctx, &orderv1.CreateOrderRequest{
			UserId: userResp.User.Id,
			Items: []*orderv1.OrderItemRequest{
				{MenuItemId: 9999, Quantity: 1},
			},
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "menu item 9999 not found")
	})
}

func TestIntegration_ConcurrentOrders(t *testing.T) {
	// Setup all services
	setupUserService(t)
	defer userListener.Close()

	setupMenuService(t)
	defer menuListener.Close()

	ctx := context.Background()

	userConn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer(userListener)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer userConn.Close()

	menuConn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer(menuListener)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer menuConn.Close()

	setupOrderService(t, userConn, menuConn)

	orderConn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer(orderListener)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer orderConn.Close()

	// Create test data
	userClient := userv1.NewUserServiceClient(userConn)
	menuClient := menuv1.NewMenuServiceClient(menuConn)
	orderClient := orderv1.NewOrderServiceClient(orderConn)

	userResp, err := userClient.CreateUser(ctx, &userv1.CreateUserRequest{
		Name:        "Concurrent User",
		Email:       "concurrent@test.com",
		IsCafeOwner: false,
	})
	require.NoError(t, err)

	itemResp, err := menuClient.CreateMenuItem(ctx, &menuv1.CreateMenuItemRequest{
		Name:        "Test Item",
		Description: "For concurrent testing",
		Price:       1.00,
	})
	require.NoError(t, err)

	// Create multiple orders concurrently
	numOrders := 10
	errChan := make(chan error, numOrders)
	respChan := make(chan *orderv1.CreateOrderResponse, numOrders)

	for i := 0; i < numOrders; i++ {
		go func() {
			resp, err := orderClient.CreateOrder(ctx, &orderv1.CreateOrderRequest{
				UserId: userResp.User.Id,
				Items: []*orderv1.OrderItemRequest{
					{MenuItemId: itemResp.MenuItem.Id, Quantity: 1},
				},
			})
			errChan <- err
			respChan <- resp
		}()
	}

	// Collect results
	// Note: SQLite may have locking issues with concurrent writes,
	// so we expect some failures in this test
	successCount := 0
	for i := 0; i < numOrders; i++ {
		err := <-errChan
		resp := <-respChan
		if err == nil && resp != nil {
			successCount++
			assert.NotZero(t, resp.Order.Id)
		}
	}

	// With SQLite, we expect at least some orders to succeed
	// In production with PostgreSQL, all should succeed
	assert.GreaterOrEqual(t, successCount, numOrders/2,
		"At least half of concurrent orders should succeed (SQLite has known locking limitations)")
}
