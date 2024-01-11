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

func remove(slice []string, element string) []string {
    var ret []string

    for _, v := range slice {
        if v != element {
            ret = append(ret, v)
        }
    }

    return ret
}