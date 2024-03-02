package main

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/13thuser/exampleauth/grpc"
)

func createJWTToken(subject string, isAdmin bool) (string, error) {
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

func main() {
	// Create a gRPC client
	// conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create a new instance of the gRPC client
	c := pb.NewBookingServiceClient(conn)

	// Prepare the JWT token
	token, err := createJWTToken("test@example.com", true)
	if err != nil {
		log.Fatalf("Failed to create JWT token: %v", err)
	}

	// Create a new context with the JWT token as metadata
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))

	response, err := c.Purchase(ctx, &pb.PurchaseRequest{
		User: &pb.User{
			EmailAddress: "test@example.com",
			FirstName:    "test",
			LastName:     "user",
		},
		Seat: &pb.Seat{
			SectionId: "A",
			SeatId:    "1",
		},
	})
	if err != nil {
		// Handle error
		statusErr, ok := status.FromError(err)
		if ok && statusErr.Code() == codes.Unauthenticated {
			log.Fatalf("Authentication failed: %v", err)
		} else {
			log.Fatalf("RPC error: %v", err)
		}
	}

	// Process the response
	log.Printf("Response: %v", response)

	stream, err := c.GetBookingsBySection(ctx, &pb.GetBookingsBySectionRequest{Section: "A"})
	if err != nil {
		log.Fatalf("Failed to get bookings by section: %v", err)
	}
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Failed to receive a response: %v", err)
		}
		log.Printf("Received streaming message: %v\n", response)
	}

	// Remove the user from the train
	// Create a new context with the JWT token as metadata
	ctx = metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))

	bookingID := response.BookingId
	if _, err := c.RemoveUserFromTrain(ctx, &pb.RemoveBookingRequest{BookingId: bookingID}); err != nil {
		log.Fatalf("Failed to remove user from train: %v", err)
	}
	log.Printf("Removed user from train: %v", bookingID)

	response, err = c.Purchase(ctx, &pb.PurchaseRequest{
		User: &pb.User{
			EmailAddress: "test@example.com",
			FirstName:    "test",
			LastName:     "user",
		},
		Seat: &pb.Seat{
			SectionId: "A",
			SeatId:    "2",
		},
	})
	if err != nil {
		// Handle error
		statusErr, ok := status.FromError(err)
		if ok && statusErr.Code() == codes.Unauthenticated {
			log.Fatalf("Authentication failed: %v", err)
		} else {
			log.Fatalf("RPC error: %v", err)
		}
	}

	// Process the response
	log.Printf("Response: %v", response)

	bookingID = response.BookingId
	// Modify seat
	response, err = c.ModifySeat(ctx, &pb.ModifySeatRequest{
		BookingId:    bookingID,
		NewSeatId:    "3",
		NewSectionId: "A",
	})
	if err != nil {
		log.Fatalf("Failed to modify seat: %v", err)
	}
	log.Printf("Modified seat: %v", response)

}
