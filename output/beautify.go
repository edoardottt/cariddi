package output

import (
	"fmt"

	"github.com/fatih/color"
)

//Beautify
func Beautify() {
	banner1 := "                 _     _     _ _ \n"
	banner2 := "   ___ __ _ _ __(_) __| | __| (_)\n"
	banner3 := "  / __/ _` | '__| |/ _` |/ _` | |\n"
	banner4 := " | (_| (_| | |  | | (_| | (_| | |\n"
	banner5 := "  \\___\\__,_|_|  |_|\\__,_|\\__,_|_| v1.0\n"
	banner6 := ""
	banner7 := " > github.com/edoardottt/cariddi\n"
	banner8 := " > edoardoottavianelli.it\n"
	banner9 := "========================================"

	bannerPart1 := banner1 + banner2 + banner3 + banner4 + banner5
	bannerPart2 := banner6 + banner7 + banner8 + banner9
	color.Cyan("%s\n", bannerPart1)
	fmt.Println(bannerPart2)
}
