package output

import "fmt"

//Beautify
func Beautify() {
	banner1 := "                _     _     _ _ "
	banner2 := "  ___ __ _ _ __(_) __| | __| (_)"
	banner3 := " / __/ _` | '__| |/ _` |/ _` | |"
	banner4 := "| (_| (_| | |  | | (_| | (_| | |"
	banner5 := " \\___\\__,_|_|  |_|\\__,_|\\__,_|_| v1.0"

	fmt.Println(banner1)
	fmt.Println(banner2)
	fmt.Println(banner3)
	fmt.Println(banner4)
	fmt.Println(banner5)
	fmt.Println()
}
