package handlers

import (
	"net/http"

	"api-gateway/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handlers holds the HTTP handlers and gRPC clients
type Handlers struct {
	clients *grpc.ServiceClients
}

// NewHandlers creates a new Handlers instance with gRPC clients
func NewHandlers(clients *grpc.ServiceClients) *Handlers {
	return &Handlers{clients: clients}
}

// handleGRPCError converts gRPC errors to appropriate HTTP status codes
func handleGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		// Not a gRPC error, return generic internal server error
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Map gRPC status codes to HTTP status codes
	var httpStatus int
	switch st.Code() {
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
	case codes.AlreadyExists:
		httpStatus = http.StatusConflict
	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
	case codes.FailedPrecondition:
		httpStatus = http.StatusPreconditionFailed
	case codes.Unimplemented:
		httpStatus = http.StatusNotImplemented
	case codes.Unavailable:
		httpStatus = http.StatusServiceUnavailable
	default:
		httpStatus = http.StatusInternalServerError
	}

	http.Error(w, st.Message(), httpStatus)
}
