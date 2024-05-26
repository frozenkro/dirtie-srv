package provision

import (
	"errors"

	"github.com/beevik/guid"
)

// Returns ID of provisioned device record from mongo
func ConnectDevice(macAddr string) (string, error) {

	// Check mongodb for mac address
	device, err := GetByMacAddress(macAddr)
	if err != nil {
		return "", err
	}

	var oid string
	// If missing, insert into devices table with new guid for ID
	if device == nil {
		oid, err = InsertDevice(&Device{
			MacAddress: macAddr,
		})
		if err != nil {
			return "", err
		}
	} else {
		// Else update last communication time
		oid = device.Id
	}

	// Return provisioned device ID
	return oid, nil
}

func newGuid() *guid.Guid {
	g := guid.New()
	return g
}

func throw(msg string) error {
	return errors.New(msg)
}
