package common

import (
	"strconv"
)

func Filter(slice []string, f func(string) bool) []string {
	var ret []string
	for _, s := range slice {
		if f(s) {
			ret = append(ret, s)
		}
	}
	return ret
}

func MapSlice(slice []string, f func(string) string) []string {
	var ret []string
	for _, s := range slice {
		ret = append(ret, f(s))
	}
	return ret
}

func Contains(slice []string, key string) bool {
	for _, s := range slice {
		if s == key {
			return true
		}
	}
	return false
}

func Remove(slice []string, element string) []string {
    var ret []string

    for _, v := range slice {
        if v != element {
            ret = append(ret, v)
        }
    }

    return ret
}

func IsPositiveIntegerToken(token string) bool {
	n, err := strconv.Atoi(token)
	if err != nil {
		return false
	}
	return n > 0
}

func IsNumericToken(token string) bool {
	_, err := strconv.ParseFloat(token, 64)
	return err == nil
}