#!/usr/bin/env python3
from matplotlib.ticker import AutoMinorLocator
import matplotlib.pyplot as plt
import matplotlib as mpl
import os
import sys
import webbrowser

def acclogs(logs, service):
    return {x[0].split(".")[0]: x[1] for x in logs if service in x[0]}
def ceil(x, l):
    return x if x < l else l
def coa(dirname, dout, ceil=lambda x: x):
    os.mkdir(dout)
    aws = dict(filter(lambda x: len(x)==2, [tuple(x.split(",")) for x in open(dirname + "/aws.txt", "r").read().split("\n")]))
    azu = dict(filter(lambda x: len(x)==2, [tuple(x.split(",")) for x in open(dirname + "/azure.txt", "r").read().split("\n")]))
    gcp = dict(filter(lambda x: len(x)==2, [tuple(x.split(",")) for x in open(dirname + "/gcp.txt", "r").read().split("\n")]))
    # Set of similar results
    x = list(set(aws.keys()) & set(azu.keys()) & set(gcp.keys()))
    print('len(x)', len(x))
    awsy = [str(ceil(float(v))) for k, v in aws.items() if k in x]
    azuy = [str(ceil(float(v))) for k, v in azu.items() if k in x]
    gcpy = [str(ceil(float(v))) for k, v in gcp.items() if k in x]
    with open(dout + "/aws.txt", "w") as f:
        f.write("\n".join(awsy))
    with open(dout + "/azure.txt", "w") as f:
        f.write("\n".join(azuy))
    with open(dout + "/gcp.txt", "w") as f:
        f.write("\n".join(gcpy))
    print("Done. Success.")

if __name__ == "__main__":
    if len(sys.argv) < 3 or len(sys.argv) > 4:
        print("usage: coallesce_cer.py oa-lev oa-lev-fmt [1.0]")
        print("       lev dir must contain: aws.txt azure.txt gcp.txt")
        sys.exit(1)
    logf = sys.argv[1]
    dst = sys.argv[2]
    ce = lambda x: x
    if len(sys.argv) == 4:
        c = sys.argv[3]
        ce = lambda x: ceil(x, float(c))
    coa(logf, dst, ce)
