package model

import (
	"net/http"
	"time"
)

// RequestType identifies the type of request being logged.
type RequestType string

const (
	// TypeHTTP represents a standard HTTP request.
	TypeHTTP RequestType = "http"

	// TypeGRPC represents a gRPC request.
	TypeGRPC RequestType = "grpc"
)

// GRPCMethodType identifies the type of gRPC method.
type GRPCMethodType string

const (
	// UnaryMethod represents a unary gRPC method.
	UnaryMethod GRPCMethodType = "unary"

	// ClientStreamMethod represents a client streaming gRPC method.
	ClientStreamMethod GRPCMethodType = "client_stream"

	// ServerStreamMethod represents a server streaming gRPC method.
	ServerStreamMethod GRPCMethodType = "server_stream"

	// BidiStreamMethod represents a bidirectional streaming gRPC method.
	BidiStreamMethod GRPCMethodType = "bidi_stream"
)

// RequestLog represents a captured HTTP or gRPC request.
type RequestLog struct {
	ID        string      `json:"ID"`
	Type      RequestType `json:"Type"`
	Timestamp time.Time   `json:"Timestamp"`
	Duration  int64       `json:"Duration"`
	Error     string      `json:"Error,omitempty"`

	// HTTP-specific fields
	Method          string                   `json:"Method,omitempty"`
	Path            string                   `json:"Path,omitempty"`
	Query           string                   `json:"Query,omitempty"`
	RequestHeaders  http.Header              `json:"RequestHeaders,omitempty"`
	ResponseHeaders http.Header              `json:"ResponseHeaders,omitempty"`
	StatusCode      int                      `json:"StatusCode,omitempty"`
	RequestBody     string                   `json:"RequestBody,omitempty"`
	ResponseBody    string                   `json:"ResponseBody,omitempty"`
	MiddlewareTrace []map[string]interface{} `json:"MiddlewareTrace,omitempty"`
	RouteTrace      map[string]interface{}   `json:"RouteTrace,omitempty"`

	// gRPC-specific fields
	GRPCService      string              `json:"GRPCService,omitempty"`
	GRPCMethod       string              `json:"GRPCMethod,omitempty"`
	GRPCMethodType   GRPCMethodType      `json:"GRPCMethodType,omitempty"`
	GRPCStatusCode   int32               `json:"GRPCStatusCode,omitempty"`
	GRPCStatusDesc   string              `json:"GRPCStatusDesc,omitempty"`
	GRPCPeer         string              `json:"GRPCPeer,omitempty"`
	GRPCRequestMD    map[string][]string `json:"GRPCRequestMD,omitempty"`
	GRPCResponseMD   map[string][]string `json:"GRPCResponseMD,omitempty"`
	GRPCRequestData  string              `json:"GRPCRequestData,omitempty"`
	GRPCResponseData string              `json:"GRPCResponseData,omitempty"`
	GRPCMessages     []GRPCMessage       `json:"GRPCMessages,omitempty"`
}

// GRPCMessage represents a single message in a streaming gRPC call.
type GRPCMessage struct {
	Timestamp time.Time           `json:"Timestamp"`
	Direction string              `json:"Direction"` // "sent" or "received"
	Data      string              `json:"Data,omitempty"`
	Metadata  map[string][]string `json:"Metadata,omitempty"`
}

// NewHTTPRequestLog creates a new request log entry for an HTTP request.
func NewHTTPRequestLog(req *http.Request) *RequestLog {
	return &RequestLog{
		ID:             generateID(),
		Type:           TypeHTTP,
		Timestamp:      time.Now(),
		Method:         req.Method,
		Path:           req.URL.Path,
		Query:          req.URL.RawQuery,
		RequestHeaders: req.Header,
	}
}

// NewGRPCRequestLog creates a new request log entry for a gRPC request.
func NewGRPCRequestLog(service, method string, methodType GRPCMethodType) *RequestLog {
	return &RequestLog{
		ID:             generateID(),
		Type:           TypeGRPC,
		Timestamp:      time.Now(),
		GRPCService:    service,
		GRPCMethod:     method,
		GRPCMethodType: methodType,
		GRPCMessages:   make([]GRPCMessage, 0),
	}
}

// generateID creates a unique ID for a request log.
func generateID() string {
	return time.Now().Format("20060102-150405.000000")
}
