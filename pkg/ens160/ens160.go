// Package ens160 provides a driver for the ScioSense ENS160 digital gas sensor.
//
// Datasheet: https://www.sciosense.com/wp-content/uploads/2023/12/ENS160-Datasheet.pdf
package ens160

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"tinygo.org/x/drivers"
)

const (
	defaultTimeout = 20 * time.Millisecond
	shortTimeout   = 1 * time.Millisecond
	longTimeout    = 1 * time.Second
)

// Device wraps an I2C connection to an ENS160 device.
type Device struct {
	bus  drivers.I2C // I²C implementation
	addr uint16      // 7‑bit bus address, promoted to uint16 per drivers.I2C

	// shadow registers / last measurements
	lastTvocPPB  uint16
	lastEco2PPM  uint16
	lastAqiUBA   uint8
	lastValidity uint8 // Store the latest validity status

	// pre‑allocated buffers
	wbuf [6]byte // longest write: reg + 4 bytes (TEMP+RH)
	rbuf [5]byte // longest read: DATA burst (5 bytes)
}

// New returns a new ENS160 driver.
func New(bus drivers.I2C, addr uint16) *Device {
	if addr == 0 {
		addr = DefaultAddress
	}

	return &Device{bus: bus, addr: addr}
}

// Configure sets up the device for reading.
func (d *Device) Configure() error {
	// 1. Soft‑reset
	if err := d.write1(regOpMode, ModeReset); err != nil {
		return err
	}

	time.Sleep(defaultTimeout)

	// 2. Enter IDLE, clear GPR registers, then go STANDARD.
	if err := d.write1(regOpMode, ModeIdle); err != nil {
		return err
	}

	time.Sleep(defaultTimeout)

	if err := d.write1(regCommand, cmdClrGPR); err != nil {
		return err
	}

	time.Sleep(defaultTimeout)

	if err := d.write1(regOpMode, ModeStandard); err != nil {
		return err
	}

	time.Sleep(longTimeout)

	return nil
}

// SetEnvDataMilli sets the ambient temperature and humidity for compensation.
//
// TempMilliC is the temperature in milli-degrees Celsius.
// RhMilliPct is the relative humidity in milli-percent.
func (d *Device) SetEnvDataMilli(tempMilliC, rhMilliPct int32) error {
	// Clip temperature
	const (
		minC = -40 * 1000
		maxC = 85 * 1000
	)
	if tempMilliC < minC {
		tempMilliC = minC
	} else if tempMilliC > maxC {
		tempMilliC = maxC
	}

	// Clip humidity
	if rhMilliPct < 0 {
		rhMilliPct = 0
	} else if rhMilliPct > 100*1000 {
		rhMilliPct = 100 * 1000
	}

	// Integer fixed‑point conversion
	tempRaw := uint16(((tempMilliC + 273150) * 64) / 1000) // Kelvin×64
	humRaw := uint16((rhMilliPct * 512) / 1000)            // %RH×512

	d.wbuf[0] = regTempIn // start address (auto‑increment)
	binary.LittleEndian.PutUint16(d.wbuf[1:3], tempRaw)
	binary.LittleEndian.PutUint16(d.wbuf[3:5], humRaw)

	if err := d.bus.Tx(d.addr, d.wbuf[:5], nil); err != nil {
		return fmt.Errorf("ENS160: write environmental data failed: %w", err)
	}

	return nil
}

// Update refreshes the concentration measurements.
func (d *Device) Update(which drivers.Measurement) error {
	if which&drivers.Concentration == 0 {
		return nil // nothing requested
	}

	const maxTries = 1000

	var (
		status   uint8
		validity uint8
		gotData  bool
	)

	// Poll DEVICE_STATUS until NEWDAT or timeout

	for range maxTries {
		var err error

		status, err = d.read1(regStatus)
		if err != nil {
			return err
		}

		if status&statusSTATER != 0 {
			return errors.New("ENS160: error (STATER set)")
		}

		validity = (status & statusValidityMask) >> statusValidityShift

		if status&statusNEWDAT != 0 {
			gotData = true
			break // Always break when data available
		}

		time.Sleep(shortTimeout)
	}

	if !gotData {
		return errors.New("ENS160: timeout waiting for NEWDAT")
	}

	// Burst-read data regardless of validity state
	d.wbuf[0] = regAQI
	if err := d.bus.Tx(d.addr, d.wbuf[:1], d.rbuf[:5]); err != nil {
		return fmt.Errorf("ENS160: burst read failed: %w", err)
	}

	d.lastAqiUBA = d.rbuf[0]
	d.lastTvocPPB = binary.LittleEndian.Uint16(d.rbuf[1:3])
	d.lastEco2PPM = binary.LittleEndian.Uint16(d.rbuf[3:5])
	d.lastValidity = validity // Store the validity status

	return nil
}

// TVOC returns the last total‑VOC concentration in parts‑per‑billion.
func (d *Device) TVOC() uint16 { return d.lastTvocPPB }

// ECO2 returns the last equivalent CO₂ concentration in parts‑per‑million.
func (d *Device) ECO2() uint16 { return d.lastEco2PPM }

// AQI returns the last Air‑Quality Index according to UBA (1–5).
func (d *Device) AQI() uint8 { return d.lastAqiUBA }

// Validity returns the current operating state of the sensor.
func (d *Device) Validity() uint8 {
	return d.lastValidity
}

// Sleep puts the device into deep sleep mode to minimize power consumption
// and self-heating.
func (d *Device) Sleep() error {
	return d.write1(regOpMode, ModeDeepSleep)
}

// Wake sets the device to idle mode. From here you can set it to standard mode
// when ready to take measurements.
func (d *Device) Wake() error {
	if err := d.write1(regOpMode, ModeIdle); err != nil {
		return err
	}

	time.Sleep(defaultTimeout)

	return nil
}

// EnableMeasurements sets the device to standard measurement mode.
func (d *Device) EnableMeasurements() error {
	if err := d.write1(regOpMode, ModeStandard); err != nil {
		return err
	}

	time.Sleep(longTimeout)

	return nil
}

// write1 writes a single byte to a register.
func (d *Device) write1(reg, val uint8) error {
	d.wbuf[0] = reg
	d.wbuf[1] = val

	if err := d.bus.Tx(d.addr, d.wbuf[:2], nil); err != nil {
		return fmt.Errorf("ENS160: write register 0x%02x failed: %w", reg, err)
	}

	return nil
}

// read1 reads a single byte from a register.
func (d *Device) read1(reg uint8) (uint8, error) {
	d.wbuf[0] = reg
	if err := d.bus.Tx(d.addr, d.wbuf[:1], d.rbuf[:1]); err != nil {
		return 0, fmt.Errorf("ENS160: read register 0x%02x failed: %w", reg, err)
	}

	return d.rbuf[0], nil
}
