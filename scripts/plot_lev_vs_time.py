#!/usr/bin/env python3

import numpy as np
from matplotlib.ticker import AutoMinorLocator
import matplotlib.pyplot as plt
import matplotlib.gridspec as gridspec
import os
import sys
import webbrowser

# BEGIN configuration
colors = ["orange", "blue", "red", "cyan", "green", "purple"]
metrics = ["cer", "millis"]
convert = [
        lambda x: min(max(float(x), 0), 1),
        lambda x: min(max(int(x)/1000, 0), 12),
]
metricsAxis = ["CER", "Time"]
metricsAxisFull = ["Character Error Rate (CER)", "Time [Seconds]"]
metricFmt = ["%.05f", "%d"]
# END configuration

def save(fig, fname, fopen=False):
    fig.savefig("" + fname, bbox_inches="tight")
    if fopen:
        webbrowser.open_new("file://" + os.path.abspath("" + fname))

def init_ax(ax, title, max_time=12, max_dist=1):
    ax.set_xlim(max_dist, 0)
    ax.set_ylim(0, max_time)
    ax.set_yticks(range(0, max_time+1, 3))
    ax.yaxis.set_minor_locator(AutoMinorLocator(2))
    ax.xaxis.set_minor_locator(AutoMinorLocator(2))
    ax.set_title(title)

def main(fname, outf, title):
    data = [x.split(",") for x in open(fname, "r").read().strip().split("\n")]
    ps = []
    res = {}
    for m in metrics:
        res[m] = {}
    for row in data:
        lbl = row[0]
        for i, m in enumerate(metrics):
            c = convert[i]
            if m in lbl.lower():
                basename = lbl.split(" " + m)[0].split(" ")[0]
                if basename not in ps:
                    ps.append(basename)
                res[m][basename] = [c(x) for x in row[1:]]

    m0, m1 = metrics[0], metrics[1]
    intersection_ps = list(set(res[m0].keys()) & set(res[m1].keys()))
    ps = [x for x in ps if x in intersection_ps]
    size = len(res[m0][ps[0]])
    pxs = [[res[m0][p], res[m1][p]] for p in ps]

    fig = plt.figure(figsize=(8, 6), dpi=300, constrained_layout=True)
    spec = gridspec.GridSpec(ncols=2, nrows=len(ps), height_ratios=[1]*len(ps),
            width_ratios=[4, 1], figure=fig)
    axs = []
    for i in range(len(ps)):
        axs.append(fig.add_subplot(spec[i, 1]))
    ax = fig.add_subplot(spec[0:, 0])

    for i in range(len(ps)):
        init_ax(axs[i], ps[i])
        axs[i].scatter(*pxs[i], color=colors[i], s=.001, label=ps[i])

    maxy = 12
    ax.set_xlim(1.0, 0)
    ax.set_xticks([x/10 for x in range(0, 11)])
    ax.set_ylim(0, maxy)
    ax.set_yticks(range(0, maxy+1, 3))
    ax.yaxis.set_minor_locator(AutoMinorLocator(2))
    ax.set_title("%s %s vs. %s: %s Images" % (title, metricsAxis[0], metricsAxis[1], '{:,}'.format(size)), size="xx-large")

    ax.grid(linewidth=1, alpha=0.2)
    ax.set_xlabel(metricsAxisFull[0])
    ax.set_ylabel(metricsAxisFull[1])

    for i in range(len(ps)):
        ax.scatter(*pxs[i], color=colors[i], s=1, label=ps[i])

    plt.close()
    save(fig, outf)

def printQuartiles(fname):
    data = [x.split(",") for x in open(fname, "r").read().strip().split("\n")]
    res = {}
    for m in metrics:
        res[m] = {}
    for row in data:
        lbl = row[0]
        for m in metrics:
            if m in lbl.lower():
                basename = lbl.split(" " + m)[0].split(" ")[0]
                res[m][basename] = [float(x) for x in row[1:]]

    # Print out the quantiles
    print("Quartiles:\tmin, Q1, med, Q2, max")
    qs = [0, 25, 50, 75, 100]
    for i, m in enumerate(metrics):
        for p in sorted(list(res[m].keys())):
            q = np.percentile(res[m][p], qs)
            print(p + " " + m + ":\t" + ((metricFmt[i] + ", ")*len(qs) % (q[0], q[1], q[2], q[3], q[4])))

if __name__ == "__main__":
    if len(sys.argv) != 4:
        print("usage: ./plot_lev_vs_time.py results.csv out.png title")
        sys.exit(1)
    fname = sys.argv[1]
    outf = sys.argv[2]
    title = sys.argv[3]
    main(fname, outf, title)
    printQuartiles(fname)
