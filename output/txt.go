package output

import (
	"log"
	"os"
)

//AppendOutputToTxt
func AppendOutputToTxt(output string, filename string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	if _, err := file.WriteString(output + "\n"); err != nil {
		log.Fatal(err)
	}
	file.Close()
}
