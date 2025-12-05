package grpc

import (
	"context"
	"database/sql"
	"menu-service/database"
	"menu-service/models"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	menuv1 "github.com/douglasswm/student-cafe-protos/gen/go/menu/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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

func TestCreateMenuItem(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer teardownTestDB(t, sqlDB)
	database.DB = db

	server := NewMenuServer()

	tests := []struct {
		name    string
		request *menuv1.CreateMenuItemRequest
		wantErr bool
	}{
		{
			name: "successful menu item creation",
			request: &menuv1.CreateMenuItemRequest{
				Name:        "Cappuccino",
				Description: "Espresso with steamed milk and foam",
				Price:       4.50,
			},
			wantErr: false,
		},
		{
			name: "create item with zero price",
			request: &menuv1.CreateMenuItemRequest{
				Name:        "Water",
				Description: "Free water",
				Price:       0.0,
			},
			wantErr: false,
		},
		{
			name: "create item with long description",
			request: &menuv1.CreateMenuItemRequest{
				Name:        "Special Brew",
				Description: "A very long description that describes the coffee in great detail with many words",
				Price:       5.99,
			},
			wantErr: false,
		},
	}

	// Test database error
	t.Run("database error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "menu_items"`)).
			WillReturnError(gorm.ErrInvalidDB)
		mock.ExpectRollback()

		ctx := context.Background()
		resp, err := server.CreateMenuItem(ctx, &menuv1.CreateMenuItemRequest{
			Name:        "Test",
			Description: "Test desc",
			Price:       1.00,
		})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Contains(t, err.Error(), "failed to create menu item")
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()

			// Mock the INSERT query
			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "menu_items"`)).
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), tt.request.Name, tt.request.Description, tt.request.Price).
				WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
					AddRow(1, now, now))
			mock.ExpectCommit()

			ctx := context.Background()
			resp, err := server.CreateMenuItem(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.NotZero(t, resp.MenuItem.Id)
				assert.Equal(t, tt.request.Name, resp.MenuItem.Name)
				assert.Equal(t, tt.request.Description, resp.MenuItem.Description)
				assert.InDelta(t, tt.request.Price, resp.MenuItem.Price, 0.001)
				assert.NotEmpty(t, resp.MenuItem.CreatedAt)
				assert.NotEmpty(t, resp.MenuItem.UpdatedAt)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetMenuItem(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer teardownTestDB(t, sqlDB)
	database.DB = db

	server := NewMenuServer()

	tests := []struct {
		name        string
		itemID      uint32
		mockSetup   func()
		wantErr     bool
		expectedErr codes.Code
	}{
		{
			name:   "get existing menu item",
			itemID: 1,
			mockSetup: func() {
				now := time.Now()
				rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "description", "price"}).
					AddRow(1, now, now, nil, "Latte", "Espresso with steamed milk", 4.00)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items" WHERE "menu_items"."id" = $1 AND "menu_items"."deleted_at" IS NULL ORDER BY "menu_items"."id" LIMIT $2`)).
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:   "get non-existent menu item",
			itemID: 9999,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items" WHERE "menu_items"."id" = $1 AND "menu_items"."deleted_at" IS NULL ORDER BY "menu_items"."id" LIMIT $2`)).
					WithArgs(9999, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			expectedErr: codes.NotFound,
		},
		{
			name:   "get item with ID 0",
			itemID: 0,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items" WHERE "menu_items"."id" = $1 AND "menu_items"."deleted_at" IS NULL ORDER BY "menu_items"."id" LIMIT $2`)).
					WithArgs(0, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			expectedErr: codes.NotFound,
		},
		{
			name:   "database error",
			itemID: 1,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items" WHERE "menu_items"."id" = $1 AND "menu_items"."deleted_at" IS NULL ORDER BY "menu_items"."id" LIMIT $2`)).
					WithArgs(1, 1).
					WillReturnError(gorm.ErrInvalidDB)
			},
			wantErr:     true,
			expectedErr: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			resp, err := server.GetMenuItem(ctx, &menuv1.GetMenuItemRequest{
				Id: tt.itemID,
			})

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedErr, st.Code())
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.itemID, resp.MenuItem.Id)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetMenu(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer teardownTestDB(t, sqlDB)
	database.DB = db

	server := NewMenuServer()

	// Test empty menu
	t.Run("empty menu", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "description", "price"})
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items"`)).
			WillReturnRows(rows)

		ctx := context.Background()
		resp, err := server.GetMenu(ctx, &menuv1.GetMenuRequest{})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.MenuItems)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test multiple items
	t.Run("multiple items", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "description", "price"}).
			AddRow(1, now, now, nil, "Coffee", "Black coffee", 2.50).
			AddRow(2, now, now, nil, "Tea", "Green tea", 2.00).
			AddRow(3, now, now, nil, "Sandwich", "Ham and cheese", 5.50)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items"`)).
			WillReturnRows(rows)

		ctx := context.Background()
		resp, err := server.GetMenu(ctx, &menuv1.GetMenuRequest{})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.MenuItems, 3)

		// Verify all items are returned
		assert.Equal(t, "Coffee", resp.MenuItems[0].Name)
		assert.Equal(t, "Black coffee", resp.MenuItems[0].Description)
		assert.InDelta(t, 2.50, resp.MenuItems[0].Price, 0.001)

		assert.Equal(t, "Tea", resp.MenuItems[1].Name)
		assert.Equal(t, "Green tea", resp.MenuItems[1].Description)
		assert.InDelta(t, 2.00, resp.MenuItems[1].Price, 0.001)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test database error
	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items"`)).
			WillReturnError(gorm.ErrInvalidDB)

		ctx := context.Background()
		resp, err := server.GetMenu(ctx, &menuv1.GetMenuRequest{})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Contains(t, err.Error(), "failed to get menu")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestModelToProto(t *testing.T) {
	now := time.Now()
	item := &models.MenuItem{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Name:        "Test Item",
		Description: "Test Description",
		Price:       3.99,
	}

	protoItem := modelToProto(item)

	assert.Equal(t, uint32(1), protoItem.Id)
	assert.Equal(t, "Test Item", protoItem.Name)
	assert.Equal(t, "Test Description", protoItem.Description)
	assert.InDelta(t, 3.99, protoItem.Price, 0.001)
	assert.Equal(t, now.Format(time.RFC3339), protoItem.CreatedAt)
	assert.Equal(t, now.Format(time.RFC3339), protoItem.UpdatedAt)
}

func TestPriceHandling(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer teardownTestDB(t, sqlDB)
	database.DB = db

	server := NewMenuServer()

	// Test various price formats
	testCases := []struct {
		name  string
		price float64
	}{
		{"integer price", 5.0},
		{"two decimal places", 5.99},
		{"three decimal places", 5.999},
		{"very small price", 0.01},
		{"large price", 999.99},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now()

			// Mock the INSERT query
			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "menu_items"`)).
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "Test Item", "Price test", tc.price).
				WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
					AddRow(1, now, now))
			mock.ExpectCommit()

			ctx := context.Background()
			resp, err := server.CreateMenuItem(ctx, &menuv1.CreateMenuItemRequest{
				Name:        "Test Item",
				Description: "Price test",
				Price:       tc.price,
			})

			require.NoError(t, err)
			assert.InDelta(t, tc.price, resp.MenuItem.Price, 0.001)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
