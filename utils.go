package ddlparse

import (
	"strconv"
)

func filter(slice []string, f func(string) bool) []string {
	var ret []string
	for _, s := range slice {
		if f(s) {
			ret = append(ret, s)
		}
	}
	return ret
}

func mapSlice(slice []string, f func(string) string) []string {
	var ret []string
	for _, s := range slice {
		ret = append(ret, f(s))
	}
	return ret
}

func contains(slice []string, key string) bool {
	for _, s := range slice {
		if s == key {
			return true
		}
	}
	return false
}

func remove(slice []string, element string) []string {
    var ret []string

    for _, v := range slice {
        if v != element {
            ret = append(ret, v)
        }
    }

    return ret
}

func isNumericToken(token string) bool {
	_, err := strconv.ParseFloat(token, 64)
	return err == nil
}

func isStringToken(token string) bool {
	tmp := token[0:1]
	return tmp == "\"" || tmp == "'" || tmp == "`"
}