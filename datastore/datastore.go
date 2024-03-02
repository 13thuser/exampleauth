package datastore

import (
	"crypto/rand"
	"fmt"
	"strconv"
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
type InvalidSeatID error

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
	owner     string
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
type BookingsMap map[BookingID]struct{}

// Current implementation of the Datastore is narrow in scope and only supports the following operations:
// - Purchase: Adds a new booking to the datastore
// - GetBookingsBySection: Returns the bookings for a given section
// - RemoveUserFromTrain: Removes a user's booking from the datastore
// - ModifySeat: Updates the seat allocation for a given section and seat
// Allows only one trian bookings with hard coded section size of 10 seats and 2 sections
type Datastore struct {
	sync.RWMutex

	// You can also map the user to the booking id to track all the users
	userBookings map[string]BookingsMap

	// map of booking ID to booking that contains the user and
	bookings map[BookingID]Booking

	// map of seat allocation to the user by section id and seat id
	seatAllocation map[SectionID]Seating

	// map of sections
	sections map[SectionID]struct{}

	// section size
	sectionSize int
}

type DatastoreOption func(*Datastore)

// WithSectionSize sets the section size for the Datastore.
func WithSectionSize(size int) DatastoreOption {
	return func(ds *Datastore) {
		ds.sectionSize = size
	}
}

// WithSections sets the sections for the Datastore.
func WithSections(sections ...SectionID) DatastoreOption {
	return func(ds *Datastore) {
		ds.sections = make(map[SectionID]struct{})
		for _, section := range sections {
			ds.sections[section] = struct{}{}
		}
	}
}

// NewDatastore creates a new instance of the Datastore with the provided options.
func NewDatastore(options ...DatastoreOption) *Datastore {
	ds := &Datastore{
		userBookings:   make(map[string]BookingsMap),
		bookings:       make(map[BookingID]Booking),
		seatAllocation: make(map[SectionID]Seating),
		sections:       map[SectionID]struct{}{SECTION_A: {}, SECTION_B: {}},
		sectionSize:    SECTION_SIZE,
	}

	for _, option := range options {
		option(ds)
	}

	return ds
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
	// check if sectionID exists in sections
	if _, ok := ds.sections[sectionID]; !ok {
		return SectionNotFound(fmt.Errorf("section not found: %v", sectionID))
	}

	if len(ds.seatAllocation[sectionID]) >= ds.sectionSize {
		return SectionIsFull(fmt.Errorf("section is full: %v", sectionID))
	}

	// convert seatID to int and check if it is within the section size
	if seat, err := strconv.Atoi(string(seatID)); err != nil || seat < 0 || seat > ds.sectionSize {
		return InvalidSeatID(fmt.Errorf("invalid seat id: %v", seatID))
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
func (ds *Datastore) Purchase(userID string, booking Booking) (Booking, error) {
	// Concurrency support
	ds.Lock()
	defer ds.Unlock()

	if booking.BookingID != "" {
		return Booking{}, fmt.Errorf("booking id must be empty: %v", booking.BookingID)
	}

	return ds.createBooking(userID, booking)
}

func (ds *Datastore) GetUserBookings(userID string) []Booking {
	// Concurrency support
	ds.RLock()
	defer ds.RUnlock()

	return ds.getUserBookings(userID)
}

// Internal get user bookings function
func (ds *Datastore) getUserBookings(userID string) []Booking {
	var bookings []Booking
	for bookingID := range ds.userBookings[userID] {
		bookings = append(bookings, ds.bookings[bookingID])
	}
	return bookings
}

// Internal purchase function
func (ds *Datastore) createBooking(userID string, booking Booking) (Booking, error) {
	// Make sure use has bookings map
	if _, ok := ds.userBookings[userID]; !ok {
		ds.userBookings[userID] = make(BookingsMap)
	}
	bookingID := BookingID(booking.BookingID)
	if bookingID != "" {
		// check if not admin then check if the user is the owner of the booking
		if _, ok := ds.userBookings[userID][bookingID]; !ok {
			return Booking{}, BookingNotFound(fmt.Errorf("booking not found: %v", bookingID))
		}
	} else {
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

	booking.owner = userID
	ds.bookings[bookingID] = booking
	ds.userBookings[userID][bookingID] = struct{}{}
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

	// Remove the existing seat allocation
	if err := ds.removeUserFromTrain(bookingID); err != nil {
		return Booking{}, err
	}

	// Update the seat
	booking.Seat = Seat{
		SectionID: string(sectionID),
		SeatID:    string(seatID),
	}

	updatedBooking, err := ds.createBooking(booking.owner, booking)
	if err != nil {
		return Booking{}, err
	}
	return updatedBooking, nil
}
