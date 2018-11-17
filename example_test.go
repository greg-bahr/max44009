package max44009_test

import (
	"fmt"
	"github.com/greg-bahr/max44009"
	"time"
)

func Example() {
	// Create new MAX44009 at address 0x4a on I²C bus 1
	sensor := max44009.New(0x4a, "1")
	defer sensor.Close()

	// Configure the sensor to run on continuous mode, manual mode,
	// and set integration time to 100ms.
	sensor.Configure(true, false, false, 2)

	go func() {
		// Continuously read and print luminosity
		for lux := range sensor.ReadLuminosityContinuously() {
			fmt.Printf("Lux: %.2f\n", lux)
		}
	}()

	// Stop the ticker after 60 seconds.
	time.Sleep(time.Second * 60)
	sensor.HaltLuminosityReading()
}
