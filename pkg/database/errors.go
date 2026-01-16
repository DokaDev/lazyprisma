package database

import "errors"

var (
	// ErrNotConnected is returned when trying to use a driver that hasn't been connected
	ErrNotConnected = errors.New("database: not connected")

	// ErrDriverNotFound is returned when a driver is not registered
	ErrDriverNotFound = errors.New("database: driver not found")

	// ErrDriverAlreadyRegistered is returned when trying to register a driver with an existing name
	ErrDriverAlreadyRegistered = errors.New("database: driver already registered")
)
