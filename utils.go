package ddlparse


func filter(array []string, f func(string) bool) []string {
	var ret []string
	for _, s := range array {
		if f(s) {
			ret = append(ret, s)
		}
	}
	return ret
}