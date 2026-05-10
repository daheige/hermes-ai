package cohere

var ModelList = initModelList()

func initModelList() []string {
	s := []string{
		"command", "command-nightly",
		"command-light", "command-light-nightly",
		"command-r", "command-r-plus",
	}

	res := make([]string, 0, len(s))
	for k := range s {
		res = append(res, s[k]+"-internet")
	}

	s = append(s, res...)
	return s
}
