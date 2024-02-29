package datastore

import (
	"crypto/rand"
	"fmt"
)

const (
	SECTION_SIZE = 10
	SECTION_A    = "A"
	SECTION_B    = "B"
)

type BookingNotFound error
type BookingAlreadyExits error
type SectionIsFull error
type SectionNotFound error

type User struct {
	EmailAddress string // main id of the user, also subject of the JWT token
	FirstName    string
	LastName     string
}

type Seat struct {
	SectionID string
	SeatID    string
}

type Booking struct {
	BookingID string
	User      User
	Seat      Seat
	From      string
	To        string
	PricePaid float64
}

type BookingID string
type SectionID string
type SeatID string
type Seating map[SeatID]BookingID

type Datastore struct {
	// map of booking ID to booking that contains the user and
	bookings map[BookingID]Booking

	// map of seat allocation to the user by section id and seat id
	seatAllocation map[SectionID]Seating
}

// NewDatastore creates a new instance of the Datastore
func NewDatastore() *Datastore {
	return &Datastore{
		bookings:       make(map[BookingID]Booking),
		seatAllocation: make(map[SectionID]Seating),
	}
}

// createRandomID generates a random booking id
func createRandomID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate booking id: %v", err)
	}
	return fmt.Sprintf("%x", b), nil
}

func (ds *Datastore) allocationSeating(sectionID SectionID, seatID SeatID, bookingID BookingID) error {
	if sectionID != SECTION_A && sectionID != SECTION_B {
		return fmt.Errorf("section id must be A or B: %v", sectionID)
	}

	if len(ds.seatAllocation[sectionID]) >= SECTION_SIZE {
		return SectionIsFull(fmt.Errorf("section is full: %v", sectionID))
	}

	if _, ok := ds.seatAllocation[sectionID]; !ok {
		ds.seatAllocation[sectionID] = make(Seating)
	}
	ds.seatAllocation[sectionID][seatID] = bookingID
	return nil
}

// Purchase adds a new booking to the datastore
func (ds *Datastore) Purchase(userID string, booking Booking) (Booking, error) {
	if booking.BookingID != "" {
		return Booking{}, fmt.Errorf("booking id must be empty: %v", booking.BookingID)
	}

	// create a new booking id
	id, err := createRandomID()
	if err != nil {
		return Booking{}, fmt.Errorf("failed to generate booking id: %v", err)
	}

	bookingID := BookingID(id)
	booking.BookingID = string(id)
	if err := ds.allocationSeating(SectionID(booking.Seat.SectionID), SeatID(booking.Seat.SeatID), bookingID); err != nil {
		return Booking{}, fmt.Errorf("failed to allocate seating: %v", err)
	}

	ds.bookings[bookingID] = booking
	return booking, nil
}

// GetBookingsBySection returns the bookings for a given section
func (ds *Datastore) GetBookingsBySection(sectionID SectionID) []Booking {
	// GetBookingsBySection returns the bookings for a given section
	var booking []Booking
	seating, ok := ds.seatAllocation[sectionID]
	if !ok {
		return booking
	}
	for _, bookingID := range seating {
		booking = append(booking, ds.bookings[bookingID])
	}
	return booking
}

// RemoveUserFromTrain removes a user's booking from the datastore
func (ds *Datastore) RemoveUserFromTrain(userID string, bookingID BookingID) error {
	// Find the booking and make sure user is the owner
	booking, ok := ds.bookings[bookingID]
	if !ok || booking.User.EmailAddress != userID {
		return BookingNotFound(fmt.Errorf("booking not found: %v", bookingID))
	}

	// Remove the allocation
	seat := SeatID(booking.Seat.SeatID)
	section := SectionID(booking.Seat.SectionID)

	// Get the seating for the section
	seating, ok := ds.seatAllocation[section]
	if !ok {
		return SectionNotFound(fmt.Errorf("section not found: %v", section))
	}

	// If booking exists then section and seat must exist
	// // check if seat is in the section
	// _, ok = seating[seat]
	// if !ok {
	// 	return fmt.Errorf("seat not found: %v", seat)
	// }

	// remove the seat
	delete(seating, seat)

	// delete the bookings
	delete(ds.bookings, bookingID)

	return nil
}

// ModifySeat updates the seat allocation for a given section and seat
func (ds *Datastore) ModifySeat(userID string, bookingID BookingID, sectionID SectionID, seatID SeatID) error {
	// Find the booking and make sure user is the owner
	booking, ok := ds.bookings[bookingID]
	if !ok || booking.User.EmailAddress != userID {
		return fmt.Errorf("booking not found: %v", bookingID)
	}

	if _, ok := ds.seatAllocation[sectionID]; !ok {
		ds.seatAllocation[sectionID] = make(Seating)
	}

	// Remove the existing seat allocation
	ds.RemoveUserFromTrain(userID, bookingID)

	// Allocate the new seat
	return ds.allocationSeating(sectionID, seatID, bookingID)
}
