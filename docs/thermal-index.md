# Which thermal index to show for sleep

Short version: the NWS heat index that `internal/types/status/heatindex.go` implements is
unusable in an air conditioned bedroom. Use the Steadman apparent temperature instead, and
put the sleep thresholds on that scale, not on dry-bulb temperature.

## Why the heat index does not work here

`HeatIndexVal` returns the air temperature unchanged whenever it is below 27 °C, because the
Rothfusz regression is only fitted above that point:

```
T=24.0 RH=55  -> HI=24.0  No heat
T=26.0 RH=62  -> HI=26.0  No heat
T=26.9 RH=96  -> HI=26.9  No heat     <- 96% humidity, index still equals temperature
T=27.0 RH=96  -> HI=31.4  Caution     <- +4.5 K step across 0.1 K of real change
T=28.0 RH=65  -> HI=30.0  Caution
```

Consequences for this device:

- In the 24-27 °C range an air conditioned room lives in, the index is a copy of the
  thermometer and humidity is ignored completely.
- `GetHeatIndex` therefore returns `NoHeat` for every night that is actually air conditioned,
  which is why the two heat-index dots on the main screen are permanently empty.
- At the 27 °C boundary the index jumps by about 4.5 K, so any bar or scale built on it would
  sit flat all night and then leap a third of its length in one step.

The heat index also models the wrong situation: it is a daytime heat-stress index for a walking
person in shade with light wind, not for a sleeping person under bedding in still air.

## Apparent temperature (Steadman)

Use the Bureau of Meteorology form without the wind term:

```
AT = T + 0.33 * e - 4.00
e  = RH/100 * 6.112 * exp(17.62 * T / (243.12 + T))
```

`e` is the water vapour pressure in hPa. The project already computes exactly this saturation
pressure inside `status.AbsoluteHumidityGM3`, so the same Magnus coefficients should be reused
rather than introducing the slightly different constants from the original BOM publication (the
difference is below 0.1 hPa).

Properties that matter here:

- Continuous and defined at every temperature, so a scale built on it moves every night.
- Humidity shifts the value by a useful amount: at 26 °C, going from 40% to 62% RH moves AT
  from 26.4 to 28.9.
- The "27 °C but unbearable" nights land where they belong: 27 °C at 96% RH gives AT 34.3,
  while 28 °C at 65% RH gives 32.1.

Worked values:

| T, °C | RH, % | e, hPa | AT, °C | heat index, °C |
| ----- | ----- | ------ | ------ | -------------- |
| 22.0  | 50    | 13.2   | 22.4   | 22.0           |
| 24.0  | 55    | 16.4   | 25.4   | 24.0           |
| 26.0  | 40    | 13.4   | 26.4   | 26.0           |
| 26.0  | 62    | 20.8   | 28.9   | 26.0           |
| 27.0  | 96    | 34.1   | 34.3   | 31.4           |
| 28.0  | 65    | 24.5   | 32.1   | 30.0           |
| 29.0  | 80    | 32.0   | 35.6   | 34.7           |

## Thresholds, and the conversion trap

Every threshold quoted in the sleep literature is a **dry-bulb** (or operative) temperature
measured at roughly 50-60% RH. Applying those numbers directly to an AT scale makes the scale
about 3 K too strict. Convert the reference conditions through the same formula first:

| Source condition                           | dry-bulb | AT at that RH |
| ------------------------------------------ | -------- | ------------- |
| Neutral reference used in chamber studies   | 26 °C / 55% | 28.1 °C    |
| Upper comfort limit with adjustable bedding | 29 °C / 55% | 32.3 °C    |

So the scale zones used by the display are:

| Zone | AT          | Meaning                                                              |
| ---- | ----------- | -------------------------------------------------------------------- |
| 1    | below 28    | At or below the neutral condition the sleep experiments use as their baseline. |
| 2    | 28 - 32     | Warm. Still reachable by uncovering, per the Lomas & Li review of 50+ papers. |
| 3    | above 32    | Matches the 29-35 °C / 50-75% RH conditions where reduced slow wave sleep, reduced REM and increased wakefulness are documented. |

Evidence behind the zones:

- 26 °C at 50-60% RH is the standard neutral comparator in the sleep chamber experiments.
- Okamoto-Mizuno's humid heat studies (29 and 35 °C at 50% and 75% RH) found disturbed core
  temperature decline, reduced slow wave sleep, reduced REM and more arousals, with humid heat
  worse than dry heat at the same temperature.
- Lomas & Li (2023) reviewed over 50 papers and concluded that with modern summer bedding and
  bedwear, whose coverage is easy to adjust, comfort is attainable up to about 29 °C. They
  propose bedroom overheating thresholds of 26-29 °C depending on the group (27 °C for
  vulnerable people, 28 °C for healthy adults), and show that the existing UK 26 °C threshold
  derives from one 1970s study whose participants used sheets and eiderdowns, with duvet users
  explicitly excluded.
- Sleep-physiology work is more conservative than the building-services work: sleep efficiency
  is reported to start declining above 24 °C, with clearer effects at 27-30 °C (longer sleep
  onset, more awakenings, less total sleep).

The two literatures disagree by roughly 2 K, so the zone edges are a starting point, not a
constant of nature. They are worth re-tuning after a few weeks of comparing the reading against
how the night actually went.

## How the display encodes this

`RenderSleepScale` draws the zones as a bar along the bottom edge
(`internal/display/thermal_scale.go`):

- The comfortable range is only a 1 px ruler line, and the boundary at AT 28 is a short post.
  An empty scale therefore means there is nothing to do.
- Past the boundary a solid 5 px block grows to the right, so its length is how far beyond the
  limit the room has gone.
- A notch inside the block marks AT 32, the point where the sleep experiments document broken
  sleep rather than merely warm sleep.

The scale spans AT 24 to 36 and clamps at both ends. Filling the bar from the left edge was
tried first and rejected: a half filled bar in a perfectly good room reads as a warning.

`RenderBarsWithTrend` uses the same zones for the two dots next to `T`, which is what the heat
index used to drive - those dots were permanently empty for the reason described above.

## Two facts that shape when to cool

- Thermoregulation is largely suspended during REM: no effective sweating or shivering. A warm
  room therefore truncates REM first, and REM-dense cycles are concentrated in the second half
  of the night, toward morning.
- Humid heat during the **first** half of the night degrades sleep stages and the core
  temperature decline more than the same exposure during the second half. If air conditioning
  has to be rationed (for example to keep CO2 down by opening the room later), spend it early.

## Sources

- [Effects of thermal environment on sleep and circadian rhythm - Okamoto-Mizuno & Mizuno, J Physiol Anthropol 31:14 (2012)](https://link.springer.com/article/10.1186/1880-6805-31-14)
- [Effects of humid heat exposure on human sleep stages and body temperature (1999)](https://pubmed.ncbi.nlm.nih.gov/10505822/)
- [Effects of partial humid heat exposure during different segments of sleep](https://www.sciencedirect.com/science/article/abs/pii/S0031938404004020)
- [An overheating criterion for bedrooms in temperate climates - Lomas & Li, BSER&T (2023)](https://journals.sagepub.com/doi/10.1177/01436244231183113)
- [Thermal environment and sleep quality: A review - Lan, Tsuzuki, Liu & Lian, Energy and Buildings](https://www.sciencedirect.com/science/article/abs/pii/S0378778817317681)
- [Optimizing bedroom thermal environment: A review - Xu & Lian, Energy and Built Environment 5 (2024) 829-839](https://www.sciencedirect.com/science/article/pii/S2666123323000570)
