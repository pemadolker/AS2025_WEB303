package grpc

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"
	"user-service/database"
	"user-service/models"

	"github.com/DATA-DOG/go-sqlmock"
	userv1 "github.com/douglasswm/student-cafe-protos/gen/go/user/v1"
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

func TestCreateUser(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer teardownTestDB(t, sqlDB)
	database.DB = db

	server := NewUserServer()

	tests := []struct {
		name        string
		request     *userv1.CreateUserRequest
		wantErr     bool
		expectedMsg string
	}{
		{
			name: "successful user creation",
			request: &userv1.CreateUserRequest{
				Name:        "John Doe",
				Email:       "john@example.com",
				IsCafeOwner: false,
			},
			wantErr: false,
		},
		{
			name: "create cafe owner",
			request: &userv1.CreateUserRequest{
				Name:        "Jane Owner",
				Email:       "jane@cafeshop.com",
				IsCafeOwner: true,
			},
			wantErr: false,
		},
		{
			name: "empty name should still work (validation is optional)",
			request: &userv1.CreateUserRequest{
				Name:        "",
				Email:       "test@example.com",
				IsCafeOwner: false,
			},
			wantErr: false,
		},
	}

	// Test database error
	t.Run("database error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
			WillReturnError(gorm.ErrInvalidDB)
		mock.ExpectRollback()

		ctx := context.Background()
		resp, err := server.CreateUser(ctx, &userv1.CreateUserRequest{
			Name:  "Test",
			Email: "test@example.com",
		})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Contains(t, err.Error(), "failed to create user")
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the INSERT query
			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), tt.request.Name, tt.request.Email, tt.request.IsCafeOwner).
				WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "is_cafe_owner"}).
					AddRow(1, time.Now(), time.Now(), nil, tt.request.Name, tt.request.Email, tt.request.IsCafeOwner))
			mock.ExpectCommit()

			ctx := context.Background()
			resp, err := server.CreateUser(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.NotZero(t, resp.User.Id)
				assert.Equal(t, tt.request.Name, resp.User.Name)
				assert.Equal(t, tt.request.Email, resp.User.Email)
				assert.Equal(t, tt.request.IsCafeOwner, resp.User.IsCafeOwner)
				assert.NotEmpty(t, resp.User.CreatedAt)
				assert.NotEmpty(t, resp.User.UpdatedAt)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetUser(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer teardownTestDB(t, sqlDB)
	database.DB = db

	server := NewUserServer()

	tests := []struct {
		name        string
		userID      uint32
		mockSetup   func()
		wantErr     bool
		expectedErr codes.Code
	}{
		{
			name:   "get existing user",
			userID: 1,
			mockSetup: func() {
				now := time.Now()
				rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "is_cafe_owner"}).
					AddRow(1, now, now, nil, "Test User", "test@example.com", false)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:   "get non-existent user",
			userID: 9999,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
					WithArgs(9999, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			expectedErr: codes.NotFound,
		},
		{
			name:   "get user with ID 0",
			userID: 0,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
					WithArgs(0, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			expectedErr: codes.NotFound,
		},
		{
			name:   "database error",
			userID: 1,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
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
			resp, err := server.GetUser(ctx, &userv1.GetUserRequest{
				Id: tt.userID,
			})

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedErr, st.Code())
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.userID, resp.User.Id)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetUsers(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer teardownTestDB(t, sqlDB)
	database.DB = db

	server := NewUserServer()

	// Test empty database
	t.Run("empty database", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "is_cafe_owner"})
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
			WillReturnRows(rows)

		ctx := context.Background()
		resp, err := server.GetUsers(ctx, &userv1.GetUsersRequest{})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test multiple users
	t.Run("multiple users", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "is_cafe_owner"}).
			AddRow(1, now, now, nil, "User 1", "user1@example.com", false).
			AddRow(2, now, now, nil, "User 2", "user2@example.com", true).
			AddRow(3, now, now, nil, "User 3", "user3@example.com", false)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
			WillReturnRows(rows)

		ctx := context.Background()
		resp, err := server.GetUsers(ctx, &userv1.GetUsersRequest{})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Users, 3)

		// Verify all users are returned
		assert.Equal(t, "User 1", resp.Users[0].Name)
		assert.Equal(t, "user1@example.com", resp.Users[0].Email)
		assert.False(t, resp.Users[0].IsCafeOwner)

		assert.Equal(t, "User 2", resp.Users[1].Name)
		assert.Equal(t, "user2@example.com", resp.Users[1].Email)
		assert.True(t, resp.Users[1].IsCafeOwner)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test database error
	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
			WillReturnError(gorm.ErrInvalidDB)

		ctx := context.Background()
		resp, err := server.GetUsers(ctx, &userv1.GetUsersRequest{})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Contains(t, err.Error(), "failed to get users")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestModelToProto(t *testing.T) {
	now := time.Now()
	user := &models.User{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Name:        "Test User",
		Email:       "test@example.com",
		IsCafeOwner: true,
	}

	protoUser := modelToProto(user)

	assert.Equal(t, uint32(1), protoUser.Id)
	assert.Equal(t, "Test User", protoUser.Name)
	assert.Equal(t, "test@example.com", protoUser.Email)
	assert.Equal(t, true, protoUser.IsCafeOwner)
	assert.Equal(t, now.Format(time.RFC3339), protoUser.CreatedAt)
	assert.Equal(t, now.Format(time.RFC3339), protoUser.UpdatedAt)
}
