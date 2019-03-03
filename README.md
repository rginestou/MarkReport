# MarkReport

This little script seamlessly converts Mardown to elegant PDF reports.

![](doc/cover.png)

* Configurable report layout
* Automatic Table Of Content
* LaTeX equations
* Syntax highlighting
* Resizable images

## How does it work?

_MarkReport_ takes a markdown file, converts it to HTML (using Go's excellent [blackfriday](https://github.com/russross/blackfriday) package), renders LaTeX equations and code highlighting with JavaScript thanks to [Selenium](https://github.com/SeleniumHQ/selenium) and finally converts the enriched HTML to PDF thanks to [WeasyPrint](https://weasyprint.org/).

## How to use it?

Just type in your document as a Markdown `.md` file, using special syntax in comments to tell MarkReport how exactly the final PDF should be structured.

### Simple example

Let's take the following example:

```md
<!-- title Test Report -->

## This is a title

### This is a subtitle

<!-- section -->

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aliquam volutpat faucibus vestibulum.
Mauris varius orci quam. Nam dui mauris, dictum at elementum at, mollis pulvinar est.
Nunc lobortis pharetra erat, id rutrum lorem malesuada in.

<!-- !section -->
```

Open the folder in which the Markdown file above is located. Then run _MarkReport_:

    cd /path/to/markdown/file
    /path/to/MarkReport/MarkReport.py

The build process completes after a few seconds, and a `output.pdf` file appears in the folder.

![](doc/markreport-example.png)

See the `example` folder for a more detailed demonstration of what _MarkReport_ can achieve.

### Available command line flags

* `--basic` Javascript interpreter is disabled, allowing faster builds but without syntax highlighting or LaTeX support
* `--watch` The _MarkReport_ script will not stop after the first build, but stay idle and will rebuild t
as soon as a change is made in the current folder, allowing for faster hot-builds
* `--quiet` No output will be displayed during the build process

## Installation instructions

Some Python packages are needed to run the program. It's easy to get them with pip3:

    pip3 install weasyprint
    pip3 install pyinotify
    pip3 install selenium

The firefow driver is used to interpret JavaScript inside the HTML page generated from Markdown. You need to grab `geckodriver` in order to make it work:

    wget https://github.com/mozilla/geckodriver/releases/download/v0.24.0/geckodriver-v0.24.0-linux64.tar.gz
    tar -xvzf geckodriver*
    sudo mv geckodriver /usr/local/bin/

You're now reeady to go.
