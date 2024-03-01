package main

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"github.com/13thuser/exampleauth/datastore"
	pb "github.com/13thuser/exampleauth/grpc"
	"github.com/dgrijalva/jwt-go"
)

const bufSize = 1024 * 1024 // 1MB

// createTestServer creates a new gRPC server and returns a client
// to communicate with the server
func createTestServer(ctx context.Context, t *testing.T) (pb.BookingServiceClient, func()) {
	lis := bufconn.Listen(bufSize)

	srvr := grpc.NewServer(
		grpc.UnaryInterceptor(validateTokenUnaryInterceptor),
		grpc.StreamInterceptor(validateTokenStreamInterceptor),
	)
	db := datastore.NewDatastore()
	pb.RegisterBookingServiceServer(srvr, NewBookingServer(db))

	go func(t *testing.T) {
		if err := srvr.Serve(lis); err != nil {
			t.Errorf("error serving server: %v", err)
		}
	}(t)

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.DialContext(
		ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("error connecting to server: %v", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			t.Errorf("error closing listener: %v", err)
		}
		srvr.Stop()
	}

	// create a client
	client := pb.NewBookingServiceClient(conn)

	// Return the client and the closer function
	return client, closer
}

// Create a JWT token for testing
func createTestingJWTToken(subject string, isAdmin bool) (string, error) {
	// Create the claims for the JWT token
	claims := jwt.MapClaims{
		"sub":      subject,
		"exp":      time.Now().Add(time.Hour).Unix(),
		"is_admin": isAdmin,
	}

	// Create the JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with a secret key
	tokenString, err := token.SignedString([]byte("my-secret-key"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// getCtxWithToken creates a new context with the JWT token as metadata
func getCtxWithToken(t *testing.T, ctx context.Context, subject string, isAdmin bool) context.Context {
	token, err := createTestingJWTToken(subject, isAdmin)
	if err != nil {
		t.Fatalf("Failed to create JWT token: %v", err)
	}
	return metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", token))
}

func TestBookingServer_Purchase(t *testing.T) {
	ctx := context.Background()

	// Create a test server and get the client
	client, closer := createTestServer(ctx, t)
	defer closer()

	tests := map[string]struct {
		subject string
		isAdmin bool
		user    *pb.User
		seat    *pb.Seat
		wantErr bool
	}{
		"admin purchase": {
			subject: "adminuser@example.com",
			isAdmin: true,
			user: &pb.User{
				EmailAddress: "adminuser@example.com",
				FirstName:    "admin",
				LastName:     "user",
			},
			seat: &pb.Seat{
				SectionId: "A",
				SeatId:    "1",
			},
			wantErr: false,
		},
		"non-admin user purchase": {
			subject: "user@example.com",
			isAdmin: true,
			user: &pb.User{
				EmailAddress: "user@example.com",
				FirstName:    "john",
				LastName:     "doe",
			},
			seat: &pb.Seat{
				SectionId: "A",
				SeatId:    "2",
			},
			wantErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := getCtxWithToken(t, ctx, tt.subject, tt.isAdmin)
			response, err := client.Purchase(ctx, &pb.PurchaseRequest{
				User: tt.user,
				Seat: tt.seat,
			})
			t.Logf("name: %v \n\tResponse: %+v\n", name, response)
			if (err != nil) != tt.wantErr {
				t.Errorf("Purchase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
