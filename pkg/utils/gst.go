package utils

import "math"

// Round2 rounds a float to 2 decimal places (paise-safe for INR amounts).
func Round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// FinancialYear returns the Indian FY label (e.g. "26-27") for month/year,
// where the year rolls over on April 1st.
func FinancialYear(year int, month int) string {
	start := year
	if month < 4 {
		start = year - 1
	}
	end := start + 1
	return twoDigit(start) + "-" + twoDigit(end)
}

func twoDigit(year int) string {
	y := year % 100
	if y < 10 {
		return "0" + itoa(y)
	}
	return itoa(y)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
