// Package max44009 communicates with a MAX44009 Ambient Light Sensor over I²C.
//
// A MAX44009 light sensor provides the ability to detect luminosity up to 188000 Lux.
//
// MAX44009 Datasheet: https://datasheets.maximintegrated.com/en/ds/MAX44009.pdf
package max44009

import (
	"log"
	"math"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/host"
	"sync"
)

const (
	interruptStatusRegister = 0x00
	interruptEnableRegister = 0x01
	configurationRegister   = 0x02
	luxHighByteRegister     = 0x03
	luxLowByteRegister      = 0x04
	upperThresholdRegister  = 0x05
	lowerThresholdRegister  = 0x06
	thresholdTimerRegister  = 0x07
)

// MAX44009 is a container for an i2c.Dev.
type MAX44009 struct {
	busCloser i2c.BusCloser
	Dev       i2c.Dev

	mu   sync.Mutex
	wg   sync.WaitGroup
	stop chan struct{}
}

// New returns a new MAX44009 that communicates over I²C to
// a MAX44009 light sensor.
func New(address uint16, bus string) *MAX44009 {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	b, err := i2creg.Open(bus)
	if err != nil {
		log.Fatal(err)
	}

	return &MAX44009{
		Dev: i2c.Dev{
			Bus:  b,
			Addr: address,
		},
		busCloser: b,
	}
}

// Close closes the I²C Bus being used by the MAX44009.
func (d *MAX44009) Close() error {
	return d.busCloser.Close()
}

// Configure is used to set continuous mode, manual mode,
// whether the current is divided, and the integration time.
//
// For more information on what each setting does, view pages 8-9
// in the datasheet: https://datasheets.maximintegrated.com/en/ds/MAX44009.pdf
func (d *MAX44009) Configure(continuous bool, manual bool, cdr bool, time byte) (error, byte) {
	var config byte
	if continuous {
		config |= 1 << 7
	}
	if manual {
		config |= 1 << 6
	}
	if cdr {
		config |= 1 << 3
	}
	config |= byte(time & 0x07)

	ret := make([]byte, 1)
	return d.Dev.Tx([]byte{configurationRegister, config}, ret), ret[0]
}

// Reads luminosity in Lux from the sensor. Returns a float from 0 - 188,000.
func (d *MAX44009) ReadLuminosityOnce() (error, float64) {
	bytes := make([]byte, 2)
	tx := d.Dev.Tx([]byte{luxHighByteRegister}, bytes)

	exponent := (bytes[0] & 0xF0) >> 4
	mantissa := ((bytes[0] & 0x0F) << 4) | (bytes[1] & 0x0F)

	return tx, math.Pow(2, float64(exponent)) * float64(mantissa) * .72
}

// ReadLuminosityContinuously reads luminosity in Lux on a continuous basis.
//
// HaltLuminosityReading() must be called to stop the sensing and close the channel.
func (d *MAX44009) ReadLuminosityContinuously() <-chan float64 {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.stop != nil {
		close(d.stop)
		d.stop = nil
		d.wg.Wait()
	}

	data := make(chan float64)
	d.stop = make(chan struct{})
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		defer close(data)
		d.readingLuminosity(data, d.stop)
	}()

	return data
}

func (d *MAX44009) readingLuminosity(data chan<- float64, done <-chan struct{}) {
	for {
		d.mu.Lock()
		err, lux := d.ReadLuminosityOnce()
		d.mu.Unlock()

		if err != nil {
			log.Printf("Sensor failed to sense: %v", err)
			return
		}

		select {
		case data <- lux:
		case <-done:
			return
		}
	}
}

// Halt stops the MAX44009 from reading luminosity and closes the channel after
// ReadLuminosityContinuously() is called.
func (d *MAX44009) HaltLuminosityReading() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.stop == nil {
		return
	}

	close(d.stop)
	d.stop = nil
	d.wg.Wait()
}
