package agent

import (
	"context"
	"encoding/json"
	"path"
	"strings"
	"time"

	"github.com/doganarif/govisual/internal/model"
	"github.com/doganarif/govisual/pkg/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// GRPCAgentConfig contains configuration options specific to gRPC agents.
type GRPCAgentConfig struct {
	AgentConfig

	// LogRequestData determines whether request message data is logged.
	LogRequestData bool

	// LogResponseData determines whether response message data is logged.
	LogResponseData bool

	// IgnoreMethods is a list of method patterns to ignore.
	IgnoreMethods []string
}

// GRPCAgent is an agent that collects data from gRPC services.
type GRPCAgent struct {
	*BaseAgent
	config GRPCAgentConfig
}

// NewGRPCAgent creates a new gRPC agent with the given transport.
func NewGRPCAgent(transport transport.Transport, opts ...GRPCOption) *GRPCAgent {
	config := GRPCAgentConfig{
		AgentConfig: AgentConfig{
			Transport: transport,
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(&config)
	}

	return &GRPCAgent{
		BaseAgent: NewBaseAgent("grpc", config.AgentConfig),
		config:    config,
	}
}

// GRPCOption is a function that configures a gRPC agent.
type GRPCOption func(*GRPCAgentConfig)

// WithGRPCRequestDataLogging enables or disables logging of gRPC request message data.
func WithGRPCRequestDataLogging(enabled bool) GRPCOption {
	return func(c *GRPCAgentConfig) {
		c.LogRequestData = enabled
	}
}

// WithGRPCResponseDataLogging enables or disables logging of gRPC response message data.
func WithGRPCResponseDataLogging(enabled bool) GRPCOption {
	return func(c *GRPCAgentConfig) {
		c.LogResponseData = enabled
	}
}

// WithIgnoreGRPCMethods sets the gRPC method patterns to ignore.
func WithIgnoreGRPCMethods(patterns ...string) GRPCOption {
	return func(c *GRPCAgentConfig) {
		c.IgnoreMethods = append(c.IgnoreMethods, patterns...)
	}
}

// UnaryServerInterceptor returns a gRPC unary server interceptor for data collection.
func (a *GRPCAgent) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip if method should be ignored
		if a.shouldIgnoreMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// Extract service and method names
		service, method := parseFullMethod(info.FullMethod)

		// Create a new request log
		reqLog := model.NewGRPCRequestLog(service, method, model.UnaryMethod)

		// Extract request metadata
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			reqLog.GRPCRequestMD = metadataToMap(md)
		}

		// Extract peer information
		reqLog.GRPCPeer = extractPeerAddress(ctx)

		// Log request message if enabled
		reqLog.GRPCRequestData = marshalMessage(req, a.config.LogRequestData)

		// Record start time
		startTime := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Record duration
		reqLog.Duration = time.Since(startTime).Milliseconds()

		// Log status code and description
		st, _ := status.FromError(err)
		reqLog.GRPCStatusCode = int32(st.Code())
		reqLog.GRPCStatusDesc = st.Message()

		// Log response message if enabled
		if err == nil {
			reqLog.GRPCResponseData = marshalMessage(resp, a.config.LogResponseData)
		} else {
			reqLog.Error = err.Error()
		}

		// Process the request log
		a.Process(reqLog)

		return resp, err
	}
}

// StreamServerInterceptor returns a gRPC stream server interceptor for data collection.
func (a *GRPCAgent) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Skip if method should be ignored
		if a.shouldIgnoreMethod(info.FullMethod) {
			return handler(srv, ss)
		}

		// Extract service and method names
		service, method := parseFullMethod(info.FullMethod)

		// Determine method type
		methodType := getMethodType(info.IsClientStream, info.IsServerStream)

		// Create a new request log
		reqLog := model.NewGRPCRequestLog(service, method, methodType)

		// Extract request metadata
		if md, ok := metadata.FromIncomingContext(ss.Context()); ok {
			reqLog.GRPCRequestMD = metadataToMap(md)
		}

		// Extract peer information
		reqLog.GRPCPeer = extractPeerAddress(ss.Context())

		// Create a wrapper around the server stream
		wrappedStream := &wrappedServerStream{
			ServerStream:    ss,
			agent:           a,
			reqLog:          reqLog,
			logRequestData:  a.config.LogRequestData,
			logResponseData: a.config.LogResponseData,
		}

		// Record start time
		startTime := time.Now()

		// Call the handler
		err := handler(srv, wrappedStream)

		// Record duration
		reqLog.Duration = time.Since(startTime).Milliseconds()

		// Log status code and description
		st, _ := status.FromError(err)
		reqLog.GRPCStatusCode = int32(st.Code())
		reqLog.GRPCStatusDesc = st.Message()

		if err != nil {
			reqLog.Error = err.Error()
		}

		// Process the request log
		a.Process(reqLog)

		return err
	}
}

// UnaryClientInterceptor returns a gRPC unary client interceptor for data collection.
func (a *GRPCAgent) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Skip if method should be ignored
		if a.shouldIgnoreMethod(method) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		// Extract service and method names
		service, methodName := parseFullMethod(method)

		// Create a new request log
		reqLog := model.NewGRPCRequestLog(service, methodName, model.UnaryMethod)

		// Extract request metadata
		if md, ok := metadata.FromOutgoingContext(ctx); ok {
			reqLog.GRPCRequestMD = metadataToMap(md)
		}

		// Log request message if enabled
		reqLog.GRPCRequestData = marshalMessage(req, a.config.LogRequestData)

		// Set peer information to the target
		reqLog.GRPCPeer = cc.Target()

		// Record start time
		startTime := time.Now()

		// Create metadata for receiving headers and trailers
		var responseHeader, responseTrailer metadata.MD
		opts = append(opts,
			grpc.Header(&responseHeader),
			grpc.Trailer(&responseTrailer),
		)

		// Call the invoker
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Record duration
		reqLog.Duration = time.Since(startTime).Milliseconds()

		// Log response metadata
		reqLog.GRPCResponseMD = metadataToMap(responseHeader)

		// Log status code and description
		st, _ := status.FromError(err)
		reqLog.GRPCStatusCode = int32(st.Code())
		reqLog.GRPCStatusDesc = st.Message()

		// Log response message if enabled
		if err == nil {
			reqLog.GRPCResponseData = marshalMessage(reply, a.config.LogResponseData)
		} else {
			reqLog.Error = err.Error()
		}

		// Process the request log
		a.Process(reqLog)

		return err
	}
}

// StreamClientInterceptor returns a gRPC stream client interceptor for data collection.
func (a *GRPCAgent) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// Skip if method should be ignored
		if a.shouldIgnoreMethod(method) {
			return streamer(ctx, desc, cc, method, opts...)
		}

		// Extract service and method names
		service, methodName := parseFullMethod(method)

		// Determine method type
		methodType := getMethodType(desc.ClientStreams, desc.ServerStreams)

		// Create a new request log
		reqLog := model.NewGRPCRequestLog(service, methodName, methodType)

		// Extract request metadata
		if md, ok := metadata.FromOutgoingContext(ctx); ok {
			reqLog.GRPCRequestMD = metadataToMap(md)
		}

		// Set peer information to the target
		reqLog.GRPCPeer = cc.Target()

		// Record start time
		startTime := time.Now()

		// Create metadata for receiving headers and trailers
		var responseHeader metadata.MD
		opts = append(opts, grpc.Header(&responseHeader))

		// Call the streamer
		clientStream, err := streamer(ctx, desc, cc, method, opts...)

		if err != nil {
			// Record duration
			reqLog.Duration = time.Since(startTime).Milliseconds()

			// Log error details
			reqLog.Error = err.Error()
			st, _ := status.FromError(err)
			reqLog.GRPCStatusCode = int32(st.Code())
			reqLog.GRPCStatusDesc = st.Message()

			// Process the request log
			a.Process(reqLog)

			return nil, err
		}

		// Create a wrapper around the client stream
		wrappedStream := &wrappedClientStream{
			ClientStream:    clientStream,
			agent:           a,
			reqLog:          reqLog,
			startTime:       startTime,
			responseHeader:  responseHeader,
			logRequestData:  a.config.LogRequestData,
			logResponseData: a.config.LogResponseData,
		}

		return wrappedStream, nil
	}
}

// shouldIgnoreMethod checks if a method should be ignored.
func (a *GRPCAgent) shouldIgnoreMethod(fullMethod string) bool {
	for _, pattern := range a.config.IgnoreMethods {
		// Check for exact match
		if pattern == fullMethod {
			return true
		}

		// Check for service-wide ignore with trailing slash
		if strings.HasSuffix(pattern, "/") && strings.HasPrefix(fullMethod, pattern) {
			return true
		}

		// Simple path matching
		matched, _ := path.Match(pattern, fullMethod)
		if matched {
			return true
		}
	}
	return false
}

// wrappedServerStream wraps a grpc.ServerStream to intercept and log messages.
type wrappedServerStream struct {
	grpc.ServerStream
	agent           *GRPCAgent
	reqLog          *model.RequestLog
	logRequestData  bool
	logResponseData bool
}

// RecvMsg intercepts and logs incoming messages.
func (w *wrappedServerStream) RecvMsg(m interface{}) error {
	err := w.ServerStream.RecvMsg(m)

	if err == nil && w.logRequestData {
		// Log the received message
		messageData := marshalMessage(m, true)

		w.reqLog.GRPCMessages = append(w.reqLog.GRPCMessages, model.GRPCMessage{
			Timestamp: time.Now(),
			Direction: "received",
			Data:      messageData,
		})

		// For the first message in client streaming, also set the request data
		if w.reqLog.GRPCRequestData == "" {
			w.reqLog.GRPCRequestData = messageData
		}
	}

	return err
}

// SendMsg intercepts and logs outgoing messages.
func (w *wrappedServerStream) SendMsg(m interface{}) error {
	// Log the sent message before sending
	if w.logResponseData {
		messageData := marshalMessage(m, true)

		w.reqLog.GRPCMessages = append(w.reqLog.GRPCMessages, model.GRPCMessage{
			Timestamp: time.Now(),
			Direction: "sent",
			Data:      messageData,
		})

		// For the first message in server streaming, also set the response data
		if w.reqLog.GRPCResponseData == "" {
			w.reqLog.GRPCResponseData = messageData
		}
	}

	return w.ServerStream.SendMsg(m)
}

// wrappedClientStream wraps a grpc.ClientStream to intercept and log messages.
type wrappedClientStream struct {
	grpc.ClientStream
	agent           *GRPCAgent
	reqLog          *model.RequestLog
	startTime       time.Time
	responseHeader  metadata.MD
	logRequestData  bool
	logResponseData bool
	finished        bool
}

// RecvMsg intercepts and logs incoming messages.
func (w *wrappedClientStream) RecvMsg(m interface{}) error {
	err := w.ClientStream.RecvMsg(m)

	if err == nil && w.logResponseData {
		// Log the received message
		messageData := marshalMessage(m, true)

		w.reqLog.GRPCMessages = append(w.reqLog.GRPCMessages, model.GRPCMessage{
			Timestamp: time.Now(),
			Direction: "received",
			Data:      messageData,
		})

		// For the first message in server streaming, also set the response data
		if w.reqLog.GRPCResponseData == "" {
			w.reqLog.GRPCResponseData = messageData
		}

		// Update response metadata if it has changed
		if header, err := w.ClientStream.Header(); err == nil {
			w.reqLog.GRPCResponseMD = metadataToMap(header)
		}
	} else if err != nil {
		w.finishStreamWithError(err)
	}

	return err
}

// SendMsg intercepts and logs outgoing messages.
func (w *wrappedClientStream) SendMsg(m interface{}) error {
	// Log the sent message before sending
	if w.logRequestData {
		messageData := marshalMessage(m, true)

		w.reqLog.GRPCMessages = append(w.reqLog.GRPCMessages, model.GRPCMessage{
			Timestamp: time.Now(),
			Direction: "sent",
			Data:      messageData,
		})

		// For the first message in client streaming, also set the request data
		if w.reqLog.GRPCRequestData == "" {
			w.reqLog.GRPCRequestData = messageData
		}
	}

	err := w.ClientStream.SendMsg(m)
	if err != nil {
		w.finishStreamWithError(err)
	}

	return err
}

// CloseSend intercepts the close send call.
func (w *wrappedClientStream) CloseSend() error {
	err := w.ClientStream.CloseSend()

	// When client closes the send direction, we don't yet finish the request
	// as we might still receive messages from the server

	return err
}

// Header intercepts the header call.
func (w *wrappedClientStream) Header() (metadata.MD, error) {
	md, err := w.ClientStream.Header()
	if err == nil {
		w.reqLog.GRPCResponseMD = metadataToMap(md)
	}
	return md, err
}

// Trailer intercepts the trailer call.
func (w *wrappedClientStream) Trailer() metadata.MD {
	md := w.ClientStream.Trailer()
	return md
}

// finishStreamWithError finalizes logging for a stream with an error.
func (w *wrappedClientStream) finishStreamWithError(err error) {
	if !w.finished {
		w.finished = true

		// Record duration
		w.reqLog.Duration = time.Since(w.startTime).Milliseconds()

		// Log status code and description
		st, _ := status.FromError(err)
		w.reqLog.GRPCStatusCode = int32(st.Code())
		w.reqLog.GRPCStatusDesc = st.Message()

		if err != nil && err != context.Canceled {
			w.reqLog.Error = err.Error()
		}

		// If error is EOF, it's a normal stream end, set success code
		if strings.Contains(err.Error(), "EOF") {
			w.reqLog.GRPCStatusCode = int32(codes.OK)
			w.reqLog.GRPCStatusDesc = "OK"
			w.reqLog.Error = ""
		}

		// Process the request log
		w.agent.Process(w.reqLog)
	}
}

// Helper functions

// parseFullMethod parses the full method string (/service/method) into service and method components.
func parseFullMethod(fullMethod string) (service, method string) {
	if fullMethod == "" {
		return "", ""
	}

	// Remove leading slash
	fullMethod = strings.TrimPrefix(fullMethod, "/")

	// Split into service and method
	parts := strings.Split(fullMethod, "/")
	if len(parts) != 2 {
		return fullMethod, ""
	}

	return parts[0], parts[1]
}

// getMethodType determines the type of gRPC method based on the stream info.
func getMethodType(isClientStream, isServerStream bool) model.GRPCMethodType {
	switch {
	case isClientStream && isServerStream:
		return model.BidiStreamMethod
	case isClientStream:
		return model.ClientStreamMethod
	case isServerStream:
		return model.ServerStreamMethod
	default:
		return model.UnaryMethod
	}
}

// extractPeerAddress extracts the peer address from the context.
func extractPeerAddress(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "unknown"
	}
	return p.Addr.String()
}

// metadataToMap converts metadata to a map.
func metadataToMap(md metadata.MD) map[string][]string {
	if md == nil {
		return nil
	}

	result := make(map[string][]string, len(md))
	for k, v := range md {
		result[k] = v
	}
	return result
}

// marshalMessage attempts to marshal a message to JSON for logging.
func marshalMessage(message interface{}, shouldLog bool) string {
	if !shouldLog || message == nil {
		return ""
	}

	data, err := json.Marshal(message)
	if err != nil {
		return "[failed to marshal: " + err.Error() + "]"
	}
	return string(data)
}
