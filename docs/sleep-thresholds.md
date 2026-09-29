# Scientific sleep-environment thresholds

This document turns the sleep-environment literature into practical thresholds for a
bedroom monitor. The values are for a sustained bedroom exposure or an overnight mean,
not for a single noisy sensor sample. `Critical` means that changing the room is strongly
justified for sleep quality; it does not mean acute poisoning or a medical emergency.

There is no universal sleep cutoff. Age, bedding, clothing, acclimatization, airflow,
ventilation, sleep-disordered breathing and the time of night all change the response. The
thresholds below are therefore evidence-based operating points, not diagnostic limits.

## Recommended operating points

| Signal | Reference / target | First measurable concern | Take action | Strong or critical action |
| --- | --- | --- | --- | --- |
| CO₂, ppm | 750–800 or lower | Around 1,000 ppm as an overnight mean | ≥1,000 ppm sustained or rising | ≥1,300 ppm is a strong ventilation signal; 1,900–2,000 ppm is clearly adverse for sleep |
| Apparent temperature, °C | Usually below 25–27, depending on bedding | Approximately 25–27; this maps to the first effects reported above 24–25 °C dry-bulb | ≥28 °C | ≥32 °C; the underlying dry-bulb evidence is approximately 29 °C at 50–55% RH or 30 °C at 50% RH |
| Relative humidity, %RH | Approximately 45–60% | ≤40% on the dry side; ≥70–75% is a warning when the room is warm | Correct <40%; investigate ≥70–75% together with elevated apparent temperature | ≥80%: dehumidify or ventilate. ≥90% is an extreme engineering trigger, not a separately established sleep-study cutoff |

The most useful single defaults for this project are therefore:

- CO₂: warn at 1,000 ppm, escalate at 1,300 ppm, and treat 2,000 ppm as critical for sleep.
- Apparent temperature: warm at 28 °C and hot/actionable at 32 °C.
- RH: use 60% as a watch line, 70–75% as a warm-room warning, and 80% as the strong action line.

The RH rule must not be read as “60% always damages sleep”. The direct evidence for a
universal upper threshold is weak; high humidity becomes much more important when it also
raises apparent temperature or prevents sweat evaporation.

## CO₂

### What the experiments found

The strongest recent threshold evidence comes from a field-lab experiment with 36 healthy
young adults. Mean bedroom CO₂ was approximately 750, 1,000 or 1,300 ppm. Compared with
750 ppm, the 1,000 ppm condition reduced sleep efficiency by 1.3% and increased awake time
by 5.0 minutes. At 1,300 ppm, sleep efficiency was 1.8% lower and awake time was 7.8 minutes
higher; deep sleep also decreased and waking cortisol increased. The authors concluded that
an average bedroom CO₂ concentration of 1,000 ppm or more should be avoided.

A separate controlled study exposed 24 subjects to approximately 800 and 2,000 ppm in a
thermally neutral bedroom. At 2,000 ppm, subjects reported lower sleep satisfaction and
less ease of awakening, and had shorter total sleep and deep-sleep duration. Another study
tested 800, 1,900 and 3,000 ppm in a bedroom chamber and found a monotonic increase in sleep
onset latency and decrease in slow-wave sleep; the questionnaire score at 3,000 ppm was only
80.8% of the 800 ppm score.

These studies do not prove that CO₂ alone is the only causal agent. In a real bedroom, CO₂
also tracks human bioeffluents and insufficient ventilation. That is still the correct
interpretation for this device: a high reading means that ventilation is probably too low
for good sleep, regardless of whether the effect is caused by CO₂ itself, bioeffluents, or
both.

### CO₂ interpretation

- `≤750–800 ppm`: low-exposure reference condition in controlled studies. This is a useful
  target, not a guarantee of good sleep.
- `800–1,000 ppm`: transition zone. A single reading here is not alarming, but a rising
  overnight trend means ventilation should be improved.
- `≥1,000 ppm`: earliest evidence-supported action point. The 1,000 ppm result is based on
  an overnight mean, so a short spike should not be treated as equivalent to a whole night.
- `≥1,300 ppm`: strong action point. The experiment found additional loss of sleep
  efficiency, more wakefulness and a deep-sleep/cortisol signal.
- `≥1,900–2,000 ppm`: clearly adverse sleep exposure in controlled studies. Ventilate now;
  this is a sleep-quality threshold, not an acute toxicity limit.

## Apparent temperature

The sleep literature usually reports dry-bulb or operative temperature, not the Steadman
apparent temperature used by this project. The conversion used here is:

```text
AT = T + 0.33 × e − 4.00
e  = RH/100 × 6.112 × exp(17.62 × T / (243.12 + T))
```

There is no wind term because the bedroom sensor does not measure air velocity. The formula
and its limitations are described in [thermal-index.md](thermal-index.md). Apparent
temperature is useful for a display because it moves continuously with humidity, but it is
not a complete model of the temperature inside the bed. Bedding, uncovered skin and a fan
can change the result substantially.

The conversion explains the project thresholds:

| Research-like bedroom condition | Same reading on this AT scale |
| --- | ---: |
| 24 °C / 50% RH | 24.9 °C |
| 25 °C / 55% RH | 26.7 °C |
| 26 °C / 55% RH | 28.1 °C |
| 29 °C / 55% RH | 32.3 °C |
| 30 °C / 50% RH | 33.0 °C |
| 32 °C / 80% RH | 40.5 °C |

The evidence behind the operating points is consistent across several study types:

- In a climate-chamber pilot with 10 young adults, increasing the room from 24 °C to 28 °C
  reduced subjective sleep quality and sleep efficiency and lengthened sleep-onset latency.
- In a controlled experiment with 16 adults over 65, 30 °C instead of 27 °C reduced total
  sleep time by 26.3 minutes, sleep efficiency by 5.5% and REM sleep by
  5.3 minutes, while awake time increased by 27.0 minutes.
- A year-long home study of older adults found the best sleep efficiency at approximately
  20–24 °C and a clinically relevant 5–10% reduction when bedroom temperature increased by
  another 5 °C.
- A review of more than 50 papers proposed mean night-time dry-bulb thresholds of 27 °C for
  vulnerable occupants and 28 °C for other healthy occupants; approximately 29 °C was the
  upper range at which healthy people could still reach comfort by adjusting bedding.

Those are dry-bulb/operative values. On this project's humidity-aware scale, `AT 28 °C`
is a reasonable first warm/action boundary and `AT 32 °C` is a strong action boundary. The
boundaries are intentionally not presented as exact biological breakpoints.

## Relative humidity

Humidity has the least defensible single upper cutoff. It affects sleep in two different
ways: it can affect the airways and breathing independently of temperature, and it reduces
evaporative cooling when the room is warm. The second mechanism is why RH cannot be evaluated
without temperature or apparent temperature.

The most direct recent experiment tested older adults at overnight mean RH values of about
40%, 58% and 82%, with 60% as the intended middle condition. Sleep quality was best near
60%; at 40%, compared with 60%, sleep efficiency decreased by 3.5 percentage points, deep
sleep decreased by 8.9 minutes, sleep-onset latency increased by 5.6 minutes and wake time
after sleep onset increased by 14.4 minutes. The high-RH condition near 80% also produced
lower overall sleep quality than the 60% condition. Higher CO₂ made the low/high humidity
effect worse.

The classic humid-heat experiment with seven young men compared 29 °C/50%, 29 °C/75%,
35 °C/50% and 35 °C/75%. The clearly broken-sleep condition was 35 °C/75%, which produced
more wakefulness and lower sleep efficiency, stage 3/4 and REM than the cooler conditions.
The 29 °C/75% versus 29 °C/50% comparison was not a strong demonstration of an RH-only
threshold. A later partial-exposure experiment similarly used 32 °C/80% versus 26 °C/50%
and found more wakefulness under the hot-humid exposure. These studies support a joint
heat-and-humidity alarm, not a claim that every room above 60% RH causes poor sleep.

### RH interpretation

- `45–60%`: practical target band. It is close to the best-controlled condition in the
  older-adult experiment and avoids both very dry and very humid air.
- `<40%`: earliest direct low-humidity action point supported by the older-adult experiment.
  Below 30% should be treated as severely dry, although 30% is not a sleep-specific causal
  cutoff.
- `60–70%`: watch band. A reading above 60% is useful to display, but by itself it is not
  enough evidence to call sleep quality impaired.
- `70–75%`: humid warning, especially when `AT ≥28 °C`. Research in this range is mixed when
  temperature is neutral, but the margin for evaporative cooling is becoming smaller.
- `≥80%`: strong action point. This is close to the high-RH condition that reduced sleep
  quality in the older-adult laboratory study and to the 32 °C/80% humid-heat exposure.
- `≥90%`: extreme operational condition. It should trigger immediate dehumidification or
  ventilation, but it should not be described as a scientifically established sleep
  breakpoint distinct from 80%.

## How to use the thresholds in the device

The studies mostly used a mean condition over the sleep period. A practical firmware policy
should therefore separate a short-term action alarm from a sleep-exposure assessment:

1. Smooth readings with a short rolling median or mean; do not alarm on one sample.
2. Use the overnight mean, or a long-duration mean, when comparing a night with the studies.
3. Treat `CO₂ ≥1,000 ppm`, `AT ≥28 °C` and `RH ≥70%` as early warnings that should prompt
   ventilation, airflow, bedding adjustment or dehumidification.
4. Treat `CO₂ ≥1,300 ppm`, `AT ≥32 °C` or `RH ≥80%` as strong action conditions.
5. Do not add the apparent-temperature and RH alarms as if they were independent points of
   heat load. AT already includes humidity; keep the separate RH alarm because humidity may
   also affect breathing and the bedroom fabric/mould environment.

For vulnerable or elderly sleepers, use the lower temperature end of the evidence: act near
`AT 28 °C` and do not wait for `AT 32 °C`. For a healthy adult with adjustable bedding, a
short excursion near 28 °C is not automatically a failed night, but a sustained value above
32 °C should be corrected.

## Implications for the current display

The Glance scales indicate increasing need for intervention. Their operating points are
engineering choices informed by the evidence above, not validated biological cutoffs.

| Signal | First section | Full bar |
| --- | ---: | ---: |
| CO₂ | 1,000 ppm | 1,900 ppm |
| Apparent temperature | 28 °C | 32 °C |
| Relative humidity | 70% | 80% |

On the sectioned screens, values below the lower threshold draw nothing. Reaching the
lower threshold draws the first section; the remaining sections appear at equally spaced
values up to the upper threshold. Only reaching the upper threshold fills the entire bar.
Intermediate warning marks are not drawn. The humidity bar only indicates excessive
humidity; dry air is not represented as a warning by this scale.

The temperature number remains the measured air temperature; its bar uses apparent
temperature. At 27 °C and 61% RH, AT is approximately 30.2 °C: three of five sections
(or four of seven narrower sections) are filled, while the humidity bar is empty.

## Sources

- [Kang et al. (2024), *Ventilation causing an average CO₂ concentration of 1,000 ppm negatively affects sleep*](https://doi.org/10.1016/j.buildenv.2023.111118)
- [Zhang et al. (2023), *Effects of exposure to carbon dioxide and human bioeffluents on sleep quality and physiological responses*](https://doi.org/10.1016/j.buildenv.2023.110382)
- [Xu et al. (2021), *Experimental study on sleep quality affected by carbon dioxide concentration*](https://doi.org/10.1111/ina.12748)
- [Fan et al. (2022), *The effects of ventilation and temperature on sleep quality and next-day work performance*](https://doi.org/10.1016/j.buildenv.2021.108666)
- [Yan et al. (2022), *Experimental study of the negative effects of raised bedroom temperature and reduced ventilation*](https://doi.org/10.1111/ina.13159)
- [Baniassadi et al. (2023), *Nighttime Ambient Temperature and Sleep in Community-Dwelling Older Adults*](https://pmc.ncbi.nlm.nih.gov/articles/PMC10529213/)
- [Lomas & Li (2023), *An overheating criterion for bedrooms in temperate climates*](https://doi.org/10.1177/01436244231183113)
- [Yan et al. (2025), *How humidity and CO₂ affect the sleep of older adults?*](https://doi.org/10.1016/j.buildenv.2025.112628)
- [Okamoto-Mizuno et al. (1999), *Effects of humid heat exposure on human sleep stages and body temperature*](https://pubmed.ncbi.nlm.nih.gov/10505822/)
- [Okamoto-Mizuno et al. (2005), *Effects of partial humid heat exposure during different segments of sleep*](https://doi.org/10.1016/j.physbeh.2004.09.009)
- [Bureau of Meteorology, apparent temperature calculation](https://www.bom.gov.au/climate/maps/averages/apparent-temperature/)
