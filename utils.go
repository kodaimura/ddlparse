package ddlparse


func filter(slice []string, f func(string) bool) []string {
	var ret []string
	for _, s := range slice {
		if f(s) {
			ret = append(ret, s)
		}
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