// services/users-service/main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	consulapi "github.com/hashicorp/consul/api"
)

const serviceName = "users-service"
const servicePort = 8081

func main() {
	// 1. Register this service with Consul.
	// This is a crucial step for service discovery.
	if err := registerServiceWithConsul(); err != nil {
		log.Fatalf("Failed to register service with Consul: %v", err)
	}

	// 2. Set up the HTTP router and define endpoints.
	r := chi.NewRouter()
	r.Get("/health", healthCheckHandler) // Consul uses this to check service health.
	r.Get("/users/{id}", getUserHandler)

	log.Printf("'%s' starting on port %d...", serviceName, servicePort)

	// 3. Start the HTTP server.
	if err := http.ListenAndServe(fmt.Sprintf(":%d", servicePort), r); err != nil {
		log.Fatalf("Failed to start server for service '%s': %v", serviceName, err)
	}
}

// getUserHandler provides a simple response for a user endpoint.
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	fmt.Fprintf(w, "Response from '%s': Details for user %s\n", serviceName, userID)
}

// healthCheckHandler is essential for Consul to monitor the service's status.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Service is healthy")
}

// registerServiceWithConsul handles the logic of registering with the Consul agent.
func registerServiceWithConsul() error {
	config := consulapi.DefaultConfig()
	consul, err := consulapi.NewClient(config)
	if err != nil {
		return err
	}

	// Get the hostname of the machine.
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	// Define the service registration details.
	registration := &consulapi.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%s", serviceName, hostname), // A unique ID for this instance
		Name:    serviceName,                                // The logical name of the service
		Port:    servicePort,
		Address: hostname,                                 // The address where this service is running
		Check: &consulapi.AgentServiceCheck{
			// Consul will hit this endpoint to check the service's health.
			HTTP:     fmt.Sprintf("http://%s:%d/health", hostname, servicePort),
			Interval: "10s", // Check every 10 seconds.
			Timeout:  "1s",
		},
	}

	// Perform the registration.
	if err := consul.Agent().ServiceRegister(registration); err != nil {
		return err
	}

	log.Printf("Successfully registered '%s' with Consul", serviceName)
	return nil
}