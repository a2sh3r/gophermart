package utils

import "strconv"

func IsValidLuhn(s string) bool {
	var sum int
	var alt bool

	for i := len(s) - 1; i >= 0; i-- {
		num, err := strconv.Atoi(string(s[i]))
		if err != nil || num < 0 || num > 9 {
			return false
		}

		if alt {
			num *= 2
			if num > 9 {
				num -= 9
			}
		}

		sum += num
		alt = !alt
	}

	return sum%10 == 0
}
