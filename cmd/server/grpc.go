package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/13thuser/exampleauth/datastore"
	pb "github.com/13thuser/exampleauth/grpc"
)

// Define your gRPC service interface
type BookingServer struct {
	pb.UnimplementedBookingServiceServer

	// Datastore
	db *datastore.Datastore
}

// NewBookingServer creates a new instance of the BookingServer
func NewBookingServer(db *datastore.Datastore) *BookingServer {
	return &BookingServer{
		db: db,
	}
}

// isUserAuthenticated checks if the user is authenticated and returns the emailID
func (s *BookingServer) isUserAuthenticated(ctx context.Context) (string, bool) {
	emailID, ok := ctx.Value(emailIDKey).(string)
	if !ok {
		return "", false
	}
	return emailID, true
}

// isAdmin checks if the user is an admin
func (s *BookingServer) isAdmin(ctx context.Context) bool {
	isAdmin, ok := ctx.Value(isAdminKey).(bool)
	if !ok {
		return false
	}
	return isAdmin
}

// Implement the gRPC service methods
func (s *BookingServer) Purchase(ctx context.Context, req *pb.PurchaseRequest) (*pb.Booking, error) {
	fmt.Printf("Received: %v\n", req)

	// Make sure user is authenticated. non-admin can work too.
	emailID, authenticated := s.isUserAuthenticated(ctx)
	if !authenticated {
		return nil, status.Errorf(codes.Unauthenticated, "user is not authenticated")
	}

	booking := datastore.Booking{
		User: datastore.User{
			EmailAddress: req.User.EmailAddress,
			FirstName:    req.User.FirstName,
			LastName:     req.User.LastName,
		},
		Seat: datastore.Seat{
			SectionID: req.Seat.SectionId,
			SeatID:    req.Seat.SeatId,
		},
		From:      "London", // From and to is fixed and hardcoded for now
		To:        "Paris",
		PricePaid: 20.00, // Currency field is eliminated because of timing constraints
	}

	booking, err := s.db.Purchase(emailID, booking)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to purchase: %v", err)
	}

	pbBooking := &pb.Booking{
		BookingId: booking.BookingID,
		User: &pb.User{
			EmailAddress: booking.User.EmailAddress,
			FirstName:    booking.User.FirstName,
			LastName:     booking.User.LastName,
		},
		Seat: &pb.Seat{
			SectionId: booking.Seat.SectionID,
			SeatId:    booking.Seat.SeatID,
		},
		From:      booking.From,
		To:        booking.To,
		PricePaid: booking.PricePaid,
	}

	return pbBooking, nil
}

func (s *BookingServer) GetBookingsBySection(req *pb.GetBookingsBySectionRequest, stream pb.BookingService_GetBookingsBySectionServer) error {
	ctx := stream.Context()

	// Check if the user is an admin
	if !s.isAdmin(ctx) {
		return status.Errorf(codes.PermissionDenied, "user is not an admin")
	}

	// Stream the bookings response
	for _, booking := range s.db.GetBookingsBySection(datastore.SectionID(req.Section)) {
		err := stream.Send(&pb.Booking{
			BookingId: booking.BookingID,
			User: &pb.User{
				EmailAddress: booking.User.EmailAddress,
				FirstName:    booking.User.FirstName,
				LastName:     booking.User.LastName,
			},
			Seat: &pb.Seat{
				SectionId: booking.Seat.SectionID,
				SeatId:    booking.Seat.SeatID,
			},
			From:      booking.From,
			To:        booking.To,
			PricePaid: booking.PricePaid,
		})
		if err != nil {
			return status.Errorf(codes.Unknown, "failed to stream booking: %v", err)
		}
	}

	return nil
}

func (s *BookingServer) RemoveUserFromTrain(ctx context.Context, req *pb.User) (*emptypb.Empty, error) {
	// Your logic here
	return nil, status.Errorf(codes.Unimplemented, "method RemoveUser not implemented")
}

func (s *BookingServer) ModifySeat(ctx context.Context, req *pb.ModifySeatRequest) (*pb.Booking, error) {
	// Your logic here
	return nil, status.Errorf(codes.Unimplemented, "method ModifySeat not implemented")
}