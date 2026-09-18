package sparkline

// Sparkline processes humidity measurements into a sparkline-friendly series.
type Sparkline struct {
	Height int // pixel height of sparkline (e.g., 14)
}

// NewSparkline constructs a Sparkline with dynamic buffer size and height.
func NewSparkline(height int) *Sparkline {
	return &Sparkline{Height: height}
}

// percentile returns the p-th percentile (0–100) from raw data using index method without floats.
func percentile(raw []int16, p int) int16 {
	// copy and sort
	sorted := make([]int16, len(raw))
	copy(sorted, raw)
	// insertion sort (O(N^2) but N is small)
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j-1] > sorted[j]; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}

	idx := p * (len(sorted) - 1) / 100

	return sorted[idx]
}

// median3 computes the median of three values without allocations.
func median3(a, b, c int16) int16 {
	if a > b {
		a, b = b, a
	}

	if b > c {
		if a > c {
			return a
		}

		return c
	}

	return b
}

// Process applies 1% trim, median smoothing, and resamples to length N.
func (s *Sparkline) Process(raw []int16) []int16 {
	rawLen := len(raw)
	if rawLen == 0 {
		return []int16{}
	}

	filtered := trim(raw, percentile(raw, 1), percentile(raw, 99))
	smoothed := smooth(filtered)

	if len(smoothed) == 0 {
		return make([]int16, rawLen)
	}

	minValue, maxValue := minMax(smoothed)

	if minValue == maxValue {
		return constantSeries(rawLen, int16(s.Height/2))
	}

	normalized := normalize(smoothed, minValue, maxValue, s.Height)

	return resample(normalized, rawLen, s.Height)
}

func trim(raw []int16, lower, upper int16) []int16 {
	filtered := make([]int16, 0, len(raw))
	for _, value := range raw {
		if value >= lower && value <= upper {
			filtered = append(filtered, value)
		}
	}

	if len(filtered) == 0 {
		return raw
	}

	return filtered
}

func smooth(values []int16) []int16 {
	smoothed := make([]int16, len(values))
	if len(values) < 2 {
		copy(smoothed, values)

		return smoothed
	}

	for i := range values {
		a, b, c := window(values, i)
		smoothed[i] = median3(a, b, c)
	}

	return smoothed
}

//nolint:nonamedreturns // The result names document the smoothing window.
func window(values []int16, i int) (first, middle, last int16) {
	switch i {
	case 0:
		return values[0], values[0], values[1]
	case len(values) - 1:
		return values[i-1], values[i], values[i]
	default:
		return values[i-1], values[i], values[i+1]
	}
}

//nolint:nonamedreturns // The result names document the two extrema.
func minMax(values []int16) (minValue, maxValue int16) {
	minValue = values[0]
	maxValue = values[0]

	for _, value := range values[1:] {
		if value < minValue {
			minValue = value
		}

		if value > maxValue {
			maxValue = value
		}
	}

	return minValue, maxValue
}

func normalize(values []int16, minValue, maxValue int16, height int) []int16 {
	scale := maxValue - minValue
	for i := range values {
		values[i] = int16((int32(values[i]-minValue) * int32(height-1)) / int32(scale))
	}

	return values
}

func constantSeries(length int, value int16) []int16 {
	series := make([]int16, length)
	for i := range series {
		series[i] = value
	}

	return series
}

func resample(values []int16, outputLength, height int) []int16 {
	result := make([]int16, outputLength)
	if len(values) <= 1 {
		value := int16(height / 2)
		if len(values) == 1 {
			value = values[0]
		}

		return constantSeries(outputLength, value)
	}

	if outputLength <= 1 {
		if outputLength == 1 {
			result[0] = values[0]
		}

		return result
	}

	inputDenominator := len(values) - 1
	outputDenominator := outputLength - 1

	for i := range result {
		numerator := i * inputDenominator
		index := numerator / outputDenominator
		remainder := numerator % outputDenominator

		next := index + 1
		if next >= len(values) {
			next = len(values) - 1
		}

		result[i] = interpolate(values[index], values[next], remainder, outputDenominator)
	}

	return result
}

func interpolate(first, second int16, numerator, denominator int) int16 {
	value := int32(first)*(int32(denominator)-int32(numerator)) + int32(second)*int32(numerator)
	return int16(value / int32(denominator))
}
