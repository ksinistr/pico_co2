package device

import (
	"errors"
	"pico_co2/internal/app"
	"pico_co2/internal/config"
	"pico_co2/internal/display"
	"pico_co2/internal/hardware/rtc"
	"pico_co2/internal/hardware/sensors"
	"testing"
	"time"

	"tinygo.org/x/drivers"
)

type fakeSPI struct{}

func (fakeSPI) Tx([]byte, []byte) error           { return nil }
func (fakeSPI) Transfer(value byte) (byte, error) { return value, nil }

type fakeBus struct{}

func (fakeBus) Tx(uint16, []byte, []byte) error { return nil }

type fakeClock struct{}

func (fakeClock) ReadTime() (time.Time, error) {
	return time.Date(2026, 10, 28, 14, 23, 0, 0, time.UTC), nil
}
func (fakeClock) SetTime(time.Time) error { return nil }

type fakeClimate struct{ samples int }

func (d *fakeClimate) Read() error                   { d.samples++; return nil }
func (*fakeClimate) Celsius() float32                { return 24.5 }
func (*fakeClimate) RelHumidity() float32            { return 55 }
func (*fakeClimate) DataReady() (bool, error)        { return true, nil }
func (d *fakeClimate) ReadCO2() (int32, error)       { d.samples++; return 800, nil }
func (*fakeClimate) ReadTemperature() (int32, error) { return 27000, nil }
func (*fakeClimate) ReadHumidity() (int32, error)    { return 60, nil }

type fakeAir struct{ samples, compensations int }

func (d *fakeAir) Update(drivers.Measurement) error   { d.samples++; return nil }
func (*fakeAir) Validity() uint8                      { return 0 }
func (*fakeAir) ECO2() uint16                         { return 900 }
func (d *fakeAir) SetEnvDataMilli(int32, int32) error { d.compensations++; return nil }

type fakeHardware struct {
	spiConfig                                      config.SPI
	spiCalls                                       int
	displayI2C, displaySPI                         bool
	stage                                          string
	busCalls, sensorCalls, clockCalls, buttonCalls int
	aht20                                          fakeClimate
	scd40                                          fakeClimate
	ens160                                         fakeAir
}

func (h *fakeHardware) NewI2C(config.I2C) (drivers.I2C, error) {
	h.busCalls++
	if h.stage == "bus" {
		return nil, errors.New("bus failed")
	}
	return fakeBus{}, nil
}

func (h *fakeHardware) NewSPI(cfg config.SPI) (drivers.SPI, error) {
	h.spiCalls++
	h.spiConfig = cfg
	if h.stage == "spi" {
		return nil, errors.New("spi failed")
	}

	return fakeSPI{}, nil
}

func TestDisplayTransport(t *testing.T) {
	for _, tt := range []struct {
		name               string
		spi, fail, invalid bool
	}{
		{name: "I2C"},
		{name: "SPI", spi: true},
		{name: "SPI initialization error", spi: true, fail: true},
		{name: "SPI pin conflict before hardware access", spi: true, invalid: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := thermometerConfig()
			if tt.spi {
				cfg.SPI = []config.SPI{{Controller: config.SPI0, SCK: 18, MOSI: 19, MISO: 16, Frequency: 4000000}}
				cfg.Display.I2C = nil
				cfg.Display.SPI = &config.SPIDisplay{Bus: config.SPI0, CS: 17, DC: 20, Reset: 21}
			}
			if tt.invalid {
				cfg.Display.SPI.CS = 18
			}
			hardware := &fakeHardware{}
			if tt.fail {
				hardware.stage = "spi"
			}
			deps, err := Build(cfg, hardware, testPeripherals(hardware))
			if (err != nil) != (tt.fail || tt.invalid) {
				t.Fatalf("error = %v", err)
			}
			if tt.invalid {
				if hardware.busCalls != 0 || hardware.spiCalls != 0 {
					t.Fatal("invalid configuration touched hardware")
				}
				return
			}
			if tt.spi && (hardware.spiCalls != 1 || hardware.spiConfig != cfg.SPI[0]) {
				t.Fatal("SPI settings were not passed to the platform")
			}
			if tt.fail {
				if deps.Display != nil || hardware.sensorCalls != 0 {
					t.Fatal("initialization continued after SPI failure")
				}
				return
			}
			if hardware.displaySPI != tt.spi || hardware.displayI2C == tt.spi {
				t.Fatal("wrong display transport")
			}
			if !tt.spi && hardware.spiCalls != 0 {
				t.Fatal("initialized absent SPI bus")
			}
			if hardware.busCalls != 1 || hardware.sensorCalls != 1 {
				t.Fatal("I2C sensor initialization changed")
			}
		})
	}
}

func (h *fakeHardware) NewDisplay(cfg config.Display, ports DisplayPorts) (drivers.Displayer, error) {
	h.displayI2C, h.displaySPI = ports.I2C != nil, ports.SPI != nil
	if (cfg.I2C != nil) != h.displayI2C || (cfg.SPI != nil) != h.displaySPI {
		return nil, errors.New("wrong display transport")
	}
	if h.stage == "display" {
		return nil, errors.New("display failed")
	}
	return display.NewVirtualDisplay(128, 32), nil
}

func (h *fakeHardware) NewSensor(cfg config.Sensor, _ drivers.I2C, compensation app.Sensors, schedule sensors.Schedule) (app.Sensors, error) {
	h.sensorCalls++
	if h.stage == "sensor" {
		return nil, errors.New("sensor failed")
	}
	switch cfg.Model {
	case config.AHT20:
		return sensors.NewAHT20Source(&h.aht20, schedule), nil
	case config.SCD40:
		return sensors.NewSCD4xSource(&h.scd40, schedule), nil
	case config.ENS160:
		return sensors.NewENS160Source(&h.ens160, compensation, schedule), nil
	default:
		return nil, errors.New("unexpected model")
	}
}

func (h *fakeHardware) NewRTC(config.RTC, drivers.I2C) (rtc.RTC, error) {
	h.clockCalls++
	if h.stage == "clock" {
		return nil, errors.New("clock failed")
	}
	return fakeClock{}, nil
}

func (h *fakeHardware) NewButton(config.GPIO) (func() bool, error) {
	h.buttonCalls++
	if h.stage == "left" || (h.stage == "right" && h.buttonCalls == 2) {
		return nil, errors.New("button failed")
	}
	return func() bool { return false }, nil
}

func thermometerConfig() config.Config {
	aht20 := &config.Sensor{Model: config.AHT20}
	return config.Config{
		I2C:         []config.I2C{{Controller: config.I2C0, SDA: 4, SCL: 5}},
		Display:     config.Display{Model: config.SSD1306, I2C: &config.I2CDisplay{Bus: config.I2C0}, Width: 128, Height: 32},
		Temperature: aht20, Humidity: aht20,
	}
}

func TestBuildDirectSources(t *testing.T) {
	for _, tt := range []struct {
		name                    string
		model                   config.SensorModel
		split, compensationOnly bool
		sensorCount             int
		temperature, co2        float64
	}{
		{"thermometer", config.AHT20, false, false, 1, 24.5, 0},
		{"SCD40 all readings", config.SCD40, false, false, 1, 27, 800},
		{"different climate and CO2 sources", config.SCD40, true, false, 2, 24.5, 800},
		{"ENS160 compensation", config.ENS160, true, false, 2, 24.5, 900},
		{"CO2 only with compensation", config.ENS160, true, true, 2, 0, 900},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := thermometerConfig()
			if tt.model == config.SCD40 && !tt.split {
				cfg.Temperature.Model = config.SCD40
				cfg.CO2 = cfg.Temperature
			}
			if tt.split {
				cfg.CO2 = &config.Sensor{Model: tt.model}
				if tt.model == config.ENS160 {
					cfg.CO2.Compensation = cfg.Temperature
				}
			}
			if tt.compensationOnly {
				cfg.Temperature, cfg.Humidity = nil, nil
			}
			hardware := &fakeHardware{}
			deps, err := Build(cfg, hardware, testPeripherals(hardware))
			if err != nil {
				t.Fatal(err)
			}
			now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
			readings := deps.Sensors.Read(now)
			deps.Sensors.Read(now.Add(50 * time.Millisecond))
			if readings.Temperature.Value != tt.temperature || readings.CO2.Value != tt.co2 {
				t.Fatalf("readings = %+v", readings)
			}
			if hardware.sensorCalls != tt.sensorCount || hardware.busCalls != 1 {
				t.Fatal("repeated hardware initialization")
			}
			if hardware.aht20.samples > 1 || hardware.scd40.samples > 1 || hardware.ens160.samples > 1 {
				t.Fatal("repeated physical measurement")
			}
			if tt.model == config.ENS160 && hardware.ens160.compensations != 1 {
				t.Fatal("compensation not connected")
			}
			if deps.Clock != nil || deps.SetTime != nil || deps.LeftPressed != nil || deps.RightPressed != nil {
				t.Fatal("created absent controls")
			}
			if len(deps.Screens) != len(display.ActiveScreens()) || deps.Screens[0].ID != "glance-left-mono9" {
				t.Fatal("device changed the standard screen list")
			}
			monitor := app.New(deps, 8)
			monitor.Tick(now)
		})
	}
}

func TestBuildErrorsAndControls(t *testing.T) {
	for _, tt := range []struct {
		name, stage string
		invalid     bool
	}{
		{name: "clock and buttons"},
		{name: "invalid config before hardware", invalid: true},
		{name: "bus failure", stage: "bus"},
		{name: "display failure", stage: "display"},
		{name: "sensor failure", stage: "sensor"},
		{name: "clock failure", stage: "clock"},
		{name: "left button failure", stage: "left"},
		{name: "right button failure", stage: "right"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := thermometerConfig()
			cfg.RTC = &config.RTC{Model: config.DS3231}
			cfg.Buttons = &config.Buttons{Left: 10, Right: 11}
			if tt.invalid {
				cfg.Humidity = &config.Sensor{Model: config.AHT20}
			}
			hardware := &fakeHardware{stage: tt.stage}
			deps, err := Build(cfg, hardware, testPeripherals(hardware))
			if (err != nil) != (tt.invalid || tt.stage != "") {
				t.Fatalf("error = %v", err)
			}
			if tt.invalid && hardware.busCalls != 0 {
				t.Fatal("invalid config touched hardware")
			}
			if err == nil && (deps.Clock == nil || deps.SetTime == nil || deps.LeftPressed == nil || deps.RightPressed == nil || hardware.buttonCalls != 2) {
				t.Fatal("controls not connected")
			}
		})
	}
}

func testPeripherals(hardware *fakeHardware) Peripherals {
	return Peripherals{
		NewDisplay: hardware.NewDisplay,
		NewSensor:  hardware.NewSensor,
		NewRTC:     hardware.NewRTC,
	}
}

func TestRequiredConstructors(t *testing.T) {
	for _, tt := range []struct {
		name                                          string
		omitDisplay, omitSensor, omitRTC, clock, fail bool
	}{
		{name: "missing display", omitDisplay: true, fail: true},
		{name: "missing sensor", omitSensor: true, fail: true},
		{name: "missing configured RTC", omitRTC: true, clock: true, fail: true},
		{name: "absent RTC needs no constructor", omitRTC: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := thermometerConfig()
			if tt.clock {
				cfg.RTC = &config.RTC{Model: config.DS3231}
			}
			hardware := &fakeHardware{}
			constructors := testPeripherals(hardware)
			if tt.omitDisplay {
				constructors.NewDisplay = nil
			}
			if tt.omitSensor {
				constructors.NewSensor = nil
			}
			if tt.omitRTC {
				constructors.NewRTC = nil
			}
			_, err := Build(cfg, hardware, constructors)
			if (err != nil) != tt.fail {
				t.Fatalf("error = %v", err)
			}
			if tt.fail && hardware.busCalls != 0 {
				t.Fatal("missing constructor touched hardware")
			}
		})
	}
}
