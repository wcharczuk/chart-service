package core

// Min returns the min value of a variable set of float64s.
func Min(values ...float64) float64 {
	if len(values) == 0 {
		return 0
	}
	minV := values[0]
	for _, v := range values {
		if minV > v {
			minV = v
		}
	}
	return minV
}

// Max returns the max value of a variable set of float64s.
func Max(values ...float64) float64 {
	if len(values) == 0 {
		return 0
	}
	maxV := values[0]
	for _, v := range values {
		if maxV < v {
			maxV = v
		}
	}
	return maxV
}
