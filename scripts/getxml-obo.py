#!/usr/bin/env python3

from string import punctuation
import csv
import json
import os
import sys
import urllib.request
import xml.etree.ElementTree as ET
import zipfile

from tqdm import tqdm

IMAGE_URL = 'https://www.oldbaileyonline.org/images.jsp?doc='
#XML_SRC = 'https://figshare.shef.ac.uk/ndownloader/articles/4775434/versions/2'
XML_SRC = 'https://obo.cs.princeton.edu/data/xml.zip'
OA_PATCH_SRC = 'https://obo.cs.princeton.edu/data/patch/OA_PTRS_PATCH.csv'
SP_PATCH_SRC = 'https://obo.cs.princeton.edu/data/patch/SP_PTRS_PATCH.csv'
XML_PATH = 'OBO_XML_72'

OA = {}
OA['NAME'] = 'ordinarysAccounts'
OA['ENCODING'] = 'iso_8859_1'
OA['DATA'] = os.path.join('data', OA['NAME'])
OA['PATH'] = os.path.join(XML_PATH, OA['NAME'])
OA['PATCH_FILENAME'] = 'OA_PTRS_PATCH.csv'
OA['PTRS_FILENAME'] = 'xml_oa_ptrs.txt'
OA['PTRS'] = []
OA['404_FILENAME'] = 'xml_oa_ptrs_404.txt'
OA['404'] = []
OA['TEXT_FILENAME'] = 'xml_oa_text.json'
OA['TEXT'] = []

SP = {}
SP['NAME'] = 'sessionsPapers'
SP['ENCODING'] = 'utf-8'
SP['DATA'] = os.path.join('data', SP['NAME'])
SP['PATH'] = os.path.join(XML_PATH, SP['NAME'])
SP['PATCH_FILENAME'] = 'SP_PTRS_PATCH.csv'
SP['PTRS_FILENAME'] = 'xml_sp_ptrs.txt'
SP['PTRS'] = []
SP['404_FILENAME'] = 'xml_sp_ptrs_404.txt'
SP['404'] = []
SP['TEXT_FILENAME'] = 'xml_sp_text.json'
SP['TEXT'] = []

def xptr(root, i, accs, file):
    """
    Extract the xptr doc values from XML
    """
    if i <= 0:
        return accs
    for child in root:
        if child.tag == 'xptr'and child.attrib['type'] == 'pageFacsimile':
            accs += [child.attrib['doc']]
        accs = xptr(child, i-1, accs, file)
    return accs

def parse_pointers(path, files, cacheptrs, patchfile, override=False):
    # Open Patchfile
    if os.path.isfile(patchfile):
        with open(patchfile, 'r') as f:
            patch = [(e.strip(), c.strip()) for (e, c) in csv.reader(f)][1:]
    else:
        patch = []

    # Open cache
    if os.path.exists(cacheptrs) and not override:
        with open(cacheptrs, 'r') as f:
            return [x.strip() for x in f.readlines()]
    else:
        accs = []
        for file in tqdm(files):
            tree = ET.parse(os.path.join(path, file))
            root = tree.getroot()
            accs = xptr(root, 50, accs, file)
        # Apply patch to correct bad pointers
        for (error, correction) in patch:
            if error in accs:
                accs[accs.index(error)] = correction
            else:
                print('Failed to apply patch: %s' % error)
        accs.sort()
        with open(cacheptrs, 'w') as f:
            f.write('\n'.join(accs))
        return accs

def text(root, i, doc, accs):
    """
    Extract the text from XML
    """
    if i <= 0:
        return accs
    for child in root:
        if child.tag == 'xptr' and child.attrib['type'] == 'pageFacsimile':
            doc = child.attrib['doc']
            accs[doc] = ""
        if child.tag == 'p':
            accs[doc] += '\n'
        if child.text is not None:
            accs[doc] += ' ' + ' '.join(child.text.split()) + ' '
        (doc, accs) = text(child, i-1, doc, accs)
        if child.tail:
            if root.tag != 'p':
                accs[doc] += ' '
            elif len(accs[doc].strip()) and not accs[doc].strip()[-1].isspace():
                if len(child.tail.strip()) and child.tail.strip()[0] not in punctuation:
                    accs[doc] += ' '
            accs[doc] += ' '.join(child.tail.split()) + ' '
        accs[doc] = accs[doc].strip()
    return (doc, accs)

def parse_text(path, files, cachetext, patchfile, override=False):
    # Open Patchfile
    if os.path.isfile(patchfile):
        with open(patchfile, 'r') as f:
            patch = [(e.strip(), c.strip()) for (e, c) in csv.reader(f)][1:]
    else:
        patch = []

    if os.path.exists(cachetext) and not override:
        return json.load(open(cachetext, 'r'))
    else:
        acc = {}
        for file in tqdm(files):
            tree = ET.parse(os.path.join(path, file))
            root = tree.getroot()
            _, results = text(root, 50, "garbage", {"garbage": ""})
            del results["garbage"]
            for key, value in results.items():
                acc[key] = value
        # Cleanup data
        for (error, correction) in patch:
            if error in acc:
                acc[correction] = acc[error]
                del acc[error]
            else:
                print('Failed to apply patch: %s' % error)
        # Save to file
        json.dump(acc, open(cachetext, 'w'))
        return acc

def main():
    # Download datafiles if do not exist
    os.makedirs(SP['DATA'], exist_ok=True)
    os.makedirs(OA['DATA'], exist_ok=True)
    if not os.path.isdir(XML_PATH):
        print('[main]', 'Downloading XML...')
        urllib.request.urlretrieve(XML_SRC, XML_PATH + '.zip')
        print('[main]', 'Unzipping XML...')
        with zipfile.ZipFile(XML_PATH + '.zip', 'r') as zf:
            zf.extractall(XML_PATH)
    if not os.path.isfile(OA['PATCH_FILENAME']):
        print('[main]', 'Downloading OA Patch...')
        urllib.request.urlretrieve(OA_PATCH_SRC, OA['PATCH_FILENAME'])
    if not os.path.isfile(SP['PATCH_FILENAME']):
        print('[main]', 'Downloading SP Patch...')
        urllib.request.urlretrieve(SP_PATCH_SRC, SP['PATCH_FILENAME'])

    # Tabulate files
    OA['FILES'] = list(filter(lambda f: f.endswith('.xml'), os.listdir(OA['PATH'])))
    SP['FILES'] = list(filter(lambda f: f.endswith('.xml'), os.listdir(SP['PATH'])))

    # Parse the pointers
    print('[main]', 'Parsing pointers...')
    OA['PTRS'] = parse_pointers(OA['PATH'], OA['FILES'], OA['PTRS_FILENAME'], OA['PATCH_FILENAME'])
    print('[parse_pointers]', OA['NAME'], 'range:', '[%s, %s]' % (OA['PTRS'][0], OA['PTRS'][-1]), 'created:', OA['PTRS_FILENAME'], '(includes duplicates)')
    SP['PTRS'] = parse_pointers(SP['PATH'], SP['FILES'], SP['PTRS_FILENAME'], SP['PATCH_FILENAME'])
    print('[parse_pointers]', SP['NAME'], 'range:', '[%s, %s]' % (SP['PTRS'][0], SP['PTRS'][-1]), 'created:', SP['PTRS_FILENAME'], '(includes duplicates)')

    # Parses the plain text
    print('[main]', 'Parsing human transcriptions...')
    OA["TEXT"] = parse_text(OA['PATH'], OA['FILES'], OA['TEXT_FILENAME'], OA['PATCH_FILENAME'])
    print('[parse_text]', OA['NAME'], 'len:', len(OA['TEXT'].keys()), 'created:', OA['TEXT_FILENAME'])
    SP["TEXT"] = parse_text(SP['PATH'], SP['FILES'], SP['TEXT_FILENAME'], SP['PATCH_FILENAME'])
    print('[parse_text]', SP['NAME'], 'len:', len(SP['TEXT'].keys()), 'created:', SP['TEXT_FILENAME'])

    # Extract json into separate files
    print('[main]', 'Extracting OA txt to', OA['DATA'])
    for p in OA['TEXT']:
        with open(os.path.join(OA['DATA'], '%s.txt' % p), 'w') as f:
            f.write(OA['TEXT'][p])
    print('[main]', 'Extracting SP txt to', SP['DATA'])
    for p in tqdm(SP['TEXT']):
        with open(os.path.join(SP['DATA'], '%s.txt' % p), 'w') as f:
            f.write(SP['TEXT'][p])

    print('[main]', 'done.')

if __name__ == "__main__":
    main()
