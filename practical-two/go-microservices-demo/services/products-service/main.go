// services/products-service/main.go
package main

// ... (same imports as users-service)
import (
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/go-chi/chi/v5"
    consulapi "github.com/hashicorp/consul/api"
)

const serviceName = "products-service"
const servicePort = 8082 // <-- Different Port

func main() {
    if err := registerServiceWithConsul(); err != nil {
        log.Fatalf("Failed to register service with Consul: %v", err)
    }

    r := chi.NewRouter()
    r.Get("/health", healthCheckHandler)
    r.Get("/products/{id}", getProductHandler) // <-- Different Endpoint

    log.Printf("'%s' starting on port %d...", serviceName, servicePort)
    if err := http.ListenAndServe(fmt.Sprintf(":%d", servicePort), r); err != nil {
        log.Fatalf("Failed to start server for service '%s': %v", serviceName, err)
    }
}

// getProductHandler provides a simple response for a product endpoint.
func getProductHandler(w http.ResponseWriter, r *http.Request) {
    productID := chi.URLParam(r, "id")
    // <-- Different Response Message
    fmt.Fprintf(w, "Response from '%s': Details for product %s\n", serviceName, productID)
}

// healthCheckHandler and registerServiceWithConsul functions are IDENTICAL
// to the ones in users-service, just with the new consts.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Service is healthy")
}

func registerServiceWithConsul() error {
    // ... (This function's code is exactly the same)
    config := consulapi.DefaultConfig()
	consul, err := consulapi.NewClient(config)
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	registration := &consulapi.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%s", serviceName, hostname),
		Name:    serviceName,
		Port:    servicePort,
		Address: hostname,
		Check: &consulapi.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", hostname, servicePort),
			Interval: "10s",
			Timeout:  "1s",
		},
	}

	if err := consul.Agent().ServiceRegister(registration); err != nil {
		return err
	}

	log.Printf("Successfully registered '%s' with Consul", serviceName)
	return nil
}