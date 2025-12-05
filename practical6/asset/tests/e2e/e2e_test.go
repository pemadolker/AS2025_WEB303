package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test configuration
var (
	apiGatewayURL = getEnv("API_GATEWAY_URL", "http://localhost:8081")
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Response structures
type User struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	IsCafeOwner bool   `json:"is_cafe_owner"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type MenuItem struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type OrderItem struct {
	ID         uint    `json:"id"`
	OrderID    uint    `json:"order_id"`
	MenuItemID uint    `json:"menu_item_id"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

type Order struct {
	ID         uint        `json:"id"`
	UserID     uint        `json:"user_id"`
	Status     string      `json:"status"`
	OrderItems []OrderItem `json:"order_items"`
	CreatedAt  string      `json:"created_at"`
	UpdatedAt  string      `json:"updated_at"`
}

// Helper functions
func makeRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, apiGatewayURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

func TestMain(m *testing.M) {
	// Wait for services to be ready
	fmt.Println("Waiting for services to be ready...")
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(apiGatewayURL + "/api/users")
		if err == nil && resp.StatusCode < 500 {
			resp.Body.Close()
			fmt.Println("Services are ready!")
			break
		}
		if i == maxRetries-1 {
			fmt.Println("Services did not become ready in time")
			os.Exit(1)
		}
		time.Sleep(2 * time.Second)
	}

	// Run tests
	code := m.Run()
	os.Exit(code)
}

func TestE2E_HealthCheck(t *testing.T) {
	// Verify all services are accessible through the gateway
	resp, err := makeRequest("GET", "/api/users", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Less(t, resp.StatusCode, 500, "API Gateway should be accessible")
}

func TestE2E_CreateUser(t *testing.T) {
	reqBody := map[string]interface{}{
		"name":          "E2E Test User",
		"email":         fmt.Sprintf("e2e-%d@test.com", time.Now().Unix()),
		"is_cafe_owner": false,
	}

	resp, err := makeRequest("POST", "/api/users", reqBody)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	require.NoError(t, err)

	assert.NotZero(t, user.ID)
	assert.Equal(t, reqBody["name"], user.Name)
	assert.Equal(t, reqBody["email"], user.Email)
	assert.Equal(t, reqBody["is_cafe_owner"], user.IsCafeOwner)
	assert.NotEmpty(t, user.CreatedAt)
	assert.NotEmpty(t, user.UpdatedAt)
}

func TestE2E_GetUsers(t *testing.T) {
	// Create a user first
	reqBody := map[string]interface{}{
		"name":          "Get Users Test",
		"email":         fmt.Sprintf("getusers-%d@test.com", time.Now().Unix()),
		"is_cafe_owner": false,
	}

	createResp, err := makeRequest("POST", "/api/users", reqBody)
	require.NoError(t, err)
	createResp.Body.Close()

	// Get all users
	resp, err := makeRequest("GET", "/api/users", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var users []User
	err = json.NewDecoder(resp.Body).Decode(&users)
	require.NoError(t, err)

	assert.NotEmpty(t, users)
}

func TestE2E_GetUserByID(t *testing.T) {
	// Create a user
	reqBody := map[string]interface{}{
		"name":          "Get User By ID Test",
		"email":         fmt.Sprintf("getuserbyid-%d@test.com", time.Now().Unix()),
		"is_cafe_owner": true,
	}

	createResp, err := makeRequest("POST", "/api/users", reqBody)
	require.NoError(t, err)
	defer createResp.Body.Close()

	var createdUser User
	err = json.NewDecoder(createResp.Body).Decode(&createdUser)
	require.NoError(t, err)

	// Get user by ID
	resp, err := makeRequest("GET", fmt.Sprintf("/api/users/%d", createdUser.ID), nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	require.NoError(t, err)

	assert.Equal(t, createdUser.ID, user.ID)
	assert.Equal(t, createdUser.Name, user.Name)
	assert.Equal(t, createdUser.Email, user.Email)
}

func TestE2E_CreateMenuItem(t *testing.T) {
	reqBody := map[string]interface{}{
		"name":        "E2E Coffee",
		"description": "End-to-end test coffee",
		"price":       4.50,
	}

	resp, err := makeRequest("POST", "/api/menu", reqBody)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var item MenuItem
	err = json.NewDecoder(resp.Body).Decode(&item)
	require.NoError(t, err)

	assert.NotZero(t, item.ID)
	assert.Equal(t, reqBody["name"], item.Name)
	assert.Equal(t, reqBody["description"], item.Description)
	assert.InDelta(t, reqBody["price"], item.Price, 0.01)
}

func TestE2E_GetMenu(t *testing.T) {
	// Create a menu item first
	reqBody := map[string]interface{}{
		"name":        "Get Menu Test Item",
		"description": "Test item for get menu",
		"price":       3.00,
	}

	createResp, err := makeRequest("POST", "/api/menu", reqBody)
	require.NoError(t, err)
	createResp.Body.Close()

	// Get all menu items
	resp, err := makeRequest("GET", "/api/menu", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var items []MenuItem
	err = json.NewDecoder(resp.Body).Decode(&items)
	require.NoError(t, err)

	assert.NotEmpty(t, items)
}

func TestE2E_CompleteOrderFlow(t *testing.T) {
	// Step 1: Create a user
	userReq := map[string]interface{}{
		"name":          "Order Flow User",
		"email":         fmt.Sprintf("orderflow-%d@test.com", time.Now().Unix()),
		"is_cafe_owner": false,
	}

	userResp, err := makeRequest("POST", "/api/users", userReq)
	require.NoError(t, err)
	defer userResp.Body.Close()

	var user User
	err = json.NewDecoder(userResp.Body).Decode(&user)
	require.NoError(t, err)

	// Step 2: Create menu items
	item1Req := map[string]interface{}{
		"name":        fmt.Sprintf("Coffee-%d", time.Now().Unix()),
		"description": "Hot coffee",
		"price":       2.50,
	}

	item1Resp, err := makeRequest("POST", "/api/menu", item1Req)
	require.NoError(t, err)
	defer item1Resp.Body.Close()

	var item1 MenuItem
	err = json.NewDecoder(item1Resp.Body).Decode(&item1)
	require.NoError(t, err)

	item2Req := map[string]interface{}{
		"name":        fmt.Sprintf("Sandwich-%d", time.Now().Unix()),
		"description": "Ham sandwich",
		"price":       5.00,
	}

	item2Resp, err := makeRequest("POST", "/api/menu", item2Req)
	require.NoError(t, err)
	defer item2Resp.Body.Close()

	var item2 MenuItem
	err = json.NewDecoder(item2Resp.Body).Decode(&item2)
	require.NoError(t, err)

	// Step 3: Create an order
	orderReq := map[string]interface{}{
		"user_id": user.ID,
		"items": []map[string]interface{}{
			{"menu_item_id": item1.ID, "quantity": 2},
			{"menu_item_id": item2.ID, "quantity": 1},
		},
	}

	orderResp, err := makeRequest("POST", "/api/orders", orderReq)
	require.NoError(t, err)
	defer orderResp.Body.Close()

	assert.Equal(t, http.StatusCreated, orderResp.StatusCode)

	var order Order
	err = json.NewDecoder(orderResp.Body).Decode(&order)
	require.NoError(t, err)

	assert.NotZero(t, order.ID)
	assert.Equal(t, user.ID, order.UserID)
	assert.Equal(t, "pending", order.Status)
	assert.Len(t, order.OrderItems, 2)

	// Verify prices were snapshotted
	assert.InDelta(t, 2.50, order.OrderItems[0].Price, 0.01)
	assert.InDelta(t, 5.00, order.OrderItems[1].Price, 0.01)

	// Step 4: Retrieve the order
	getOrderResp, err := makeRequest("GET", fmt.Sprintf("/api/orders/%d", order.ID), nil)
	require.NoError(t, err)
	defer getOrderResp.Body.Close()

	assert.Equal(t, http.StatusOK, getOrderResp.StatusCode)

	var retrievedOrder Order
	err = json.NewDecoder(getOrderResp.Body).Decode(&retrievedOrder)
	require.NoError(t, err)

	assert.Equal(t, order.ID, retrievedOrder.ID)
	assert.Len(t, retrievedOrder.OrderItems, 2)
}

func TestE2E_OrderValidation(t *testing.T) {
	// Try to create order with invalid user
	t.Run("invalid user", func(t *testing.T) {
		orderReq := map[string]interface{}{
			"user_id": 999999,
			"items": []map[string]interface{}{
				{"menu_item_id": 1, "quantity": 1},
			},
		}

		resp, err := makeRequest("POST", "/api/orders", orderReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Create a valid user for next test
	userReq := map[string]interface{}{
		"name":          "Validation Test User",
		"email":         fmt.Sprintf("validation-%d@test.com", time.Now().Unix()),
		"is_cafe_owner": false,
	}

	userResp, err := makeRequest("POST", "/api/users", userReq)
	require.NoError(t, err)
	defer userResp.Body.Close()

	var user User
	err = json.NewDecoder(userResp.Body).Decode(&user)
	require.NoError(t, err)

	// Try to create order with invalid menu item
	t.Run("invalid menu item", func(t *testing.T) {
		orderReq := map[string]interface{}{
			"user_id": user.ID,
			"items": []map[string]interface{}{
				{"menu_item_id": 999999, "quantity": 1},
			},
		}

		resp, err := makeRequest("POST", "/api/orders", orderReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestE2E_GetNonExistentUser(t *testing.T) {
	resp, err := makeRequest("GET", "/api/users/999999", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestE2E_GetNonExistentMenuItem(t *testing.T) {
	resp, err := makeRequest("GET", "/api/menu/999999", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestE2E_GetNonExistentOrder(t *testing.T) {
	resp, err := makeRequest("GET", "/api/orders/999999", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestE2E_ConcurrentOrders(t *testing.T) {
	// Create test user and menu item
	userReq := map[string]interface{}{
		"name":          "Concurrent Test User",
		"email":         fmt.Sprintf("concurrent-%d@test.com", time.Now().Unix()),
		"is_cafe_owner": false,
	}

	userResp, err := makeRequest("POST", "/api/users", userReq)
	require.NoError(t, err)
	defer userResp.Body.Close()

	var user User
	err = json.NewDecoder(userResp.Body).Decode(&user)
	require.NoError(t, err)

	itemReq := map[string]interface{}{
		"name":        fmt.Sprintf("Concurrent Item-%d", time.Now().Unix()),
		"description": "For concurrent testing",
		"price":       1.00,
	}

	itemResp, err := makeRequest("POST", "/api/menu", itemReq)
	require.NoError(t, err)
	defer itemResp.Body.Close()

	var item MenuItem
	err = json.NewDecoder(itemResp.Body).Decode(&item)
	require.NoError(t, err)

	// Create multiple orders concurrently
	numOrders := 10
	results := make(chan error, numOrders)

	for i := 0; i < numOrders; i++ {
		go func() {
			orderReq := map[string]interface{}{
				"user_id": user.ID,
				"items": []map[string]interface{}{
					{"menu_item_id": item.ID, "quantity": 1},
				},
			}

			resp, err := makeRequest("POST", "/api/orders", orderReq)
			if resp != nil {
				resp.Body.Close()
			}

			if err != nil {
				results <- err
			} else if resp.StatusCode != http.StatusCreated {
				results <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			} else {
				results <- nil
			}
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < numOrders; i++ {
		err := <-results
		if err == nil {
			successCount++
		} else {
			t.Logf("Concurrent order %d failed: %v", i, err)
		}
	}

	// We expect most orders to succeed
	assert.GreaterOrEqual(t, successCount, numOrders-2, "Most concurrent orders should succeed")
}
