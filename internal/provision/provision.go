package provision

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
