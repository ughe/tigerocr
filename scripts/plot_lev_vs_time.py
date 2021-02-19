#!/usr/bin/env python3

import numpy as np
from matplotlib.ticker import AutoMinorLocator
import matplotlib.pyplot as plt
import matplotlib.gridspec as gridspec
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

    m0, m1 = metrics[0], metrics[1]
    ps = sorted(providers)
    size = len(res[m0][ps[0]])
    p0 = [res[m0][ps[0]], res[m1][ps[0]]]
    p1 = [res[m0][ps[1]], res[m1][ps[1]]]
    p2 = [res[m0][ps[2]], res[m1][ps[2]]]

    fig = plt.figure(figsize=(8, 6), dpi=300, constrained_layout=True)
    spec = gridspec.GridSpec(ncols=2, nrows=3, height_ratios=[1, 1, 1],
            width_ratios=[4, 1], figure=fig)
    ax0 = fig.add_subplot(spec[0, 1])
    ax1 = fig.add_subplot(spec[1, 1])
    ax2 = fig.add_subplot(spec[2, 1])
    ax = fig.add_subplot(spec[0:, 0])

    init_ax(ax0, providersAxis[0])
    init_ax(ax1, providersAxis[1])
    init_ax(ax2, providersAxis[2])
    ax0.scatter(*p0, color="orange", s=.001, label=providersAxis[0])
    ax1.scatter(*p1, color="blue", s=.001, label=providersAxis[1])
    ax2.scatter(*p2, color="red", s=.001, label=providersAxis[2])

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
    ax.scatter(*p0, color="orange", s=1, label=providersAxis[0])
    ax.scatter(*p1, color="blue", s=1, label=providersAxis[1])
    ax.scatter(*p2, color="red", s=1, label=providersAxis[2])

    plt.close()
    save(fig, outf)

def printQuartiles(fname):
    data = [x.split(",") for x in open(fname, "r").read().strip().split("\n")]
    res = {}
    for m in metrics:
        res[m] = {}
    for row in data:
        lbl = row[0].lower()
        for m in metrics:
            if m in lbl:
                for p in providers:
                    if p in lbl:
                        res[m][p] = [float(x) for x in row[1:]]

    # Print out the quantiles
    print("Quartiles:\tmin, Q1, med, Q2, max")
    qs = [0, 25, 50, 75, 100]
    for i, m in enumerate(metrics):
        for p in providers:
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
