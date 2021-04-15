#!/usr/bin/env python3

from string import punctuation
from html.parser import HTMLParser
import csv
import json
import os
import sys
import urllib.request
import zipfile

from tqdm import tqdm

JUSTIA_URL='https://supreme.justia.com/cases/federal/us/%s/%s/'
OUTPUT_TXT = 'justia-txt'
OUTPUT_HTML = 'html'
TAG_PREFIX = 'tab-opinion-'

ACTUAL_TEXT = ["ra", "k", "h"] # i.e. a bond marked <H>
WILD_TAG = "l=|" # i.e. l=|<a or l=|255 real world examples
IGNORE_TAGS = ["span", "br", "ul", "font", "symbol", "s"]
SLOPPY_TAGS = ["p", "a", "b"] # Okay to not balance (usually an extra closing tag

EXCLUDE_CASES = ["010us008.html"] # Excluded because contains 1 plaintext table and 1 html table

# At the end, each tag in tags should have value of 0 (every tag should be closed)
class ParsePrefix(HTMLParser):
    def __init__(self, default_pagenum):
        super().__init__()
        self.tags = {"div":0, "strong":0, "p":0, "em":0, "h2":0, "h3":0, "a":0, "b":0}
        self.relevant = False
        self.relevantLevel = -1
        self.in_p = False
        self.first_write_in_p = False
        self.in_pagenum = False
        self.default_pagenum = default_pagenum
        self.current_pagenum = default_pagenum
        self.records = []
        self.in_footnote = False
        self.in_footnote_erase = False # used to stop writing when in footnote
        self.read_page_number = False

    def handle_starttag(self, tag, attrs):
        if self.relevant and tag in ACTUAL_TEXT: # A tag that was intended to be text! See 139us337-002
            self.records[-1][self.current_pagenum] += "<%s>" % tag
            return
        if self.relevant and tag not in self.tags and tag not in IGNORE_TAGS and tag[:3] != WILD_TAG:
            raise Exception("Unexpected tag: %s" % tag)
        keys, vals = [x[0] for x in attrs], [x[1] for x in attrs]
        if self.relevant and tag not in IGNORE_TAGS and tag[:3] != WILD_TAG:
            self.tags[tag] += 1
        elif tag == "div" and "id" in keys:
            # Start recording relevant information if tag prefix matched
            v = vals[keys.index("id")]
            if v[:len(TAG_PREFIX)] == TAG_PREFIX:
                self.relevant = True
                self.relevantLevel = self.tags[tag]
                self.tags[tag] += 1
                self.records.append({self.default_pagenum: ""})
                self.current_pagenum = self.default_pagenum
                self.in_footnote = False
                self.in_footnote_erase = False # used to stop writing when in footnote
        if self.relevant and tag == "p":
            self.in_p = True
            self.first_write_in_p = True
        if self.relevant and tag == "a" and "class" in keys and vals[keys.index("class")] == "page-number":
            if "name" in keys:
                self.current_pagenum = vals[keys.index("name")]
                self.records[-1][self.current_pagenum] = ""
                self.in_pagenum = True
            else:
                self.in_pagenum = True
                # Must try to read the page number!!!
                # i.e. <a class="page-number">Page 10 U. S. 207
                self.read_page_number = True
                self.read_page_number_level = self.tags[tag]

    def handle_endtag(self, tag):
        if self.relevant and tag not in self.tags and tag not in IGNORE_TAGS and tag[:3] != WILD_TAG:
            raise Exception("Unexpected tag: %s" % tag)
        if self.relevant and tag == "ul":
            self.in_footnote = True
        if self.relevant and tag not in IGNORE_TAGS and tag[:3] != WILD_TAG:
            self.tags[tag] -= 1
        if tag == "p" and self.read_page_number and self.read_page_number_level < self.tags[tag]:
            # Error if no page number was set
            raise Exception("Expected a page number to be set")
        if self.in_pagenum and self.relevant and tag == "a": # </a>
            self.in_pagenum = False
        if self.in_p and self.relevant and tag == "p": # </p>
            self.in_p = False
        if self.relevant and tag == "div" and self.tags[tag] == self.relevantLevel: # </div>
            self.in_p = False
            self.in_pagenum = False
            self.relevant = False

    def handle_data(self, data):
        if self.relevant and self.read_page_number and self.tags["p"] == self.read_page_number_level:
            t = data.replace(" ", "")
            delim = "U.S."
            if delim not in t:
                delim = ","
            if delim not in t:
                return # Aborted attempt to set page number
            # Update to latest page
            self.current_pagenum = t.split(delim)[1]
            self.records[-1][self.current_pagenum] = ""
            self.read_page_number = False
        elif self.relevant and self.in_p and not self.in_pagenum:
            text = data.replace("\r", "")
            if text != "":
                # Add newlines at start of new <p> tags
                if self.first_write_in_p:
                    if self.records[-1][self.current_pagenum] != "":
                        self.records[-1][self.current_pagenum] += "\n"
                    self.first_write_in_p = False
                if self.in_footnote:
                    if text[0] == "[" or self.in_footnote_erase: # Remove brackets
                        if ']' in text:
                            text = text[text.index(']'):]
                            self.in_footnote_erase = False
                        else:
                            return # Ignore whole string if bracket started
                self.records[-1][self.current_pagenum] += text

def main():
    if len(sys.argv) != 2:
        print("usage: %s ptrs.txt" % sys.argv[0], file=sys.stderr)
        print("usage: %s html/" % sys.argv[0], file=sys.stderr)
        sys.exit(1)

    # Download html files if they do not exist
    htmldir = sys.argv[1]
    if not os.path.isdir(htmldir):
        htmldir = OUTPUT_HTML
        os.makedirs(htmldir)
        # Download datafiles
        ptrfile = sys.argv[1]
        with open(ptrfile, 'r') as f:
            ptrs = f.readlines()
        for ptr in tqdm(ptrs):
            if ptr[3:5] != "us":
                print("invalid ptr: %s" % ptr, file=sys.stderr)
            try:
                u = JUSTIA_URL % (ptr[0:3], ptr.strip()[5:])
                urllib.request.urlretrieve(u, os.path.join(htmldir, ptr.strip()+".html"))
            except urllib.error.HTTPError:
                print("failed: " + u, file=sys.stderr)

    os.makedirs(OUTPUT_TXT)

    # Tabulate files
    files = list(filter(lambda f: f.endswith('.html'), os.listdir(htmldir)))

    for ex in EXCLUDE_CASES:
        if ex in files:
            files.remove(ex)

    print('[main]', 'Parsing html...')
    for f in tqdm(files):
        with open(os.path.join(htmldir, f), 'r') as fd:
            buf = fd.read()
        prefix, suffix = f.strip(".html").split("us")
        p = ParsePrefix(suffix)
        p.feed(buf)
        if len(p.records) != 2:
            raise Exception("Expected two tags: tab-opinion-... but found %d in %s" % (len(p.records), f))
        # The first is the syllabus (a duplicate). Only look at the second
        pages = []
        for k in p.records[-1].keys():
            name = prefix + "us" + suffix + "-%03d" % (int(k) - int(suffix))
            with open(os.path.join(OUTPUT_TXT, '%s.txt' % name), 'w') as fdo:
                fdo.write(p.records[-1][k])
        for tag in p.tags:
            if tag not in IGNORE_TAGS and tag[:3] != WILD_TAG and p.tags[tag] != 0: # Exception for br and span
                if tag in SLOPPY_TAGS: # Exception for sloppy tags
                    #print("warning: extra %s in %s" % (tag, f), file=sys.stderr)
                    pass
                else:
                    raise Exception("Tag is not closed %s: %s (%d)" % (f, tag, p.tags[tag]))

    print('[main]', 'done.')

if __name__ == "__main__":
    main()
