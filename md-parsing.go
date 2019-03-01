package main

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	blackfriday "gopkg.in/russross/blackfriday.v2"
)

// Data ...
type Data struct {
	Title    string
	Subtitle string
	Cover    string
}

var commentToHTML = map[string]string{
	"columns":     "<article class='columns'>\n",
	"!columns":    "</article>\n",
	"items":       "<article class='items'>\n",
	"!items":      "</article>\n",
	"offers":      "<article class='offers'>\n",
	"!offers":     "</article>\n",
	"chapter":     "<article class='chapter'>\n",
	"!chapter":    "</article>\n",
	"typography":  "<article class='typography'>\n",
	"!typography": "</article>\n",
	"section":     "<section>\n",
	"!section":    "</section>\n",
}

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

	var data Data
	inCover := false

	mdContent, _ := ioutil.ReadFile(dir + "/" + mdFile)
	html := string(blackfriday.Run(mdContent))

	htmlOut := "{{define \"content\"}}"
	scanner := bufio.NewScanner(strings.NewReader(html))
	re, _ := regexp.Compile(`<!--(.*)-->`)
	reGroup, _ := regexp.Compile(`(\w+) (.*)?`)
	for scanner.Scan() {
		txt := scanner.Text()
		if !inCover {
			htmlOut += txt + "\n"
		} else if len(txt) > 4 {
			htmlOut += txt[:3] + " class='cover'" + txt[3:len(txt)] + "\n"
		}

		res := re.FindAllStringSubmatch(txt, -1)
		if len(res) == 0 {
			continue
		}

		comment := strings.TrimSpace(res[0][1])
		if val, ok := commentToHTML[comment]; ok {
			htmlOut += val
			continue

		}

		if comment == "!cover" {
			htmlOut += "</article>\n"
			inCover = false
			continue
		}

		res = reGroup.FindAllStringSubmatch(comment, -1)
		if len(res) == 0 {
			continue
		}

		if res[0][1] == "title" {
			data.Title = res[0][2]
		} else if res[0][1] == "subtitle" {
			data.Subtitle = res[0][2]
		} else if res[0][1] == "cover" {
			data.Cover = res[0][2]
			inCover = true
			htmlOut += "<article id='cover'>\n"
			// htmlOut += "<h1>Salut</h1>\n"
		}
	}
	htmlOut += "{{end}}"

	// Makdown HTML
	f, _ := os.Create(dir + "/md-output.html")
	w := bufio.NewWriter(f)
	w.WriteString(htmlOut)
	w.Flush()
	f.Close()

	f, _ = os.Create(dir + "/output.html")
	w = bufio.NewWriter(f)
	t, _ := template.ParseFiles(dir+"/base.html", dir+"/md-output.html")
	t.ExecuteTemplate(w, "base", data)
	w.Flush()
	f.Close()
}
