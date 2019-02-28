package main

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	blackfriday "gopkg.in/russross/blackfriday.v2"
)

func main() {
	dir := os.Args[1]
	d, _ := os.Open(dir)
	files, _ := d.Readdir(-1)
	d.Close()

	mdFile := ""
	for _, file := range files {
		mdFile = file.Name()
		if file.Mode().IsRegular() && filepath.Ext(mdFile) == ".md" {
			break
		}
	}

	mdContent, _ := ioutil.ReadFile(dir + "/" + mdFile)
	html := string(blackfriday.Run(mdContent))

	htmlOut := ""
	scanner := bufio.NewScanner(strings.NewReader(html))
	re, _ := regexp.Compile(`<!--(.*)-->`)
	for scanner.Scan() {
		txt := scanner.Text()
		htmlOut += txt + "\n"

		res := re.FindAllStringSubmatch(txt, -1)
		if len(res) == 0 {
			continue
		}

		comment := strings.TrimSpace(res[0][1])
		if comment == "columns" {
			htmlOut += "<article class='columns'>\n"
		} else if comment == "!columns" {
			htmlOut += "</article>\n"
		} else if comment == "items" {
			htmlOut += "<article class='items'>\n"
		} else if comment == "!items" {
			htmlOut += "</article>\n"
		} else if comment == "offers" {
			htmlOut += "<article class='offers'>\n"
		} else if comment == "!offers" {
			htmlOut += "</article>\n"
		} else if comment == "chapter" {
			htmlOut += "<article class='chapter'>\n"
		} else if comment == "!chapter" {
			htmlOut += "</article>\n"
		} else if comment == "typography" {
			htmlOut += "<article class='typography'>\n"
		} else if comment == "!typography" {
			htmlOut += "</article>\n"
		} else if comment == "section" {
			htmlOut += "<section>\n"
		} else if comment == "!section" {
			htmlOut += "</section>\n"
		}
	}

	f, _ := os.Create(dir + "/md-output.html")
	defer f.Close()

	w := bufio.NewWriter(f)
	w.WriteString(htmlOut)
	w.Flush()
}
