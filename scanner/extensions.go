package scanner

//FileType
type FileType struct {
	extension string
	severity  int
	alert     bool
}

func GetExtensions() []FileType {
	//extensions contains a list of known extensions
	//and the TYPICAL (also say `in general`) associated severity.
	var extensions = []FileType{
		{"key", 1, true},
		{"env", 1, true},
		{"pem", 1, true},
		{"git", 1, true},
		{"ovpn", 1, true},
		{"log", 1, true},
		{"secret", 1, true},
		{"secrets", 1, true},
		{"bak", 1, true},
		{"dat", 1, true},
		{"db", 1, true},
		{"sh", 2, true},
		{"py", 2, true},
		{"json", 2, true},
		{"xml", 2, true},
		{"yml", 2, true},
		{"yaml", 2, true},
		{"properties", 2, true},
		{"toml", 2, true},
		{"dtd", 2, true},
		{"php4", 2, true},
		{"conf", 2, true},
		{"zip", 3, true},
		{"doc", 3, false},
		{"docx", 3, false},
		{"dochtml", 3, false},
		{"csv", 3, false},
		{"odt", 3, false},
		{"xls", 3, false},
		{"xlsx", 3, false},
		{"txt", 3, false},
		{"ts", 4, false},
		{"js", 4, false},
		{"php", 5, false},
		{"phtml", 5, false},
		{"php5", 5, false},
		{"html", 6, false},
		{"htm", 6, false},
		{"pdf", 7, false},
	}

	return extensions
}
