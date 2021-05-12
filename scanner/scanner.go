package scanner

//Scan
func Scan() {

}

//isEmailUrl
func isEmailUrl(inp string) bool {
	return inp[:7] == "mailto:"
}

//isFtpUrl
func isFtpUrl(inp string) bool {
	return inp[:4] == "ftp:"
}

//isHttpUrl
func isHttpUrl(inp string) bool {
	return inp[:5] == "http:"
}

//isHttpsUrl
func isHttpsUrl(inp string) bool {
	return inp[:6] == "https:"
}
