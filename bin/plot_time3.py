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
def plot_time(title, filenamein, fout, ceil=lambda x: x, minylim=None):
    logs = [x.split(":") for x in open(filenamein).read().split("\n")]
    logs_aws = acclogs(logs, "aws")
    logs_azure = acclogs(logs, "azure")
    logs_gcp = acclogs(logs, "gcp")
    # Exclusive logs
    x = list(set(logs_aws.keys()) & set(logs_azure.keys()) & set(logs_gcp.keys()))
    print('len(x)', len(x))
    awsy = [ceil(int(v)/1000) for k, v in logs_aws.items() if k in x]
    azuy = [ceil(int(v)/1000) for k, v in logs_azure.items() if k in x]
    gcpy = [ceil(int(v)/1000) for k, v in logs_gcp.items() if k in x]
    print('len(awsy)', len(awsy))

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
    ax1.set_ylabel("Seconds")
    ax1.scatter(rx, awsy, s=.4, color="orange", label="AWS")
    if minylim is not None:
        ax1.set_ylim(0.0, max(float(minylim), ax1.get_ylim()[1]))

    ax2 = fig.add_subplot(312)
    ax2.set_title("Azure")
    ax2.yaxis.set_minor_locator(AutoMinorLocator(5))
    ax2.xaxis.set_minor_locator(AutoMinorLocator(5))
    ax2.set_xlabel("Page Image Pointer")
    ax2.set_ylabel("Seconds")
    ax2.scatter(rx, azuy, s=.4, color="blue", label="Azure")
    if minylim is not None:
        ax2.set_ylim(0.0, max(float(minylim), ax2.get_ylim()[1]))

    ax3 = fig.add_subplot(313)
    ax3.set_title("GCP")
    ax3.yaxis.set_minor_locator(AutoMinorLocator(5))
    ax3.xaxis.set_minor_locator(AutoMinorLocator(5))
    ax3.set_xlabel("Page Image Pointer")
    ax3.set_ylabel("Seconds")
    ax3.scatter(rx, gcpy, s=.4, color="red", label="GCP")
    if minylim is not None:
        ax3.set_ylim(0.0, max(float(minylim), ax3.get_ylim()[1]))
    #lgnd = ax1.legend(bbox_to_anchor=(0.75, -0.2), ncol=3)
    #lgnd.legendHandles[0]._sizes = [30]
    #lgnd.legendHandles[1]._sizes = [30]
    #lgnd.legendHandles[2]._sizes = [30]
    save(fig, fout)
    plt.close()

if __name__ == "__main__":
    if len(sys.argv) < 4 or len(sys.argv) > 6:
        print("usage: plot_time.py OA zlog.txt out.png [12] [6]")
        sys.exit(1)
    name = sys.argv[1]
    logf = sys.argv[2]
    dst = sys.argv[3]
    ce = lambda x: x
    if len(sys.argv) >= 5:
        c = sys.argv[4]
        ce = lambda x: ceil(x, int(c))
    minylim = None
    if len(sys.argv) == 6:
        minylim = sys.argv[5]
    plot_time("OCR Time: " + name, logf, dst, ceil=ce,minylim=minylim)
