package max44009

import (
	"log"
	"math"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/host"
)

const (
	InterruptStatusRegister = 0x00
	InterruptEnableRegister = 0x01
	ConfigurationRegister   = 0x02
	LuxHighByteRegister     = 0x03
	LuxLowByteRegister      = 0x04
	UpperThresholdRegister  = 0x05
	LowerThresholdRegister  = 0x06
	ThresholdTimerRegister  = 0x07
)

type MAX44009 struct {
	busCloser i2c.BusCloser
	Dev       i2c.Dev
}

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

func (d *MAX44009) Close() error {
	return d.busCloser.Close()
}

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
	return d.Dev.Tx([]byte{ConfigurationRegister, config}, ret), ret[0]
}

func (d *MAX44009) Luminosity() (error, float64) {
	bytes := make([]byte, 2)
	tx := d.Dev.Tx([]byte{LuxHighByteRegister}, bytes)

	exponent := (bytes[0] & 0xF0) >> 4
	mantissa := ((bytes[0] & 0x0F) << 4) | (bytes[1] & 0x0F)

	return tx, math.Pow(2, float64(exponent)) * float64(mantissa) * .72
}
