package main

import (
	"fmt"
	"time"
	"bufio"
	"html"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	blackfriday "gopkg.in/russross/blackfriday.v2"
	"gopkg.in/yaml.v2"
)

// Data ...
type Data struct {
	Title    string
	Subtitle string
	Cover    string
	Header   bool
}

// TOCEntry ...
type TOCEntry struct {
	Level  int
	Anchor string
	Title  string
}

var commentToHTML = map[string]string{
	"text":          "<article class='text'>\n",
	"!text":         "</article>\n",
	"columns":       "<article class='columns'>\n",
	"!columns":      "</article>\n",
	"items":         "<article class='items'>\n",
	"!items":        "</article>\n",
	"offers":        "<article class='offers'>\n",
	"!offers":       "</article>\n",
	"chapter":       "<article class='chapter'>\n",
	"!chapter":      "</article>\n",
	"specs":         "<article class='specs'>\n",
	"!specs":        "</article>\n",
	"section":       "<section>\n",
	"!section":      "</section>\n",
	"section-bold":  "<section class='bold'>\n",
	"!section-bold": "</section>\n",
}

var figNum = 1

// split front matter from markdown (see jekyll)
func splitMarkDownFrontMatter(input string) (markdown string, front_matter *Data) {
	re, _ := regexp.Compile(`---`)
	res := re.FindAllStringSubmatchIndex(input, -1)

	// must contain at least two occurances of ---
	if len(res) < 2 {
		return input, nil
	}

	// it must be on the begining of the document
	if res[0][0] != 0 {
		return input, nil
	}
	front := input[res[0][1]:res[1][0]]

	yamlData := Data{}
	yamlData.Header = true // default value

	err := yaml.Unmarshal([]byte(front), &yamlData)
	if err != nil {
		return input, nil
	}

	md := input[res[1][1]:]
	return md, &yamlData
}

func getMarkdownContent(dir string) ([]byte, *Data) {
	d, _ := os.Open(dir)
	files, _ := d.Readdir(-1)
	d.Close()

	// List md files
	mdFiles := make(map[string]bool)
	for _, file := range files {
		mdFile := file.Name()
		if file.Mode().IsRegular() && filepath.Ext(mdFile) == ".md" {
			mdFiles[mdFile] = true
		}
	}

	if len(mdFiles) == 0 {
		return []byte{}, nil
	}

	// Look for content.txt
	mdFilesPicked := []string{}
	for _, file := range files {
		if file.Mode().IsRegular() && file.Name() == "content.txt" {
			f, _ := os.Open(dir + "/" + file.Name())

			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				mdFilesPicked = append(mdFilesPicked, scanner.Text()+".md")
			}

			f.Close()
			break
		}
	}

	// Choose md files
	if len(mdFilesPicked) == 0 {
		for f := range mdFiles {
			mdFilesPicked = append(mdFilesPicked, f)
			break
		}
	}

	mdContent := ""
	var data *Data

	reMDImg, _ := regexp.Compile(`!\[\]\(([^\ #)]*)[^\)]*\)`)

	for _, f := range mdFilesPicked {
		if (f == ".md") {
			continue
		}
		c, err := ioutil.ReadFile(dir + "/" + f)
		if err != nil {
			panic(err)
		}
		cMD, cFrontMatter := splitMarkDownFrontMatter(string(c))
		if cFrontMatter != nil {
			data = cFrontMatter
		}

		relpath := dir + "/" + filepath.Dir(f)

		res := reMDImg.FindAllStringSubmatch(cMD, -1)
		indexes := reMDImg.FindAllStringSubmatchIndex(cMD, -1)

		indexChange := 0
		for i, item := range res {
			index := indexes[i]
			imgFilename := item[1]
			destFilename := relpath + "/" + imgFilename
			startIndex := index[2] + indexChange
			endIndex := index[3] + indexChange
			cMD = cMD[:startIndex] + destFilename + cMD[endIndex:]
			indexChange = indexChange + len(destFilename) - (index[3] - index[2])
			if _, err := os.Stat(destFilename); os.IsNotExist(err) {
				err_msg := fmt.Sprintf("file \"%v\" does not exist\n", destFilename)
				panic(err_msg)
			}

		}

		mdContent += "\n\n" + cMD
	}

	return []byte(mdContent), data
}

func main() {

	if (len(os.Args) != 2) {
		print("usage: " + os.Args[0] + " directory\n")
		return
	}
	dir := os.Args[1]

	var data Data
	data.Header = true

	mdContent, yamlData := getMarkdownContent(dir)

	if yamlData != nil {
		data = *yamlData
	}

//	fmt.Printf("data: %v\n\n", data)

	inCover := false
	inChapter := false
	coverHTML := ""
	titleAnchor := 0
	toc := make([]TOCEntry, 0)
	tocName := ""

	ext := blackfriday.CommonExtensions & ^blackfriday.Autolink
	htmlStr := string(blackfriday.Run(mdContent, blackfriday.WithExtensions(ext)))

	htmlOut := ""
	scanner := bufio.NewScanner(strings.NewReader(htmlStr))
	re, _ := regexp.Compile(`<!--([^>]*)-->`)
	reGroup, _ := regexp.Compile(`(\w+) (.*)?`)
	reImg, _ := regexp.Compile(`<img[^>]*src="([^#\"]+)(#?[^"]*)?(?: =(\d*)?x(\d*)?)?"[^>]*(?: alt="(.+)")?[^>]*>`)


	reH, _ := regexp.Compile(`<h(\d)>(.*)</h\d>`)
	for scanner.Scan() {
		txt := scanner.Text()
		txt = strings.Replace(txt, "&amp;nbsp;", "&nbsp;", -1)
		if !inCover {

			// Test for image
			res := reImg.FindAllStringSubmatch(txt, -1)
			for _, i := range res {
				filename := i[1]
				filenameAnchor := i[2]
				width := i[3]
				height := i[4]
				alt := html.UnescapeString(i[5])
				htmlOut += replaceImage(filename+filenameAnchor, width, height, alt) + "\n"
			}
			if len(res) != 0 {
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

		if comment == "!BUILD_DATETIME" {
			coverHTML = strings.Replace(coverHTML, "<!-- !BUILD_DATETIME -->", "{{template \"BUILD_DATETIME\" }}", -1)
		}

		if comment == "!BUILD_DATE" {
			coverHTML = strings.Replace(coverHTML, "<!-- !BUILD_DATE -->", "{{template \"BUILD_DATE\" }}", -1)
		}

		if comment == "!BUILD_VERSION" {
			coverHTML = strings.Replace(coverHTML, "<!-- !BUILD_VERSION -->", "{{template \"BUILD_VERSION\" }}", -1)
		}

		if comment == "!cover" {
			coverHTML += "</article>\n"
			inCover = false
			coverHTML = strings.Replace(coverHTML, "<p>", "<address>\n", -1)
			coverHTML = strings.Replace(coverHTML, "</p>", "\n</address>", -1)
			coverHTML = strings.Replace(coverHTML, "\n\n", "\n", -1)
			continue
		}

		if comment == "no-header" {
			data.Header = false
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
	htmlOut = htmlOut + "{{define \"BUILD_DATETIME\"}}" + time.Now().Format(time.RFC3339) + " {{end}}\n"
	htmlOut = htmlOut + "{{define \"BUILD_DATE\"}}" + time.Now().Format("2006-01-02") + " {{end}}\n"
	htmlOut = htmlOut + "{{define \"BUILD_VERSION\"}} " + os.Getenv("BUILD_VERSION") + " {{end}}\n"


	// Makdown HTML
	f, _ := os.Create(dir + "/md-output.html")
	w := bufio.NewWriter(f)
	w.WriteString(htmlOut)
	w.Flush()
	f.Close()

	f, _ = os.Create(dir + "/output.html")
	w = bufio.NewWriter(f)
	t, err := template.ParseFiles(dir+"/base.html", dir+"/md-output.html")
	if err != nil {
		panic(err)
	}
	t.ExecuteTemplate(w, "base", data)
	w.Flush()
	f.Close()
}

func replaceImage(src, width, height, alt string) string {
	h := ""
	if height != "" {
		h = "height:" + height + "%"
	}
	str := "<div style='margin: 0 auto;width:" + width + "%;" + h + "'>"
	str += "<img src='" + src + "' style='width:100%'>"
	if alt != "" {
		str += "<p style='text-align:center'><b>Figure " + strconv.Itoa(figNum) + "</b> " + alt + "</p>"
		figNum++
	}
	str += "</div>"
	return str
}
