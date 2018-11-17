package max44009_test

import (
	"fmt"
	"github.com/greg-bahr/max44009"
	"log"
	"time"
)

func Example() {
	// Create new MAX44009 at address 0x4a on I²C bus 1
	sensor := max44009.New(0x4a, "1")
	defer sensor.Close()

	// Configure the sensor to run on continuous mode, manual mode,
	// and set integration time to 100ms.
	sensor.Configure(true, true, false, 3)

	// Set up a ticker to run every 100ms.
	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for range ticker.C {
			// Retrieve and print luminosity from sensor.
			err, lux := sensor.Luminosity()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(lux)
		}
	}()

	// Stop the ticker after 60 seconds.
	time.Sleep(time.Second * 60)
	ticker.Stop()
}
