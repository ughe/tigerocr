#!/usr/bin/env python3
from matplotlib.ticker import AutoMinorLocator
import matplotlib.pyplot as plt
import matplotlib as mpl
import os
import sys
import webbrowser

def save(fig, fname, fopen=True):
    fig.savefig(fname, bbox_inches="tight")
    if fopen:
        webbrowser.open_new("file://" + os.path.abspath(fname))
def acclogs(logs, service):
    return {x[0].split(".")[0]: x[1] for x in logs if service in x[0]}
def ceil(x, l):
    return x if x < l else l
def plot_cer(title, dirname, fout, ceil=lambda x: x):
    aws = dict(filter(lambda x: len(x)==2, [tuple(x.split(",")) for x in open(dirname + "/aws.txt", "r").read().split("\n")]))
    azu = dict(filter(lambda x: len(x)==2, [tuple(x.split(",")) for x in open(dirname + "/azure.txt", "r").read().split("\n")]))
    gcp = dict(filter(lambda x: len(x)==2, [tuple(x.split(",")) for x in open(dirname + "/gcp.txt", "r").read().split("\n")]))
    # Set of similar results
    x = list(set(aws.keys()) & set(azu.keys()) & set(gcp.keys()))
    print('len(x)', len(x))
    awsy = [ceil(float(v)) for k, v in aws.items() if k in x]
    azuy = [ceil(float(v)) for k, v in azu.items() if k in x]
    gcpy = [ceil(float(v)) for k, v in gcp.items() if k in x]

    fig = plt.figure(figsize=(6, 6), dpi=300)
    fig.suptitle(title, size="xx-large")
    fig.subplots_adjust(hspace=1)
    fig.tight_layout()
    rx = list(range(len(x)))
    ax1 = fig.add_subplot(311)
    ax1.set_title("AWS")
    ax1.yaxis.set_minor_locator(AutoMinorLocator(5))
    ax1.xaxis.set_minor_locator(AutoMinorLocator(5))
    ax1.set_xlabel("Page Image Pointer")
    ax1.set_ylabel("Character Error Rate")
    ax1.scatter(rx, awsy, s=.4, color="orange", label="AWS")

    ax2 = fig.add_subplot(312)
    ax2.set_title("Azure")
    ax2.yaxis.set_minor_locator(AutoMinorLocator(5))
    ax2.xaxis.set_minor_locator(AutoMinorLocator(5))
    ax2.set_xlabel("Page Image Pointer")
    ax2.set_ylabel("Character Error Rate")
    ax2.scatter(rx, azuy, s=.4, color="blue", label="Azure")

    ax3 = fig.add_subplot(313)
    ax3.set_title("GCP")
    ax3.yaxis.set_minor_locator(AutoMinorLocator(5))
    ax3.xaxis.set_minor_locator(AutoMinorLocator(5))
    ax3.set_xlabel("Page Image Pointer")
    ax3.set_ylabel("Character Error Rate")
    ax3.scatter(rx, gcpy, s=.4, color="red", label="GCP")
    #lgnd = ax2.legend(bbox_to_anchor=(0.75, -0.2), ncol=3)
    #lgnd.legendHandles[0]._sizes = [30]
    #lgnd.legendHandles[1]._sizes = [30]
    #lgnd.legendHandles[2]._sizes = [30]
    save(fig, fout)
    plt.close()

if __name__ == "__main__":
    if len(sys.argv) < 4 or len(sys.argv) > 5:
        print("usage: plot_cer3.py OA oa-lev out.png [1.0]")
        print("       lev dir must contain: aws.txt azure.txt gcp.txt")
        sys.exit(1)
    name = sys.argv[1]
    logf = sys.argv[2]
    dst = sys.argv[3]
    ce = lambda x: x
    if len(sys.argv) == 5:
        c = sys.argv[4]
        ce = lambda x: ceil(x, float(c))
    plot_cer("OCR CER: " + name, logf, dst, ce)
