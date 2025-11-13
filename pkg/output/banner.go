/*
==========
Cariddi
==========

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see http://www.gnu.org/licenses/.

	@Repository:  https://github.com/edoardottt/cariddi

	@Author:      edoardottt, https://www.edoardottt.com

	@License: https://github.com/edoardottt/cariddi/blob/main/LICENSE

*/

package output

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

const (
	version = "v1.4.4"
	banner  = `                 _     _     _ _ 
                (_)   | |   | (_)
   ___ __ _ _ __ _  __| | __| |_ 
  / __/ _` + "`" + ` | '__| |/ _` + "`" + ` |/ _` + "`" + ` | |
 | (_| (_| | |  | | (_| | (_| | |
  \___\__,_|_|  |_|\__,_|\__,_|_| `
)

// Banner prints the banner + version.
func Banner() {
	links := " > github.com/edoardottt/cariddi\n > https://edoardottt.com/\n"
	sepLine := "========================================\n"

	bannerPart1 := banner + version + "\n\n"
	bannerPart2 := links + sepLine

	color.Set(color.FgCyan)
	fmt.Fprint(os.Stderr, bannerPart1)
	color.Unset()
	fmt.Fprint(os.Stderr, bannerPart2)
}
