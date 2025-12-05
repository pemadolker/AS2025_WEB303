package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	consulapi "github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "practical-three/proto/gen"
)

// Remove global clients - we'll create them dynamically per request

// A struct to hold the aggregated data
type UserPurchaseData struct {
	User    *pb.User    `json:"user"`
	Product *pb.Product `json:"product"`
}

// Function to discover service address from Consul
func discoverService(serviceName string) (string, error) {
	config := consulapi.DefaultConfig()
	config.Address = "consul:8500" // Use Docker service name
	consul, err := consulapi.NewClient(config)
	if err != nil {
		return "", err
	}

	services, _, err := consul.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", err
	}

	if len(services) == 0 {
		return "", fmt.Errorf("no healthy instances of service %s found", serviceName)
	}

	// Use the first healthy service instance
	service := services[0]
	address := fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port)
	return address, nil
}

// Function to get a users service client by discovering it from Consul
func getUsersServiceClient() (pb.UserServiceClient, *grpc.ClientConn, error) {
	log.Println("Discovering users-service from Consul...")
	usersServiceAddr, err := discoverService("users-service")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to discover users-service: %v", err)
	}
	log.Printf("Discovered users-service at: %s", usersServiceAddr)

	conn, err := grpc.Dial(usersServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to users-service: %v", err)
	}

	client := pb.NewUserServiceClient(conn)
	return client, conn, nil
}

// Function to get a products service client by discovering it from Consul
func getProductsServiceClient() (pb.ProductServiceClient, *grpc.ClientConn, error) {
	log.Println("Discovering products-service from Consul...")
	productsServiceAddr, err := discoverService("products-service")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to discover products-service: %v", err)
	}
	log.Printf("Discovered products-service at: %s", productsServiceAddr)

	conn, err := grpc.Dial(productsServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to products-service: %v", err)
	}

	client := pb.NewProductServiceClient(conn)
	return client, conn, nil
}

func main() {
	r := mux.NewRouter()
	// User routes
	r.HandleFunc("/api/users", createUserHandler).Methods("POST")
	r.HandleFunc("/api/users/{id}", getUserHandler).Methods("GET")
	// Product routes
	r.HandleFunc("/api/products", createProductHandler).Methods("POST")
	r.HandleFunc("/api/products/{id}", getProductHandler).Methods("GET")

	// The new endpoint to get combined data
	r.HandleFunc("/api/purchases/user/{userId}/product/{productId}", getPurchaseDataHandler).Methods("GET")

	log.Println("API Gateway listening on port 8080...")
	log.Println("Service discovery will be performed on each request via Consul")
	http.ListenAndServe(":8080", r)
}

// User Handlers
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	// Discover and connect to users-service via Consul
	usersClient, conn, err := getUsersServiceClient()
	if err != nil {
		http.Error(w, fmt.Sprintf("Service discovery failed: %v", err), http.StatusServiceUnavailable)
		return
	}
	defer conn.Close()

	var req pb.CreateUserRequest
	json.NewDecoder(r.Body).Decode(&req)
	res, err := usersClient.CreateUser(context.Background(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.User)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	// Discover and connect to users-service via Consul
	usersClient, conn, err := getUsersServiceClient()
	if err != nil {
		http.Error(w, fmt.Sprintf("Service discovery failed: %v", err), http.StatusServiceUnavailable)
		return
	}
	defer conn.Close()

	vars := mux.Vars(r)
	id := vars["id"]
	res, err := usersClient.GetUser(context.Background(), &pb.GetUserRequest{Id: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.User)
}

// Product Handlers
func createProductHandler(w http.ResponseWriter, r *http.Request) {
	// Discover and connect to products-service via Consul
	productsClient, conn, err := getProductsServiceClient()
	if err != nil {
		http.Error(w, fmt.Sprintf("Service discovery failed: %v", err), http.StatusServiceUnavailable)
		return
	}
	defer conn.Close()

	var req pb.CreateProductRequest
	json.NewDecoder(r.Body).Decode(&req)
	res, err := productsClient.CreateProduct(context.Background(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.Product)
}

func getProductHandler(w http.ResponseWriter, r *http.Request) {
	// Discover and connect to products-service via Consul
	productsClient, conn, err := getProductsServiceClient()
	if err != nil {
		http.Error(w, fmt.Sprintf("Service discovery failed: %v", err), http.StatusServiceUnavailable)
		return
	}
	defer conn.Close()

	vars := mux.Vars(r)
	id := vars["id"]
	res, err := productsClient.GetProduct(context.Background(), &pb.GetProductRequest{Id: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.Product)
}

// New handler for combined data
func getPurchaseDataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]
	productId := vars["productId"]

	var wg sync.WaitGroup
	var user *pb.User
	var product *pb.Product
	var userErr, productErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		// Discover and connect to users-service via Consul
		usersClient, conn, err := getUsersServiceClient()
		if err != nil {
			userErr = fmt.Errorf("failed to discover users-service: %v", err)
			return
		}
		defer conn.Close()

		res, err := usersClient.GetUser(context.Background(), &pb.GetUserRequest{Id: userId})
		if err != nil {
			userErr = err
			return
		}
		user = res.User
	}()

	go func() {
		defer wg.Done()
		// Discover and connect to products-service via Consul
		productsClient, conn, err := getProductsServiceClient()
		if err != nil {
			productErr = fmt.Errorf("failed to discover products-service: %v", err)
			return
		}
		defer conn.Close()

		res, err := productsClient.GetProduct(context.Background(), &pb.GetProductRequest{Id: productId})
		if err != nil {
			productErr = err
			return
		}
		product = res.Product
	}()

	wg.Wait()

	if userErr != nil || productErr != nil {
		errMsg := "Could not retrieve all data"
		if userErr != nil {
			errMsg += fmt.Sprintf(" - User error: %v", userErr)
		}
		if productErr != nil {
			errMsg += fmt.Sprintf(" - Product error: %v", productErr)
		}
		http.Error(w, errMsg, http.StatusNotFound)
		return
	}

	purchaseData := UserPurchaseData{
		User:    user,
		Product: product,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(purchaseData)
}
