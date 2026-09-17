# Bedding, and what "comfort up to 29 °C" actually requires

The claim that a bedroom can stay comfortable up to about 29 °C comes from Lomas & Li (2023),
who reviewed over 50 papers while deriving a new bedroom overheating criterion. The exact
wording is that with **modern summertime bedding and bedwear, which enables body coverage to be
easily adjusted**, comfort is attainable up to roughly 29 °C night-time bedroom temperature.

The conditional half of that sentence is the important one. The number does not come from a
thicker or thinner duvet; it comes from bedding you can uncover in stages during the night.

## What the body actually regulates

The target is not room air but the **bed microclimate**, the air pocket between skin and
covers. Measured neutral bedding temperature is about 31.1 °C in chamber experiments and
32.1 °C in field studies. Room air, bedding insulation and coverage are just the three knobs
that set it.

## The insulation trade-off

Measured total insulation of real bedding systems spans a very wide range:

| Study | Range | What was varied |
| ----- | ----- | --------------- |
| Lin & Deng, subtropical manikin (2008) | 0.90 - 4.89 clo | bedding, bed, mattress, sleepwear, coverage |
| Manikin + JOS-3 model (2025) | 1.06 - 5.71 clo | 84 conditions: supine/lateral, nude/pyjamas, blanket/duvet, coverage 23.3 / 54.4 / 70.6 / 94.1% |

The exchange rate between room temperature and insulation, from Lin & Deng's Fanger-based
model, is:

> Between 26 and 32 °C operative temperature, a 1 K rise is compensated by **0.19 clo less**
> bedding plus bedwear. At 1.0 clo (mattress only, i.e. effectively uncovered) the neutral
> temperature is 29.5 °C at 50% RH.

Unit conversion, since duvets are sold in tog: 1 clo = 0.155 m²K/W, 1 tog = 0.1 m²K/W, so
1 clo ≈ 1.55 tog and **0.19 clo ≈ 0.3 tog per degree**. Going from a 4.5 tog summer duvet down
to a sheet is worth roughly 3-4 K of room temperature, which is exactly the gap between an AC
setpoint of 25 and one of 28-29.

Air movement buys a little more: at under 0.6 clo of summer bedding and bedwear, about
0.4 m/s of air raises the neutral temperature by roughly 0.9 K. Larger effects are reported for
faster flow - a fan at 0.6-0.7 m/s at 30 °C gave sleep variables comparable to AC at 0.2 m/s and
27 °C in an elderly cohort, and sufficient isothermal airflow made high temperature and humidity
close to 26 °C / 50% in terms of total sleep time and sleep efficiency.

## What to buy and how to use it

Ordered by how much it actually moves the number:

1. **Layers you can shed, not one optimal duvet.** Coverage percentage moves insulation across
   most of the 1-5 clo range on its own, and subjects in chamber studies kept adjusting coverage
   toward neutrality even when the quilt was held constant. A sheet plus a light duvet beats a
   single medium one, because at 03:00 the choice is not "thinner duvet", it is "legs out".
2. **Low tog for the top layer.** Retailer convention (commercial guidance, not research):
   summer duvets are 2.5-4.5 tog, with 1-3 tog marketed for hot sleepers. 4.5 tog ≈ 2.9 clo is
   already a lot for a Cyprus August; the useful summer pair is a cotton or linen flat sheet
   (well under 1 tog) with a 2.5-4.5 tog duvet kept folded at the foot of the bed.
3. **Fills and shells that move moisture.** Every insulation figure above comes from a dry
   thermal manikin, which cannot sweat; the 2025 study had to bolt on a thermoregulation model
   to get total heat loss. At 90%+ RH evaporation is what fails first, so shell and fill matter
   more than tog: cotton percale, linen, wool or bamboo rather than sealed microfibre, and
   crucially **no plastic-backed waterproof mattress protector**, which blocks the largest
   evaporative surface you have.
4. **Sleepwear as a regulator.** The manikin studies treat nude vs pyjamas as a separate axis
   worth a meaningful slice of the total clo. Light long sleepwear in wicking fabric can be
   better than bare skin under a sheet, because it spreads sweat over a larger evaporating area.
5. **A fan, for the 1-2 K it is worth.** Cheaper than the same setpoint drop on the AC, and it
   also breaks up the stagnant humid layer over the bed. It does not dehumidify: at 96% RH the
   fan improves heat loss but only the AC or a dehumidifier removes water.

## Applying it to a closed, air conditioned room

The specific bind here is that closing doors for the AC drives CO2 up, so the AC cannot simply
run all night with the room sealed. Given the evidence above:

- Spend the cold early. Humid heat in the first half of the night is more damaging than the
  same exposure later, so a lower setpoint before and just after falling asleep is worth more
  than the same energy spent at 04:00.
- Use bedding, not the setpoint, to cover the last 2-3 K. That is the difference between a
  setpoint the room can hold with the door cracked open and one that requires it sealed.
- Treat humidity as a separate axis from temperature. The AC removes water only while its coil
  actually runs; a high setpoint with short compressor cycles cools without drying, which is
  how a 27 °C room ends up at 96% RH. If the reading is high and humid, a lower setpoint for a
  while (or dry mode) dehumidifies more than fan speed does.
- Open up before CO2, not after. Since ventilation is the only lever for CO2 and it costs
  temperature, the cheapest time to air the room is early in the night while the outdoor
  temperature is still falling.

## Sources

- [An overheating criterion for bedrooms in temperate climates - Lomas & Li, BSER&T (2023)](https://journals.sagepub.com/doi/10.1177/01436244231183113)
- [A study on thermal comfort in sleeping environments in the subtropics - Lin & Deng, Building and Environment 43:905-916 (2008)](https://www.sciencedirect.com/science/article/abs/pii/S0360132307000261)
- [Effect of bedding on total thermal insulation in different sleeping postures, manikin + JOS-3 (2025)](https://www.sciencedirect.com/science/article/pii/S0360132325005554)
- [Effects of bedding insulation and indoor temperature on bed microclimate and thermal comfort, Energy and Buildings (2020)](https://www.sciencedirect.com/science/article/abs/pii/S0378778819332360)
- [Optimizing bedroom thermal environment: A review - Xu & Lian (2024)](https://www.sciencedirect.com/science/article/pii/S2666123323000570)
- [A model for predicting thermal comfort during sleep in response to air velocity, Building and Environment (2022)](https://www.sciencedirect.com/science/article/abs/pii/S0360132322007090)
- [Air-conditioning for sleeping environments in tropics and/or sub-tropics - a review, Energy (2013)](https://www.sciencedirect.com/science/article/abs/pii/S0360544213000182)
