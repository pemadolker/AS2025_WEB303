// api-gateway/main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
)

const gatewayPort = 8080

func main() {
	// The handler function is responsible for all routing logic.
	http.HandleFunc("/", routeRequest)

	log.Printf("API Gateway starting on port %d...", gatewayPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", gatewayPort), nil); err != nil {
		log.Fatalf("Failed to start API Gateway: %v", err)
	}
}

// routeRequest determines the target service from the request path.
func routeRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("Gateway received request for: %s", r.URL.Path)

	// A simple routing rule: /api/users/* goes to "users-service".
	// And /api/products/* goes to "products-service".
	pathSegments := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(pathSegments) < 3 || pathSegments[0] != "api" {
		http.Error(w, "Invalid request path", http.StatusBadRequest)
		return
	}
	serviceName := pathSegments[1] + "-service" // e.g., "users" -> "users-service"

	// Step 1: Discover the service's location using Consul.
	serviceURL, err := discoverService(serviceName)
	if err != nil {
		log.Printf("Error discovering service '%s': %v", serviceName, err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	log.Printf("Discovered '%s' at %s", serviceName, serviceURL)

	// Step 2: Create a reverse proxy to forward the request.
	proxy := httputil.NewSingleHostReverseProxy(serviceURL)

	// Step 3: Rewrite the request URL to be sent to the downstream service.
	// We strip `/api/<service-name>` from the original path.
	// e.g., /api/users/123 becomes /users/123
	r.URL.Path = "/" + strings.Join(pathSegments[1:], "/")
	log.Printf("Forwarding request to: %s%s", serviceURL, r.URL.Path)

	// Step 4: Forward the request.
	proxy.ServeHTTP(w, r)
}

// discoverService queries Consul to find a healthy instance of a service.
func discoverService(name string) (*url.URL, error) {
	config := consulapi.DefaultConfig()
	consul, err := consulapi.NewClient(config)
	if err != nil {
		return nil, err
	}

	// Query Consul's health endpoint for healthy service instances.
	services, _, err := consul.Health().Service(name, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("could not query Consul for service '%s': %w", name, err)
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no healthy instances of service '%s' found in Consul", name)
	}

	// For simplicity, we use the first available healthy instance.
	// In a production system, you might use a load-balancing strategy here.
	service := services[0].Service
	serviceAddress := fmt.Sprintf("http://%s:%d", service.Address, service.Port)

	return url.Parse(serviceAddress)
}