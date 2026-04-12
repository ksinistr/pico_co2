package app

import (
	"cmp"
	"fmt"
	"machine"
	"pico_co2/internal/button"
	"pico_co2/internal/clockedit"
	"pico_co2/internal/display"
	"pico_co2/internal/rtc"
	"pico_co2/internal/types"
	"pico_co2/pkg/ens160"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/aht20"
	"tinygo.org/x/drivers/ds3231"
	"tinygo.org/x/drivers/scd4x"
	"tinygo.org/x/drivers/ssd1306"
)

type Config struct {
	Display struct {
		Width   int16
		Height  int16
		Address uint16
	}
	I2C struct {
		Frequency uint32
		SDA       machine.Pin
		SCL       machine.Pin
	}
	Buttons struct {
		Button1 machine.Pin
		Button2 machine.Pin
	}
	Timeouts struct {
		Startup time.Duration
		Minute  time.Duration
		Second  time.Duration
	}
	QueueCapacity       int
	DefaultDisplayIndex int
}

func DefaultConfig() Config {
	cfg := Config{}
	cfg.Display.Width = 128
	cfg.Display.Height = 32
	cfg.Display.Address = ssd1306.Address_128_32
	cfg.I2C.Frequency = 400 * machine.KHz
	cfg.I2C.SDA = machine.GP4
	cfg.I2C.SCL = machine.GP5
	cfg.Buttons.Button1 = machine.GP10
	cfg.Buttons.Button2 = machine.GP11
	cfg.Timeouts.Startup = 1 * time.Minute
	cfg.Timeouts.Minute = 1 * time.Minute
	cfg.Timeouts.Second = 1 * time.Second
	cfg.QueueCapacity = 480
	cfg.DefaultDisplayIndex = 0
	return cfg
}

func (c Config) initI2C() error {
	return machine.I2C0.Configure(machine.I2CConfig{
		Frequency: c.I2C.Frequency,
		SDA:       c.I2C.SDA,
		SCL:       c.I2C.SCL,
	})
}

func (c Config) initDisplay(bus drivers.I2C) (display.Renderer, error) {
	disp := ssd1306.NewI2C(bus)
	disp.Configure(ssd1306.Config{
		Width:   c.Display.Width,
		Height:  c.Display.Height,
		Address: c.Display.Address,
	})
	// REDUCE BRIGHTNESS
	// reduce contrast for night viewing
	disp.Command(ssd1306.SETCONTRAST)
	disp.Command(0x01)
	// precharge period
	disp.Command(ssd1306.SETPRECHARGE)
	disp.Command(
		0xE1,
	) // 0xF1 default, 0xE1 for lower power, 0xD2 for even lower
	// VCOMH deselect level
	disp.Command(ssd1306.SETVCOMDETECT)
	disp.Command(
		0x30,
	) // 0x20 default, 0x30 for lower power, 0x40 for even lower
	return display.NewSSD1306Adapter(&disp), nil
}

type RawReadings struct {
	CO2         uint16
	Temperature float32
	Humidity    float32
}

type Sensors struct {
	aht20  *aht20.Device
	ens160 *ens160.Device
	scd4x  *scd4x.Device
}

func NewSensors(bus drivers.I2C) (*Sensors, error) {
	s := &Sensors{}

	if err := s.initAHT20(bus); err != nil {
		return nil, fmt.Errorf("aht20 init: %w", err)
	}

	if err := s.initENS160(bus); err != nil {
		return nil, fmt.Errorf("ens160 init: %w", err)
	}

	if err := s.initSCD4x(bus); err != nil {
		return nil, fmt.Errorf("scd4x init: %w", err)
	}

	return s, nil
}

func (s *Sensors) initAHT20(bus drivers.I2C) error {
	aht20Sensor := aht20.New(bus)
	aht20Sensor.Reset()
	aht20Sensor.Configure()
	s.aht20 = &aht20Sensor
	return nil
}

func (s *Sensors) initENS160(bus drivers.I2C) error {
	ens160Sensor := ens160.New(bus, ens160.DefaultAddress)
	if err := ens160Sensor.Sleep(); err != nil {
		return err
	}
	s.ens160 = ens160Sensor
	return nil
}

func (s *Sensors) initSCD4x(bus drivers.I2C) error {
	scd4xSensor := scd4x.New(bus)
	time.Sleep(1500 * time.Millisecond)
	if err := scd4xSensor.Configure(); err != nil {
		return err
	}

	time.Sleep(1500 * time.Millisecond)

	if err := scd4xSensor.StartPeriodicMeasurement(); err != nil {
		return err
	}

	time.Sleep(1500 * time.Millisecond)

	s.scd4x = scd4xSensor
	return nil
}

func (s *Sensors) Read() (*RawReadings, error) {
	if err := s.aht20.Read(); err != nil {
		return nil, fmt.Errorf("aht20 read: %w", err)
	}

	co2, err := s.scd4x.ReadCO2()
	if err != nil {
		return nil, fmt.Errorf("scd4x read: %w", err)
	}

	return &RawReadings{
		CO2:         uint16(co2),
		Temperature: s.aht20.Celsius(),
		Humidity:    s.aht20.RelHumidity(),
	}, nil
}

type DisplayManager struct {
	renderer     display.Renderer
	currentIndex int
}

func NewDisplayManager(
	renderer display.Renderer,
	currentIndex int,
) *DisplayManager {
	return &DisplayManager{
		renderer:     renderer,
		currentIndex: currentIndex,
	}
}

func (dm *DisplayManager) NextDisplay() {
	dm.currentIndex = (dm.currentIndex + 1) % len(display.MethodRegistry)
}

func (dm *DisplayManager) PreviousDisplay() {
	dm.currentIndex = (dm.currentIndex - 1 + len(display.MethodRegistry)) % len(display.MethodRegistry)
}

func (dm *DisplayManager) Render(readings *types.Readings) {
	if readings.Error != "" {
		display.RenderError(dm.renderer, readings)
		return
	}

	if dm.currentIndex >= len(display.MethodRegistry) {
		dm.currentIndex = 0
	}

	renderMethod := display.MethodRegistry[dm.currentIndex]
	renderMethod.Fn(dm.renderer, readings)
}

type App struct {
	config         Config
	sensors        *Sensors
	displayManager *DisplayManager
	button1        *button.TouchButton
	button2        *button.TouchButton
	rtc            rtc.RTC
	editor         clockedit.EditorState
	lastLeftTime   time.Time
	lastRightTime  time.Time
}

func New(cfg Config) (*App, error) {
	err := cfg.initI2C()
	if err != nil {
		return nil, fmt.Errorf("i2c init: %w", err)
	}

	renderer, err := cfg.initDisplay(machine.I2C0)
	if err != nil {
		return nil, fmt.Errorf("display init: %w", err)
	}

	sensors, err := NewSensors(machine.I2C0)
	if err != nil {
		return nil, fmt.Errorf("sensors init: %w", err)
	}

	ds3231Sensor := ds3231.New(machine.I2C0)
	if ok := ds3231Sensor.Configure(); !ok {
		return nil, fmt.Errorf("failed to configure DS3231 sensor")
	}

	dt, _ := ds3231Sensor.ReadTime()

	if dt.Year() > 2200 || dt.Year() < 2024 {
		now := time.Date(2025, 12, 21, 14, 35, 0, 0, time.UTC)
		ds3231Sensor.SetTime(now)
		println("DS3231 time set to:", now.Format(time.DateTime))
	}

	running := ds3231Sensor.IsRunning()
	if !running {
		err := ds3231Sensor.SetRunning(true)
		if err != nil {
			return nil, fmt.Errorf("ds3231 set running: %w", err)
		}
	}

	return &App{
		config:         cfg,
		sensors:        sensors,
		displayManager: NewDisplayManager(renderer, cfg.DefaultDisplayIndex),
		button1:        button.NewTouchButton(cfg.Buttons.Button1),
		button2:        button.NewTouchButton(cfg.Buttons.Button2),
		rtc:            &ds3231Sensor,
	}, nil
}

func (a *App) Run() {
	readings := types.InitReadings(a.config.QueueCapacity)

	wd := machine.Watchdog
	wd.Configure(machine.WatchdogConfig{
		TimeoutMillis: machine.WatchdogMaxTimeout,
	})
	wd.Start()

	println("starting loop")

	for {
		wd.Update()

		a.handleInput(readings)
		a.updateReadings(readings)
		a.render(readings)

		time.Sleep(50 * time.Millisecond)
	}
}

func (a *App) handleInput(readings *types.Readings) {
	now := time.Now()
	leftPressed := a.button1.Consume()
	rightPressed := a.button2.Consume()

	if leftPressed {
		a.lastLeftTime = now
	}
	if rightPressed {
		a.lastRightTime = now
	}

	ev := clockedit.ButtonEvent{
		LeftPressed:  leftPressed,
		RightPressed: rightPressed,
		LeftTime:     a.lastLeftTime,
		RightTime:    a.lastRightTime,
	}

	onTimeScreen := a.displayManager.currentIndex == 0
	result := clockedit.ProcessInput(a.editor, ev, onTimeScreen, clockedit.DefaultSimultaneousWindow)

	if result.AttemptEntry {
		a.editor = clockedit.Enter(readings.Time.Hour, readings.Time.Minute)
		a.lastLeftTime = time.Time{}
		a.lastRightTime = time.Time{}
		readings.IsDrawen = false
		return
	}

	if result.Exited {
		if result.ShouldSave {
			a.handleSave(readings, result.State)
		} else {
			a.editor = clockedit.EditorState{}
		}
		a.lastLeftTime = time.Time{}
		a.lastRightTime = time.Time{}
		readings.IsDrawen = false
		return
	}

	a.editor = result.State
	if result.NavLeft {
		a.displayManager.PreviousDisplay()
		a.lastLeftTime = time.Time{}
	}
	if result.NavRight {
		a.displayManager.NextDisplay()
		a.lastRightTime = time.Time{}
	}
	if result.RedrawNeeded {
		readings.IsDrawen = false
	}
}

func (a *App) handleSave(readings *types.Readings, state clockedit.EditorState) {
	curTime, readErr := a.rtc.ReadTime()
	if readErr != nil {
		a.editor = clockedit.HandleSaveAttempt(state, readErr)
		readings.Error = fmt.Sprintf("RTC read: %v", readErr)
		return
	}

	newTime, saveErr := rtc.SaveTime(a.rtc, state.Hour, state.Minute, curTime)
	a.editor = clockedit.HandleSaveAttempt(state, saveErr)
	if saveErr != nil {
		readings.Error = fmt.Sprintf("RTC save: %v", saveErr)
	} else {
		readings.Time.Hour = newTime.Hour()
		readings.Time.Minute = newTime.Minute()
		readings.Time.LastRead = time.Now()
		readings.Error = ""
	}
}

func (a *App) updateReadings(readings *types.Readings) {
	shouldUpdateTimeRead := readings.Time.LastRead.IsZero() ||
		time.Since(readings.Time.LastRead) >= time.Duration(time.Second)

	if shouldUpdateTimeRead {
		curTime, err := a.rtc.ReadTime()
		if err != nil {
			readings.Error = fmt.Sprintf("DS3231: %v", err)
			readings.IsDrawen = false
		} else {
			println("DS3231 time read:", curTime.Format(time.DateTime))
			readings.Time.LastRead = time.Now()
			if readings.Time.Minute != curTime.Minute() {
				println("DS3231 minute changed:", curTime.Format(time.DateTime))
				readings.Time.Minute = curTime.Minute()
				readings.Time.Hour = curTime.Hour()
				readings.IsDrawen = false
			}
		}
	}

	shouldUpdateSensorRead := cmp.Or(
		// First ever reading
		readings.LastUpdateAt.IsZero(),
		// Initial startup period
		time.Since(readings.LastUpdateAt) >= time.Duration(time.Second) &&
			time.Since(readings.FirstReadingAt) < time.Duration(time.Minute),
		// Regular update interval
		time.Since(readings.LastUpdateAt) >= time.Duration(time.Minute),
	)

	if shouldUpdateSensorRead {
		if raw, err := a.sensors.Read(); err != nil {
			readings.Error = err.Error()
			readings.IsDrawen = false
		} else {
			readings.AddReadings(
				raw.CO2,
				raw.Temperature,
				raw.Humidity,
			)
			fmt.Printf("%s, time: %02d:%02d, CO2: %d ppm, T: %.2f °C, H: %.2f %%, co2 len: %d, temp len: %d, hum len: %d\n",
			    time.Now().Format(time.DateTime),
				readings.Time.Hour,
				readings.Time.Minute,
				raw.CO2, raw.Temperature, raw.Humidity,
				readings.History.CO2.Len(),
				readings.History.Temperature.Len(),
				readings.History.Humidity.Len(),
			)
			readings.Error = ""
			readings.IsDrawen = false
		}
	}
}

func (a *App) render(readings *types.Readings) {
	if !readings.IsDrawen {
		readings.ClockEdit = types.ClockEdit{
			Active: a.editor.Active,
			Field:  int(a.editor.Field),
			Hour:   a.editor.Hour,
			Minute: a.editor.Minute,
		}
		a.displayManager.Render(readings)
		readings.IsDrawen = true
	}
}
