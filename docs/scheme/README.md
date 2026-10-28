# Generating the wiring diagrams

This directory contains two standalone SVG images and their source, [generate.py](generate.py):

- [pico-scd40.svg](pico-scd40.svg): SCD40 variant.
- [pico-aht20-ens160.svg](pico-aht20-ens160.svg): AHT20 + ENS160 variant.

[hardware.md](../hardware.md) embeds them with ordinary Markdown: `![description](scheme/name.svg)`. GitHub displays SVG images and their text stays sharp when enlarged. The files contain no scripts, external fonts or linked images. In a browser, `width: 100%; height: auto` preserves the aspect ratio and allows vertical scrolling.

## Regenerate

From the repository root, using Python 3.11+ and `uv`:

```sh
uv run --no-project docs/scheme/generate.py
```

No external Python dependencies are needed. The command overwrites both SVG files beside the generator; output is deterministic. PNG copies are not required for GitHub.

## Update the diagrams

1. Check the current configuration in [pico_scd40/main.go](../../cmd/pico_scd40/main.go) and [pico_aht20ens160/main.go](../../cmd/pico_aht20ens160/main.go), address defaults in [config.go](../../internal/config/config.go), and flash commands in the [Makefile](../../Makefile).
2. Edit `pico()` for physical Pico pins, `module()` for module terminals, and `render()` for components, wires and captions. `main()` passes the palette and bus positions explicitly. Pin assignments are maintained manually; the generator does not parse Go source. When changing GPIO assignments, also update the checklist and SVG description in `render()`.
3. Regenerate both images. If Ruff is available, check the source:

   ```sh
   ruff format --check docs/scheme/generate.py
   ruff check docs/scheme/generate.py
   ```

4. Open **both** SVG files in a browser. Check text readability, wire endpoints and solder joints. A filled dot means an electrical connection; a white gap at a crossing means no connection. Also check the images in a Markdown preview.
5. Save the generator, both SVGs and documentation changes together. Do not edit the SVG files manually: regeneration will overwrite those changes.

## Pin assignments

The Pico is shown from above, with components facing the viewer and USB at the top. Numbers 1–40 are physical pins; GP labels identify firmware GPIOs.

| Pico pin | Function | Connect to |
| --- | --- | --- |
| 36 · 3V3 OUT | 3.3 V power | VCC on OLED/RTC/buttons; VDD on SCD40; 3V3, ADD and CS on the combined sensor board |
| 38 · GND | Common ground | GND on all modules and buttons |
| 6 · GP4 | I²C0 SDA | SDA on the display, clock and sensors |
| 7 · GP5 | I²C0 SCL | SCL on the display, clock and sensors |
| 14 · GP10 | Left button | OUT/I/O on the left TTP223 |
| 15 · GP11 | Right button | OUT/I/O on the right TTP223 |

Current firmware addresses: SSD1306 `0x3C`, DS3231 `0x68`, SCD40 `0x62`, AHT20 `0x38`, ENS160 `0x53`.

## Module orientations

Keep the following board orientations and pin orders when regenerating the images; they determine the wire endpoints in `module()` and routing in `render()`.

| Module | View in the diagrams | Pin order |
| --- | --- | --- |
| SSD1306 | Display facing the viewer, header at the top | Left to right: GND, VCC, SCL, SDA |
| DS3231 + AT24C32 | Battery facing the viewer, six-pin header on the right | Top to bottom: 32K, SQW, SCL, SDA, VCC, GND |
| TTP223 | Components facing the viewer, header on the left | Top to bottom: GND, I/O, VCC |
| SCD40 | Sensor facing the viewer, header at the bottom | Left to right: GND, VDD, SCL, SDA |
| AHT20 + ENS160 | Components facing the viewer, header at the bottom | Left to right: VIN, 3V3, GND, SCL, SDA, ADD, CS, NT/INT |

On DS3231, 32K and SQW are marked **NC** and must remain unconnected. The spare four-pin header on the left repeats SCL, SDA, VCC and GND; leave it unused when wiring the six-pin header shown. TTP223 I/O is the button output connected to GP10 or GP11.

On the pictured combined sensor board, use **3V3** as the regulated 3.3 V supply input and leave **VIN** unconnected. Connect **ADD** to 3.3 V for address `0x53` and **CS** to 3.3 V for I²C mode. Leave **NT/INT** (the interrupt output) unconnected. This is the selected 3.3 V wiring arrangement; do not supply VIN at the same time. The combined board configuration is supported by [Digital Concepts' firsthand wiring and tests](https://digitalconcepts.net.au/arduino/index.php?op=ENS160); address and interface selection follow the [ScioSense datasheet](https://www.sciosense.com/wp-content/uploads/2023/12/ENS160-Datasheet.pdf).

The display uses **128×32** pixels, as configured in the firmware.

## Module assumptions

The diagrams show breakout boards rated for a 3.3 V input, rather than bare chips. Pin orders are listed in the table above; other board revisions may arrange them differently. Check each board's labels and schematic before soldering. Supply 3.3 V to VCC on the OLED/RTC/buttons, VDD on SCD40, or 3V3 on the pictured combined board. Do not assume the same power-pin function on a different breakout. The bare ENS160 requires a 1.8 V core supply, so its breakout must provide the appropriate regulator.

The combined AHT20 + ENS160 board is shown with all eight terminals. For separate boards, connect the appropriate 3.3 V power input, GND, SDA and SCL on **each** board to the matching nets. Configure ENS160 for I²C address `0x53`; follow each module's documentation for CS/ADDR terminals. Extra INT/SQW/32K outputs are unused.

The diagrams assume onboard SDA/SCL pull-ups to 3.3 V. If absent, add pull-ups following the module documentation; account for the combined resistance when several boards include them. Join the selected power inputs to Pico 3V3 OUT and all GND terminals to common ground. Keep 5 V off the GPIO and I²C lines.

USB powers the Pico. The SCD40 can draw up to 205 mA peak at 3.3 V; the supply and wiring must support all modules operating together. Touch buttons must be active HIGH in momentary mode, matching the pull-down inputs and rising-edge handling in [button.go](../../internal/hardware/pico/button.go).

## Reference sources

- [Official Raspberry Pi Pico / Pico 2 pinout](https://datasheets.raspberrypi.com/pico/Pico-2-Pinout.pdf): physical pin numbers and GPIO names.
- [Raspberry Pi: Pico power pins](https://www.raspberrypi.com/documentation/microcontrollers/pico-series.html#pin-functions): 3V3 OUT, VBUS and VSYS.
- [Sensirion: SCD4x Datasheet](https://admin.sensirion.com/media/documents/48C4B7FB/64C134E7/Sensirion_SCD4x_Datasheet.pdf): power, I²C and SCD40 peak current.
- [ScioSense: ENS160 Datasheet](https://www.sciosense.com/wp-content/uploads/2023/12/ENS160-Datasheet.pdf): core and interface supplies, I²C mode, address selection and eCO₂.
- [DFRobot: an example ENS160 breakout](https://wiki.dfrobot.com/sen0537/): 3.3–5 V input and address `0x53`; an example, not an identification of the user's board.
- [Adafruit: AHT20 breakout pinout](https://learn.adafruit.com/adafruit-aht20/pinouts): address `0x38`, power and onboard pull-ups.
