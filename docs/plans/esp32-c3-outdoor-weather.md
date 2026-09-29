# Plan: outdoor weather and air conditioner advice on ESP32-C3

Status: idea, hardware not bought yet.

## Problem

The monitor only sees the bedroom. A typical Cyprus summer night:

1. The room is hot in the evening, the air conditioner goes on.
2. Around midnight the room is cool, the air conditioner goes off.
3. The room drifts back to the outdoor temperature, and outdoors it is still hot
   and humid, so the room ends up too warm for sleep again.

The device cannot say in advance whether switching the air conditioner off is safe,
because it does not know the outdoor conditions or where they are heading overnight.

## Plan

1. Buy an ESP32-C3 board. TinyGo now supports it, and it has Wi-Fi on the chip.
   The current Raspberry Pi Pico has no network.
2. Port the firmware: I2C pins, board target in the `Makefile`, check that the
   SSD1306, SCD4x, AHT20 and DS3231 drivers work on the new target.
3. Fetch outdoor temperature and humidity from [open-meteo](https://open-meteo.com)
   (no API key), current values plus the hourly forecast for the rest of the night.
4. Compute the outdoor heat index with the same `status.HeatIndexVal` used indoors.
5. Compare indoor and outdoor values and show advice, for example:
   - outdoor heat index now and later tonight is at or above the warm line (28):
     keep the air conditioner on;
   - outdoor is cooler than indoor: switch it off and open the window;
   - outdoor is cooling down later tonight: switch it off at a suggested hour.

## Open questions

- Exact advice rules and thresholds; start from the sleep scale in
  [sleep-thresholds.md](../sleep-thresholds.md).
- How the advice fits on the 128x32 display: a separate screen or a short word
  on the glance screen.
- Behaviour without Wi-Fi: hide the advice, keep the indoor screens working.
- Memory headroom for HTTPS and JSON parsing on the ESP32-C3 with TinyGo.
