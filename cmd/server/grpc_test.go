package main

import (
	"context"
	"io"
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
func createTestServer(t *testing.T, ctx context.Context, db *datastore.Datastore) (pb.BookingServiceClient, func()) {
	lis := bufconn.Listen(bufSize)

	srvr := grpc.NewServer(
		grpc.UnaryInterceptor(validateTokenUnaryInterceptor),
		grpc.StreamInterceptor(validateTokenStreamInterceptor),
	)
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
	db := datastore.NewDatastore(
		datastore.WithSections("A"),
		datastore.WithSectionSize(2))
	client, closer := createTestServer(t, ctx, db)
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
		"invalid section": {
			subject: "user@example.com",
			isAdmin: false,
			user: &pb.User{
				EmailAddress: "user@example.com",
				FirstName:    "john",
				LastName:     "doe",
			},
			seat: &pb.Seat{
				SectionId: "C", // invalid section
				SeatId:    "1",
			},
			wantErr: true,
		},
		"invalid seat": {
			subject: "user@example.com",
			isAdmin: false,
			user: &pb.User{
				EmailAddress: "user@example.com",
				FirstName:    "john",
				LastName:     "doe",
			},
			seat: &pb.Seat{
				SectionId: "A",
				SeatId:    "3", // invalid seat
			},
			wantErr: true,
		},
		"non-admin user purchase": {
			subject: "user@example.com",
			isAdmin: false,
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

func TestBookingServer_PurchaseWhenSectionIsFull(t *testing.T) {
	ctx := context.Background()

	// Create a test server and get the client
	db := datastore.NewDatastore(
		datastore.WithSections("A", "B"),
		datastore.WithSectionSize(1))
	client, closer := createTestServer(t, ctx, db)
	defer closer()

	tests := map[string]struct {
		subject string
		isAdmin bool
		user    *pb.User
		seat    *pb.Seat
		wantErr bool
	}{
		"when space is available": {
			subject: "user@example.com",
			isAdmin: false,
			user: &pb.User{
				EmailAddress: "user@example.com",
				FirstName:    "john",
				LastName:     "doe",
			},
			seat: &pb.Seat{
				SectionId: "A",
				SeatId:    "1",
			},
			wantErr: false,
		},
		"section is full": {
			subject: "user@example.com",
			isAdmin: false,
			user: &pb.User{
				EmailAddress: "user@example.com",
				FirstName:    "john",
				LastName:     "doe",
			},
			seat: &pb.Seat{
				SectionId: "A",
				SeatId:    "2",
			},
			wantErr: true,
		},
		"book in another section": {
			subject: "user@example.com",
			isAdmin: false,
			user: &pb.User{
				EmailAddress: "user@example.com",
				FirstName:    "john",
				LastName:     "doe",
			},
			seat: &pb.Seat{
				SectionId: "B",
				SeatId:    "1",
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

func TestBookingServer_GetBookingsBySection(t *testing.T) {
	ctx := context.Background()

	// Create a test server and get the client
	db := datastore.NewDatastore(
		datastore.WithSections("A", "B"),
		datastore.WithSectionSize(2))
	client, closer := createTestServer(t, ctx, db)
	defer closer()

	tests := map[string]struct {
		subject      string
		isAdmin      bool
		user         *pb.User
		seat         *pb.Seat
		querySection string
		wantErr      bool
	}{
		"admin gets bookings by section": {
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
			querySection: "A",
			wantErr:      false,
		},
		"non-admin cannot get bookings by section": {
			subject: "user@example.com",
			isAdmin: false,
			user: &pb.User{
				EmailAddress: "user@example.com",
				FirstName:    "john",
				LastName:     "doe",
			},
			seat: &pb.Seat{
				SectionId: "A",
				SeatId:    "2",
			},
			querySection: "A",
			wantErr:      true,
		},
	}

	numBookings := 0

	// Create bookings
	for _, tt := range tests {
		ctx := getCtxWithToken(t, ctx, tt.subject, tt.isAdmin)
		_, err := client.Purchase(ctx, &pb.PurchaseRequest{
			User: tt.user,
			Seat: tt.seat,
		})
		if err != nil {
			t.Fatalf("Purchase() error = %v, wantErr %v", err, tt.wantErr)
		}
		t.Logf("Created booking for user: %v", tt.user.EmailAddress)
		numBookings++
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := getCtxWithToken(t, ctx, tt.subject, tt.isAdmin)
			stream, err := client.GetBookingsBySection(ctx, &pb.GetBookingsBySectionRequest{
				Section: tt.querySection,
			})
			if err != nil {
				t.Fatalf("unable to get stream for GetBookingsBySection: %v", err)
			}
			if tt.wantErr {
				_, err := stream.Recv()
				if err == nil {
					t.Errorf("GetBookingsBySection() error = %v, wantErr %v", err, tt.wantErr)
				}
				// t.Logf("GetBookingsBySection() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				gotNumBookings := 0
				for {
					booking, err := stream.Recv()
					if err == io.EOF {
						break
					} else if err != nil {
						break
					}
					t.Logf("Booking: %+v", booking)
					gotNumBookings++
				}
				if gotNumBookings != numBookings {
					t.Errorf("GetBookingsBySection() gotNumBookings = %v, want %v", gotNumBookings, numBookings)
				}
			}
		})
	}
}

func TestBookingServer_RemoveUserFromTrain(t *testing.T) {
	ctx := context.Background()

	// Create a test server and get the client
	db := datastore.NewDatastore(
		datastore.WithSections("A", "B"),
		datastore.WithSectionSize(2))
	client, closer := createTestServer(t, ctx, db)
	defer closer()

	tests := map[string]struct {
		subject      string
		isAdmin      bool
		user         *pb.User
		seat         *pb.Seat
		newSeatId    string
		NewSectionId string
		wantErr      bool
	}{
		"admin able to remove user from the train": {
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
		"non-admin user cannot remove user from the train": {
			subject: "user@example.com",
			isAdmin: false,
			user: &pb.User{
				EmailAddress: "user@example.com",
				FirstName:    "john",
				LastName:     "doe",
			},
			seat: &pb.Seat{
				SectionId: "A",
				SeatId:    "2",
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := getCtxWithToken(t, ctx, tt.subject, tt.isAdmin)
			booking, err := client.Purchase(ctx, &pb.PurchaseRequest{
				User: tt.user,
				Seat: tt.seat,
			})
			if err != nil {
				t.Fatalf("Purchase() error = %v, wantErr %v", err, tt.wantErr)
			}

			bookingID := booking.BookingId

			// Remove the user
			_, err = client.RemoveUserFromTrain(ctx, &pb.RemoveBookingRequest{
				BookingId: bookingID,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("Purchase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				// Get all the seats from that section and check if the seat still available
				section := booking.Seat.SectionId

				stream, err := client.GetBookingsBySection(ctx, &pb.GetBookingsBySectionRequest{
					Section: section,
				})
				if err != nil {
					t.Fatalf("unable to get stream for GetBookingsBySection: %v", err)
				}

				// Check if the booking is still in the section
				for {
					booking, err := stream.Recv()
					if err == io.EOF {
						break
					} else if err != nil {
						break
					}
					if booking.BookingId == bookingID {
						t.Errorf("RemoveUserFromTrain() booking still exists in section: %v", section)
					}
				}
			}
		})
	}
}

func TestBookingServer_ModifySeat(t *testing.T) {
	ctx := context.Background()

	// Create a test server and get the client
	db := datastore.NewDatastore(
		datastore.WithSections("A", "B"),
		datastore.WithSectionSize(2))
	client, closer := createTestServer(t, ctx, db)
	defer closer()

	tests := map[string]struct {
		subject      string
		isAdmin      bool
		user         *pb.User
		seat         *pb.Seat
		newSeatId    string
		NewSectionId string
		wantErr      bool
	}{
		"admin able to modify seat": {
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
			newSeatId:    "2",
			NewSectionId: "B",
			wantErr:      false,
		},
		"non-admin user cannot modify seat": {
			subject: "user@example.com",
			isAdmin: false,
			user: &pb.User{
				EmailAddress: "user@example.com",
				FirstName:    "john",
				LastName:     "doe",
			},
			seat: &pb.Seat{
				SectionId: "B",
				SeatId:    "1",
			},
			newSeatId:    "2",
			NewSectionId: "B",
			wantErr:      true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := getCtxWithToken(t, ctx, tt.subject, tt.isAdmin)
			booking, err := client.Purchase(ctx, &pb.PurchaseRequest{
				User: tt.user,
				Seat: tt.seat,
			})
			if err != nil {
				t.Fatalf("Purchase() error = %v, wantErr %v", err, tt.wantErr)
			}

			updatedBooking, err := client.ModifySeat(ctx, &pb.ModifySeatRequest{
				BookingId:    booking.BookingId,
				NewSeatId:    tt.newSeatId,
				NewSectionId: tt.NewSectionId,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("Purchase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if updatedBooking.Seat.SeatId != tt.newSeatId {
					t.Errorf("ModifySeat() updated seat id = %v, want %v", updatedBooking.Seat.SeatId, tt.newSeatId)
				}
				if updatedBooking.Seat.SectionId != tt.NewSectionId {
					t.Errorf("ModifySeat() updated section id = %v, want %v", updatedBooking.Seat.SectionId, tt.NewSectionId)
				}
			}
		})
	}
}
