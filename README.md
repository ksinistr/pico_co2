# pico_co2

Raspberry Pico CO2 measurements

This project is designed to show time, temperature, CO2 levels using a Raspberry Pi Pico microcontroller. It integrates sensors and displays to provide real-time data visualization.

## Hardware installation

### Required components

- Raspberry Pico board
- SSD1306 display
- AHT20+ENS160 sensor
- SCD4x sensor
- DS3231
- touch buttons (2x)

## Button Controls

Two capacitive touch buttons control display navigation and clock setup.

### Normal Mode

- Left button: previous display screen
- Right button: next display screen

### Clock Setup Mode

Clock setup allows setting the DS3231 real-time clock on-device using the two touch buttons.

**Entering clock setup**

1. Navigate to the time screen (first display)
2. Press both buttons within a ~400 ms window
3. The display switches to edit mode, showing the current hour in brackets

**Editing time**

- Left button: cycle through fields in order: hour -> minute -> save -> cancel
- Right button:
  - On hour field: increment hour (wraps 23 -> 0)
  - On minute field: increment minute (wraps 59 -> 0)
  - On save: write the new time to the RTC and exit edit mode
  - On cancel: discard changes and exit edit mode

**Display indicators in edit mode**

- `[14]:05` - hour selected (shown in brackets)
- `14:[05]` - minute selected
- `[SAVE] EXIT` - save selected
- `SAVE [EXIT]` - cancel selected

**Exiting without saving**

- Press both buttons again at any time to cancel and return to normal mode
- Navigate to the cancel field and press right

## Software installation

```bash
make flash
```

## Generate all possible display themes

```bash
make test-displays
```

## Display Themes

Below are examples of the different display themes available:

### Main Display

![Main Display](images/RenderTime-normal.png)

### CO2 Graph Display

![CO2 Graph Display](images/RenderSparklineCO2-normal.png)

## Case

![20250714_065327_](https://github.com/user-attachments/assets/1119b4b1-9fd0-45b7-aa6f-d544de5fbf7b)

## Development

### requirements for vim development

- go version go1.24.4 linux/amd64
- tinygo version 0.39.0 linux/amd64
- go install github.com/sago35/tinygo-edit@latest
- run as `tinygo-edit --target pico --editor nvim --wait`
