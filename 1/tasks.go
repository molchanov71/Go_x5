package lesson2

// FizzBuzz возвращает FizzBuzz, Fizz, Buzz или пустую строку для обычного числа.
// Для n <= 0 возвращает invalid.
func FizzBuzz(n int) string {
	switch {
	case n > 0:
		switch {
		case n%15 == 0:
			return "FizzBuzz"
		case n%3 == 0:
			return "Fizz"
		case n%5 == 0:
			return "Buzz"
		default:
			return ""
		}
	default:
		return "invalid"
	}
}

// Sum возвращает сумму. Для пустого среза и nil возвращает 0.
func Sum(values []int64) int64 {
	if values == nil {
		return 0
	}
	var res int64 = 0
	for _, v := range values {
		res += v
	}
	return res
}

// Min возвращает минимум. Для пустого среза и nil возвращает 0 по условию задачи.
func Min(values []int64) int64 {
	if (values == nil) || (len(values) == 0) {
		return 0
	}
	mmin := values[0]
	for _, v := range values {
		if v < mmin {
			mmin = v
		}
	}
	return mmin
}

// Max возвращает максимум. Для пустого среза и nil возвращает 0 по условию задачи.
func Max(values []int64) int64 {
	if (values == nil) || (len(values) == 0) {
		return 0
	}
	mmax := values[0]
	for _, v := range values {
		if v > mmax {
			mmax = v
		}
	}
	return mmax
}

// Average возвращает среднее как float64. Для пустого среза и nil возвращает 0.
func Average(values []int64) float64 {
	if (values == nil) || (len(values) == 0) {
		return 0
	}
	var sum float64 = 0
	count := 0
	for _, v := range values {
		sum += float64(v)
		count++
	}
	return sum / float64(count)
}

// Category возвращает invalid, small, regular или large по сумме в копейках.
func Category(amount int64) string {
	switch {
	case amount < 0:
		return "invalid"
	case amount <= 9999:
		return "small"
	case amount <= 49999:
		return "regular"
	default:
		return "large"
	}
}

// RuneCount считает кодовые точки Unicode в строке через range.
func RuneCount(text string) int {
	count := 0
	for range text {
		count++
	}
	return count
}
