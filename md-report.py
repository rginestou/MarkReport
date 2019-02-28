#!/usr/bin/env python3

from weasyprint import HTML
from markdown2 import Markdown
from mako.template import Template
from mako.lookup import TemplateLookup

from selenium import webdriver
from selenium.webdriver.firefox.options import Options

import pyinotify

from distutils.dir_util import copy_tree
from tempfile import gettempdir
from time import time
from sys import stdout
import glob, os
import re

# Temp dir

timestamp = str(int(time()))
timestamp = "1111"
tmp_dir = gettempdir() + "/" + timestamp + "_md-report/"
os.makedirs(tmp_dir, exist_ok=True)


# Headless browser

options = Options()
options.headless = True
driver = webdriver.Firefox(options=options)
driver.set_page_load_timeout(2)

prev_compile_time = 0
def recompile(notifier):
    if notifier is not None and notifier.maskname != "IN_MODIFY":
        return
    global prev_compile_time
    if time() - prev_compile_time < 1:
        return
    prev_compile_time = time()

    stdout.write("\rBuilding the PDF file...")
    stdout.flush()

    copy_tree("report", tmp_dir)
    copy_tree("example", tmp_dir)

    # Base HTML Template

    base_html = ""
    with open(tmp_dir + "base.html", "r") as base_html_file:
        base_html = base_html_file.read()

    # Markdown parsing

    md = ""
    md_file_name = glob.glob(tmp_dir + "*.md")[0]
    with open(md_file_name, "r") as md_file:
        md = md_file.readlines()

    os.system("./md-parsing " + tmp_dir)

    md_html = ""
    with open(tmp_dir + "/md-output.html", "r") as md_html_file:
        md_html = md_html_file.read()

    # Create HTML file

    lookup = TemplateLookup()
    lookup.put_string("base.html", base_html)
    lookup.put_string("index.html", "<%inherit file='base.html'/>\n" + md_html)
    index_template = lookup.get_template("index.html")
    html = index_template.render()

    html_file_name = tmp_dir + "output.html"
    with open(html_file_name, "w") as html_out_file:
        html_out_file.write(html)

    # Interpret JS code

    driver.get("file:///" + tmp_dir + "output.html")
    elem = driver.find_element_by_xpath("//*")
    interpreted_html = elem.get_attribute("outerHTML")

    with open(html_file_name, "w") as html_out_file:
        html_out_file.write(interpreted_html)

    # Create final PDF file

    pdf = HTML(html_file_name).write_pdf()
    f = open("output.pdf",'wb')
    f.write(pdf)

    stdout.write("\rDone.                   ")
    stdout.flush()

recompile(None)

watch_manager = pyinotify.WatchManager()
event_notifier = pyinotify.Notifier(watch_manager, recompile)

watch_manager.add_watch(os.path.abspath("."), pyinotify.ALL_EVENTS)
event_notifier.loop()
driver.quit()
