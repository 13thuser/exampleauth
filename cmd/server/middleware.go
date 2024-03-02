package main

import (
	"context"
	"log"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/dgrijalva/jwt-go"
)

// Context key for the email ID and is_admin claims
type contextKey string

const (
	emailIDKey contextKey = "email_id"
	isAdminKey contextKey = "is_admin"
)

// You can also use a configuration file or environment variables
var PublicURLs = []string{"/BookingService/Purchase"}

// tokenValidator is a helper function to validate the JWT token
func tokenValidator(ctx context.Context) (context.Context, error) {
	// Extract the JWT token from the gRPC metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}
	token := md.Get("authorization")
	if len(token) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	// Parse the JWT token and extract the claims
	parsedToken, err := jwt.Parse(token[0], func(token *jwt.Token) (interface{}, error) {
		// Provide the key or public key to verify the token's signature
		// You can customize this based on your JWT implementation
		return []byte(JWT_SECRET_KEY), nil
	})
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to parse JWT token: %v", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return nil, status.Errorf(codes.Unauthenticated, "invalid JWT token")
	}

	log.Printf("Claims: %v\n", claims)

	// Access the sub and is_admin claims and add them to the context
	ctx = context.WithValue(ctx, emailIDKey, claims["sub"])
	ctx = context.WithValue(ctx, isAdminKey, claims["is_admin"])

	return ctx, nil
}

// Token validation interceptor for unary RPCs
func validateTokenUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if _, ok := metadata.FromIncomingContext(ctx); !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	// check if incoming request is a public URL
	for _, url := range PublicURLs {
		if strings.Contains(info.FullMethod, url) {
			return handler(ctx, req)
		}
	}

	// Validate the token and create a new context
	ctx, err := tokenValidator(ctx)
	if err != nil {
		return nil, err
	}

	// Call the next handler if the token is valid
	return handler(ctx, req)
}

// Wrapped stream to update the context
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the updated context
func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// Token validation interceptor for streaming RPCs
func validateTokenStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// Extract the metadata from the context
	_, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return status.Errorf(codes.DataLoss, "myStreamInterceptor: failed to get metadata from context")
	}

	// Validate the token and create a new context
	newCtx, err := tokenValidator(ss.Context())
	if err != nil {
		return err
	}

	// Create a new stream with the updated context
	wrapped := &wrappedStream{ServerStream: ss, ctx: newCtx}

	// Call the next handler in the chain with the wrapped stream
	return handler(srv, wrapped)
}
