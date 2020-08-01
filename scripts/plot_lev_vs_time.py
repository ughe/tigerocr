#!/usr/bin/env python3

from matplotlib.ticker import AutoMinorLocator
import matplotlib.pyplot as plt
import matplotlib.gridspec as gridspec
import os
import sys
import webbrowser

def save(fig, fname, fopen=False):
    fig.savefig("" + fname, bbox_inches="tight")
    if fopen:
        webbrowser.open_new("file://" + os.path.abspath("" + fname))

def init_ax(ax, max_time, max_dist):
    ax.set_xlim(max_dist, 0)
    ax.set_ylim(0, max_time)
    ax.set_yticks(range(0, max_time+1, 3))
    ax.yaxis.set_minor_locator(AutoMinorLocator(2))
    ax.xaxis.set_minor_locator(AutoMinorLocator(2))

def main(fname, outf, max_time, max_dist):
    data = [x.split(",") for x in open(fname, "r").read().strip().split("\n")]
    aws = [(min(int(x[1]), max_dist), min(int(x[1+1])/1000, max_time)) for x in data]
    azu = [(min(int(x[3]), max_dist), min(int(x[3+1])/1000, max_time)) for x in data]
    gcp = [(min(int(x[5]), max_dist), min(int(x[5+1])/1000, max_time)) for x in data]
    fig = plt.figure(figsize=(8, 6), dpi=300, constrained_layout=True)
    spec = gridspec.GridSpec(ncols=2, nrows=3, height_ratios=[1, 1, 1],
            width_ratios=[4, 1], figure=fig)
    ax0 = fig.add_subplot(spec[0, 1])
    ax1 = fig.add_subplot(spec[1, 1])
    ax2 = fig.add_subplot(spec[2, 1])
    ax = fig.add_subplot(spec[0:, 0])
    
    ax.set_title("Pairwise Levenshtein vs. Time: %d Images" % len(data), size="xx-large")
    init_ax(ax, max_time, max_dist)
    init_ax(ax0, max_time, max_dist)
    init_ax(ax1, max_time, max_dist)
    init_ax(ax2, max_time, max_dist)
    ax0.scatter(*zip(*aws), color="orange", s=.001, label="AWS vs Azure")
    ax1.scatter(*zip(*azu), color="blue", s=.001, label="Azure vs GCP")
    ax2.scatter(*zip(*gcp), color="red", s=.001, label="GCP vs AWS")
    ax0.set_title("AWS vs Azure")
    ax1.set_title("Azure vs GCP")
    ax2.set_title("GCP vs AWS")
    
    ax.grid(linewidth=1, alpha=0.2)
    ax.set_xlabel("Pairwise Levenshtein Distance")
    ax.set_ylabel("Time [Seconds]")
    ax.scatter(*zip(*aws), color="orange", s=1, label="AWS vs Azure")
    ax.scatter(*zip(*azu), color="blue", s=1, label="Azure vs GCP")
    ax.scatter(*zip(*gcp), color="red", s=1, label="GCP vs AWS")
    
    plt.close()
    save(fig, outf)

if __name__ == "__main__":
    if len(sys.argv) != 5:
        print("usage: ./plot_lev_vs_time.py combined.csv out.png max_time max_dist")
        sys.exit(1)
    fname = sys.argv[1]
    outf = sys.argv[2]
    max_time = int(sys.argv[3])
    max_dist = int(sys.argv[4])
    main(fname, outf, max_time, max_dist)
