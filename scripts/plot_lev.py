#!/usr/bin/env python3
import matplotlib.pyplot as plt
import os
import sys
import webbrowser

# BEGIN configuration
colors = ["orange", "blue", "red", "cyan", "green", "purple"]
metrics = ["cer", "millis"]
metricsLim = [[0, 1], [0, 12]] # Min/Max threshold after conversion
convert = [
        lambda x: min(max(float(x), metricsLim[0][0]), metricsLim[0][1]),
        lambda x: min(max(int(x)/1000, metricsLim[1][0]), metricsLim[1][1]),
]
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
    mi = metricIndex
    m = metrics[mi]
    res = [] # list of (tuple of name, values)

    for row in data:
        lbl = row[0]
        c = convert[mi]
        if m in lbl.lower():
            basename = lbl.split(" " + m)[0].split(" ")[0]
            res.append((basename, [c(x) for x in row[1:]]))

    size = len(res[0][1])

    n = len(res)

    fig, axs = plt.subplots(nrows=n, ncols=1, figsize=(8.5, 11), dpi=300)
    for i, ax in enumerate(axs):
        init_ax(res[i][0], metricsAxisLong[mi], ax, mi, size)
    fig.suptitle("%s OCR %s: %d Images" % (title, metricsAxis[mi], size), size="xx-large", y=1.01)
    fig.tight_layout()
    x = list(range(size))
    for i, ax in enumerate(axs):
        ax.scatter(x, res[i][1], s=1, color=colors[i])
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
    if metricI < 0 or metricI >= len(metrics):
        print("usage: ./plot_lev.py results.csv out.png title N")
        print("required: 0 <= N < %d. Instead got N == %d" % (len(metrics), metricI))
        sys.exit(1)
    plot_cer(fname, outf, title, metricI)
