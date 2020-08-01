#!/usr/bin/env python3
import matplotlib.pyplot as plt
import os
import sys
import webbrowser

def save(fig, fname, fopen=False):
    fig.savefig(fname, bbox_inches="tight")
    if fopen:
        webbrowser.open_new("file://" + os.path.abspath(fname))

def init_ax(title, ax):
    ax.set_xlabel("Image Pointer")
    ax.set_ylabel("Levenshtein Distance")
    ax.set_title(title)

def plot_cer(fname, outf):
    data = [x.split(",") for x in open(fname, "r").read().strip().split("\n")]
    aws = [(int(x[1]), (int(x[1+1])/1000)) for x in data]
    azu = [(int(x[3]), (int(x[3+1])/1000)) for x in data]
    gcp = [(int(x[5]), (int(x[5+1])/1000)) for x in data]

    fig, (ax0, ax1, ax2) = plt.subplots(nrows=3, ncols=1, figsize=(8, 9), dpi=300)
    init_ax("AWS vs Azure", ax0)
    init_ax("Azure vs GCP", ax1)
    init_ax("GCP vs AWS", ax2)
    fig.suptitle("OCR Levenshtein Distances: %d Images" % len(data), size="xx-large", y=1.01)
    fig.tight_layout()
    x = list(range(len(aws)))
    ax0.scatter(x, [x[0] for x in aws], s=1, color="orange", label="AWS")
    ax1.scatter(x, [x[0] for x in azu], s=1, color="blue", label="Azure")
    ax2.scatter(x, [x[0] for x in gcp], s=1, color="red", label="GCP")
    save(fig, outf)
    plt.close()

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("usage: ./plot_lev.py combined.csv out.png")
        sys.exit(1)
    fname = sys.argv[1]
    outf = sys.argv[2]
    plot_cer(fname, outf)
