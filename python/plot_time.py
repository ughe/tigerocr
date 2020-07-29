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
def plot_time(title, filenamein, fout, ceil=lambda x: x):
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

    fig = plt.figure(figsize=(8, 3), dpi=300)
    fig.suptitle(title, size="xx-large")
    fig.subplots_adjust(hspace=1)
    fig.tight_layout()
    ax1 = fig.add_subplot(111)
    ax1.yaxis.set_minor_locator(AutoMinorLocator(5))
    ax1.xaxis.set_minor_locator(AutoMinorLocator(5))
    ax1.set_xlabel("Page Image Pointer")
    ax1.set_ylabel("Seconds")
    rx = list(range(len(x)))
    ax1.scatter(rx, awsy, s=.4, color="orange", label="AWS")
    ax1.scatter(rx, azuy, s=.4, color="blue", label="Azure")
    ax1.scatter(rx, gcpy, s=.4, color="red", label="GCP")
    lgnd = ax1.legend(bbox_to_anchor=(0.75, -0.2), ncol=3)
    lgnd.legendHandles[0]._sizes = [30]
    lgnd.legendHandles[1]._sizes = [30]
    lgnd.legendHandles[2]._sizes = [30]
    save(fig, fout)
    plt.close()

if __name__ == "__main__":
    if len(sys.argv) < 4 or len(sys.argv) > 5:
        print("usage: plot_time.py OA zlog.txt out.png [12]")
        sys.exit(1)
    name = sys.argv[1]
    logf = sys.argv[2]
    dst = sys.argv[3]
    ce = lambda x: x
    if len(sys.argv) == 5:
        c = sys.argv[4]
        ce = lambda x: ceil(x, int(c))
    plot_time("OCR Time: " + name, logf, dst, ce)
