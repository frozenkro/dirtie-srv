package provision

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/joho/godotenv"
)

// Live DB integration testing
// Run these locally but use docker container for db

func TestInsertDeviceLiveDb(t *testing.T) {
	// load local env vars so mongo client can access them
	godotenv.Load("../../.env")

	client := Connect()

	testMacAddr := "testMacAddr"
	oid, err := InsertDevice(&Device{MacAddress: testMacAddr})

	want := regexp.MustCompile("/^[a-f\\d]{24}$/i")

	if err != nil {
		t.Fatalf(err.Error())
	}
	if !want.MatchString(oid) {
		t.Fatalf("Object ID: '%s' does not match expected format", oid)
	}

	fmt.Printf("Inserted device with Object ID: '%s'\n", oid)
	fmt.Printf("Mac Address: '%s'\n", testMacAddr)

	Disconnect(client)
}

func TestGetByMacAddress(t *testing.T) {
	// load local env vars so mongo client can access them
	godotenv.Load("../../.env")

	client := Connect()

	testMacAddr := "testMacAddr"
	device, err := GetByMacAddress(testMacAddr)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if device == nil {
		t.Fatalf("Unable to find device with Mad Address: '%s'", testMacAddr)
	}

	if device.MacAddress != testMacAddr {
		t.Fatalf("Retrieved device with incorrect Mac Address '%s', expected '%s'", device.MacAddress, testMacAddr)
	}

	fmt.Printf("Retrieved device with Object ID: '%s'\n", device.Id)
	fmt.Printf("Mac Address '%s'", device.MacAddress)

	Disconnect(client)
}
