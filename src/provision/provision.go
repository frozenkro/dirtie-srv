package provision

import (
	"errors"

	"github.com/beevik/guid"
)

// Returns ID of provisioned device record from mongo
func ConnectDevice(macAddr string) (int, error) {

	// Check mongodb for mac address

	// If missing, insert into devices table with new guid for ID

	// Else update last communication time

	// Return provisioned device ID
	return 0, nil
}

func newGuid() *guid.Guid {
	g := guid.New()
	return g
}

func throw(msg string) error {
	return errors.New(msg)
}
