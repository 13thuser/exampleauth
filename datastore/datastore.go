package datastore

import (
	"crypto/rand"
	"fmt"
	"sync"
)

// See Datastore notes below
const (
	SECTION_SIZE = 10
	SECTION_A    = "A"
	SECTION_B    = "B"
)

type BookingNotFound error
type BookingAlreadyExits error
type SectionIsFull error
type SectionNotFound error
type SeatNotAvailable error

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

// Current implementation of the Datastore is narrow in scope and only supports the following operations:
// - Purchase: Adds a new booking to the datastore
// - GetBookingsBySection: Returns the bookings for a given section
// - RemoveUserFromTrain: Removes a user's booking from the datastore
// - ModifySeat: Updates the seat allocation for a given section and seat
// Allows only one trian bookings with hard coded section size of 10 seats and 2 sections
type Datastore struct {
	sync.RWMutex

	// You can also map the user to the booking id to track all the users

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

// allocationSeating updates the seat allocation for a given section and seat
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

	// check if seat is already allocated
	if _, ok := ds.seatAllocation[sectionID][seatID]; ok {
		return SeatNotAvailable(fmt.Errorf("seat already allocated: %v", seatID))
	}

	ds.seatAllocation[sectionID][seatID] = bookingID
	return nil
}

// Purchase adds a new booking to the datastore
func (ds *Datastore) Purchase(booking Booking) (Booking, error) {
	// Concurrency support
	ds.Lock()
	defer ds.Unlock()

	if booking.BookingID != "" {
		return Booking{}, fmt.Errorf("booking id must be empty: %v", booking.BookingID)
	}

	return ds.createBooking(booking)
}

// Internal purchase function
func (ds *Datastore) createBooking(booking Booking) (Booking, error) {
	bookingID := BookingID(booking.BookingID)
	if bookingID == "" {
		// create a new booking id
		id, err := createRandomID()
		if err != nil {
			return Booking{}, fmt.Errorf("failed to generate booking id: %v", err)
		}
		booking.BookingID = id
		bookingID = BookingID(id)
	}
	if err := ds.allocationSeating(SectionID(booking.Seat.SectionID), SeatID(booking.Seat.SeatID), bookingID); err != nil {
		return Booking{}, fmt.Errorf("failed to allocate seating: %v", err)
	}

	ds.bookings[bookingID] = booking
	return booking, nil
}

// GetBookingsBySection returns the bookings for a given section
func (ds *Datastore) GetBookingsBySection(sectionID SectionID) []Booking {
	// Concurrency support
	ds.RLock()
	defer ds.RUnlock()

	return ds.getBookingsBySection(sectionID)
}

// Internal get bookings by section function
func (ds *Datastore) getBookingsBySection(sectionID SectionID) []Booking {
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
func (ds *Datastore) RemoveUserFromTrain(bookingID BookingID) error {
	// Concurrency support
	ds.Lock()
	defer ds.Unlock()

	return ds.removeUserFromTrain(bookingID)
}

// Internal remove user from train function
func (ds *Datastore) removeUserFromTrain(bookingID BookingID) error {
	// Check if booking exists
	booking, ok := ds.bookings[bookingID]
	if !ok {
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
func (ds *Datastore) ModifySeat(bookingID BookingID, sectionID SectionID, seatID SeatID) (Booking, error) {
	// Concurrency support
	ds.Lock()
	defer ds.Unlock()

	return ds.modifySeat(bookingID, sectionID, seatID)
}

// Internal modify seat function
func (ds *Datastore) modifySeat(bookingID BookingID, sectionID SectionID, seatID SeatID) (Booking, error) {
	// Check if booking exists
	booking, ok := ds.bookings[bookingID]
	if !ok {
		return Booking{}, fmt.Errorf("booking not found: %v", bookingID)
	}

	// if _, ok := ds.seatAllocation[sectionID]; !ok {
	// 	ds.seatAllocation[sectionID] = make(Seating)
	// }

	// Remove the existing seat allocation
	if err := ds.removeUserFromTrain(bookingID); err != nil {
		return Booking{}, err
	}

	// Update the seat
	booking.Seat = Seat{
		SectionID: string(sectionID),
		SeatID:    string(seatID),
	}

	updatedBooking, err := ds.createBooking(booking)
	if err != nil {
		return Booking{}, err
	}
	return updatedBooking, nil
}
