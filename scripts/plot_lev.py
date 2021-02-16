#!/usr/bin/env python3
import matplotlib.pyplot as plt
import os
import sys
import webbrowser

# BEGIN configuration
providers = ["aws", "azu", "gcp"]
metrics = ["cer", "millis"]
metricsLim = [[0, 1], [0, 12]] # Min/Max threshold after conversion
convert = [
        lambda x: min(max(float(x), metricsLim[0][0]), metricsLim[0][1]),
        lambda x: min(max(int(x)/1000, metricsLim[1][0]), metricsLim[1][1]),
]
providersAxis = ["AWS", "Azure", "GCP"]
metricsAxis = ["CER", "Time"]
metricsAxisLong = ["Character Error Rate", "Time [Seconds]"]
# END configuration

def save(fig, fname, fopen=False):
    fig.savefig(fname, bbox_inches="tight")
    if fopen:
        webbrowser.open_new("file://" + os.path.abspath(fname))

def init_ax(title, ylbl, ax, mi, size):
    ax.set_ylim(metricsLim[mi][0], metricsLim[mi][1])
    ax.set_xlim(0, size)
    ax.set_xlabel("Image Pointer")
    ax.set_ylabel(ylbl)
    ax.set_title(title)

def plot_cer(fname, outf, title, metricIndex):
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
    size = len(res[m][providers[0]])

    fig, (ax0, ax1, ax2) = plt.subplots(nrows=3, ncols=1, figsize=(8, 9), dpi=300)
    init_ax(providersAxis[0], metricsAxisLong[mi], ax0, mi, size)
    init_ax(providersAxis[1], metricsAxisLong[mi], ax1, mi, size)
    init_ax(providersAxis[2], metricsAxisLong[mi], ax2, mi, size)
    fig.suptitle("%s OCR %s: %d Images" % (title, metricsAxis[mi], size), size="xx-large", y=1.01)
    fig.tight_layout()
    x = list(range(len(res[m][providers[0]])))
    ax0.scatter(x, res[m][providers[0]], s=1, color="orange")
    ax1.scatter(x, res[m][providers[1]], s=1, color="blue")
    ax2.scatter(x, res[m][providers[2]], s=1, color="red")
    save(fig, outf)
    plt.close()

if __name__ == "__main__":
    if len(sys.argv) != 5:
        print("usage: ./plot_lev.py results.csv out.png title 0")
        sys.exit(1)
    fname = sys.argv[1]
    outf = sys.argv[2]
    title = sys.argv[3]
    metricI = int(sys.argv[4])
    plot_cer(fname, outf, title, metricI)
