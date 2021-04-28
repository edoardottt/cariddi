package scanner

//Secret
type Secret struct {
	name        string
	description string
	regex       string
	poc         string
}

var regexes = map[string]Secret{}
