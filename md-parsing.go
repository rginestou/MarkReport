package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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

// TOCEntry ...
type TOCEntry struct {
	Level  int
	Anchor string
	Title  string
}

var commentToHTML = map[string]string{
	"columns":  "<article class='columns'>\n",
	"!columns": "</article>\n",
	"items":    "<article class='items'>\n",
	"!items":   "</article>\n",
	"offers":   "<article class='offers'>\n",
	"!offers":  "</article>\n",
	"chapter":  "<article class='chapter'>\n",
	"!chapter": "</article>\n",
	"specs":    "<article class='specs'>\n",
	"!specs":   "</article>\n",
	"section":  "<section>\n",
	"!section": "</section>\n",
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
	inChapter := false
	coverHTML := ""
	titleAnchor := 0
	toc := make([]TOCEntry, 0)
	tocName := ""

	mdContent, _ := ioutil.ReadFile(dir + "/" + mdFile)
	ext := blackfriday.CommonExtensions & ^blackfriday.Autolink
	html := string(blackfriday.Run(mdContent, blackfriday.WithExtensions(ext)))

	htmlOut := ""
	scanner := bufio.NewScanner(strings.NewReader(html))
	re, _ := regexp.Compile(`<!--(.*)-->`)
	reGroup, _ := regexp.Compile(`(\w+) (.*)?`)
	reImg, _ := regexp.Compile(`<img src="([^\ ]+) =(\d*)?x(\d*)?`)
	reH, _ := regexp.Compile(`<h(\d)>(.*)</h\d>`)
	for scanner.Scan() {
		txt := scanner.Text()
		if !inCover {
			// Test for image
			res := reImg.FindAllStringSubmatch(txt, -1)
			if len(res) != 0 {
				width := res[0][2]
				height := ""
				if len(res) > 3 {
					height = res[0][3]
				}
				htmlOut += replaceImage(res[0][1], width, height) + "\n"
				continue
			}

			res = reH.FindAllStringSubmatch(txt, -1)
			if len(res) != 0 {
				anchor := "anchor" + strconv.Itoa(titleAnchor)
				htmlOut += txt[:3] + " id='" + anchor + "'" + txt[3:len(txt)] + "\n"
				i, _ := strconv.Atoi(res[0][1])
				if inChapter {
					i = 0
					inChapter = false
				}
				toc = append(toc, TOCEntry{i, anchor, res[0][2]})
				titleAnchor++
				continue
			}

			htmlOut += txt + "\n"
		} else {
			coverHTML += txt + "\n"
		}

		res := re.FindAllStringSubmatch(txt, -1)
		if len(res) == 0 {
			continue
		}

		comment := strings.TrimSpace(res[0][1])
		if val, ok := commentToHTML[comment]; ok {
			htmlOut += val
			if comment == "chapter" {
				inChapter = true
			}
			continue
		}

		if comment == "!cover" {
			coverHTML += "</article>\n"
			inCover = false
			coverHTML = strings.Replace(coverHTML, "<p>", "<address>\n", -1)
			coverHTML = strings.Replace(coverHTML, "</p>", "\n</address>", -1)
			coverHTML = strings.Replace(coverHTML, "\n\n", "\n", -1)
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
			coverHTML += "<article id='cover'>\n"
		} else if res[0][1] == "toc" {
			tocName = res[0][2]
		}
	}

	fmt.Println(toc)
	tocHTML := ""
	if tocName != "" {
		tocHTML += "<article id='contents'>\n"
		tocHTML += "<h2>" + tocName + "</h2>\n"
		level := 0
		for _, entry := range toc {
			if entry.Level > 2 {
				continue
			}

			if entry.Level > level {
				tocHTML += "<ul>"
				level = entry.Level
			}
			if entry.Level < level {
				tocHTML += "</ul>"
				level = entry.Level
			}

			if entry.Level == 0 {
				tocHTML += "<h3>" + entry.Title + "</h3>\n"
				continue
			}
			tocHTML += "<li><a href='#" + entry.Anchor + "'></a></li>\n"
		}
		tocHTML += "</article>"
	}

	htmlOut = "{{define \"content\"}}\n" + coverHTML + tocHTML + htmlOut + "{{end}}\n"

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

func replaceImage(src, width, height string) string {
	h := ""
	if height != "" {
		h = "height:" + height + "px"
	}
	return "<img src='" + src + "' style='width:" + width + "px;" + h + "'>"
}
