package output

import (
	"log"
	"os"
)

//bannerHTML
func bannerHTML(target string, filename string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	file.WriteString("<html><body><div style='" + "background-color:#4adeff;color:white" + "'><h1>Cariddi</h1>")
	file.WriteString("<ul>")
	file.WriteString("<li><a href='" + "https://github.com/edoardottt/cariddi'" + ">github.com/edoardottt/cariddi</a></li>")
	file.WriteString("<li>edoardottt, <a href='" + "https://www.edoardoottavianelli.it'" + ">edoardoottavianelli.it</a></li>")
	file.WriteString("<li>Released under <a href='" + "http://www.gnu.org/licenses/gpl-3.0.html'" + ">GPLv3 License</a></li></ul></div>")
	file.WriteString("<h4>target: " + target + "</h4>")
	file.Close()
}

//appendOutputToHtml
func appendOutputToHTML(output string, status string, filename string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	var statusColor string
	if status != "" {
		if string(status[0]) == "2" || string(status[0]) == "3" {
			statusColor = "<p style='color:green;display:inline'>" + status + "</p>"
		} else {
			statusColor = "<p style='color:red;display:inline'>" + status + "</p>"
		}
	} else {
		statusColor = status
	}
	if _, err := file.WriteString("<li><a target='_blank' href='" + output + "'>" + output + "</a> " + statusColor + "</li>"); err != nil {
		log.Fatal(err)
	}
	file.Close()
}

//headerHtml
func headerHTML(header string, filename string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	if _, err := file.WriteString("<h3>" + header + "</h3><ul>"); err != nil {
		log.Fatal(err)
	}
	file.Close()
}

//footerHTML
func footerHTML(filename string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	if _, err := file.WriteString("</ul>"); err != nil {
		log.Fatal(err)
	}
	file.Close()
}

//bannerFooterHTML
func bannerFooterHTML(filename string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	file.WriteString("<div style='" + "background-color:#4adeff;color:white" + "'>")
	file.WriteString("<ul><li><a href='" + "https://github.com/edoardottt/cariddi'" + ">Contribute to cariddi</a></li>")
	file.WriteString("<li>Released under <a href='" + "http://www.gnu.org/licenses/gpl-3.0.html'" + ">GPLv3 License</a></li></ul></div>")
	file.Close()
}
