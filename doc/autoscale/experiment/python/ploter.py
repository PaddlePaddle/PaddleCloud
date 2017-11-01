import matplotlib.pyplot as plt
import numpy as np
import matplotlib.ticker as mticker
import math
import os

DATAFILEPATH = os.getenv("DATAPATH", "./ts.txt")

if __name__ == '__main__':
    data_raw = np.rot90(np.loadtxt(DATAFILEPATH, delimiter=','))
    ax_data = data_raw[-1]
    data_arr = data_raw[0:-1]
    feature_count = len(data_arr)

    fig, axes = plt.subplots(feature_count, sharex=True)
    plt.xlabel('time (s)')

    column_names = list(reversed([
        "cpu util",
        "# of trainers",
        "# of not existing jobs",
        "# of pending jobs",
        "# of running jobs",
        "# of completed jobs",
        "# of ngix pods"
    ]))
    for i, d in enumerate(data_arr):
        ax = axes[i]
        ax.plot(ax_data, d)
        ax.set_ylim(bottom=0)
        ax.yaxis.set_major_locator(mticker.MaxNLocator(4, integer=True))
        ax.set_title(column_names[i])
        if i == feature_count-1:
            average_arr = np.empty(len(d))
            average_arr.fill(np.average(d))
            ax.plot(ax_data, average_arr, color="r")

    plt.subplots_adjust(left=0.07, bottom=0.11, right=0.96, top=0.93, wspace=0.2, hspace=0.57)
    plt.show()
    
