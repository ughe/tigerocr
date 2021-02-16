#!/usr/bin/env python3
import matplotlib.pyplot as plt
import os
import sys
import webbrowser

# BEGIN configuration
providers = ["aws", "azu", "gcp"]
metrics = ["cer", "millis"]
convert = [
        lambda x: min(max(float(x), 0), 1),
        lambda x: min(max(int(x)/1000, 0), 12),
]
providersAxis = ["AWS", "Azure", "GCP"]
metricsAxis = ["CER", "Seconds"]
# END configuration

def save(fig, fname, fopen=False):
    fig.savefig(fname, bbox_inches="tight")
    if fopen:
        webbrowser.open_new("file://" + os.path.abspath(fname))

def init_ax(title, ax):
    ax.set_xlabel("Image Pointer")
    ax.set_ylabel("CER")
    ax.set_title(title)

def plot_cer(fname, outf, metricIndex):
    data = [x.split(",") for x in open(fname, "r").read().strip().split("\n")]
    res = {}
    for m in metrics:
        res[m] = {}
    for row in data:
        lbl = row[0].lower()
        for i, m in enumerate(metrics):
            c = convert[i]
            if m in lbl:
                for p in providers:
                    if p in lbl:
                        res[m][p] = [c(x) for x in row[1:]]

    mi = metricIndex
    m = metrics[mi]

    fig, (ax0, ax1, ax2) = plt.subplots(nrows=3, ncols=1, figsize=(8, 9), dpi=300)
    init_ax(providersAxis[0] + " " + metricsAxis[mi], ax0)
    init_ax(providersAxis[1] + " " + metricsAxis[mi], ax1)
    init_ax(providersAxis[2] + " " + metricsAxis[mi], ax2)
    fig.suptitle("OCR %s: %d Images" % (metricsAxis[mi], len(res[m][providers[0]])), size="xx-large", y=1.01)
    fig.tight_layout()
    x = list(range(len(res[m][providers[0]])))
    ax0.scatter(x, res[m][providers[0]], s=1, color="orange", label=providersAxis[0])
    ax1.scatter(x, res[m][providers[1]], s=1, color="blue", label=providersAxis[1])
    ax2.scatter(x, res[m][providers[2]], s=1, color="red", label=providersAxis[2])
    save(fig, outf)
    plt.close()

if __name__ == "__main__":
    if len(sys.argv) != 4:
        print("usage: ./plot_lev.py results.csv out.png 0")
        sys.exit(1)
    fname = sys.argv[1]
    outf = sys.argv[2]
    metricI = int(sys.argv[3])
    plot_cer(fname, outf, metricI)
