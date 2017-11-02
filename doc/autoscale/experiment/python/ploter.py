import matplotlib.pyplot as plt
import numpy as np
import matplotlib.ticker as mticker
import os
import csv

DATAFILEPATH = os.getenv("DATAPATH", "./ts.txt")

if __name__ == '__main__':
    column_names = [
        "timestamp",
        "cpu util",
        "# of trainers",
        "# of not existing jobs",
        "# of pending jobs",
        "# of running jobs",
        "# of completed jobs",
        "# of ngix pods",
        "running trainers for each job",
        "cpu utils for each job"
    ]

    plots_data = map(lambda x: { 'name': x, 'data':[] }, column_names)
    def storeAverage (average_store, target_store) :
        for index, row in enumerate(average_store):
            if len(row) != 0:
                mean = np.mean(np.array(row), axis=0)
                target_store[index]["data"].append(mean)


    #generate plots_data
    with open(DATAFILEPATH, 'rb') as csvfile:
        csv_reader = csv.reader(csvfile, delimiter=',')
        #sort data by time stamp
        sorted_csv = sorted(csv_reader, key=lambda row: int(row[0]), reverse=False)
        previous_ts = -1
        average_data_store = [[] for _ in range(len(column_names))]
        for row in sorted_csv:
            for index, item in enumerate(row):
                if index == 0:
                    #jumped to another ts
                    if item != previous_ts:
                        #process previous data and append average to main store
                        storeAverage(average_data_store, plots_data)
                        average_data_store = [[] for _ in range(len(column_names))]
                    previous_ts = item
                if "|" in item:
                    item = item.split("|")
                else:
                    item  = [item]
                average_data_store[index].append(np.array(item).astype(np.float))
        storeAverage(average_data_store, plots_data)

    #timestamp as x axes
    ax_data = np.array(plots_data[0]["data"]).flatten()
    plots_data = plots_data[1:]
    fig, axes = plt.subplots(len(plots_data), sharex=True)
    for index, plot in enumerate(plots_data):
        ax = axes[index]
        subplots = np.rot90(plot["data"])
        #add average line for cpu util chart
        if plot["name"] == "cpu util":
            average_arr = np.empty(len(subplots[0]))
            average_arr.fill(np.average(subplots[0]))
            subplots = np.append(subplots, [average_arr], axis=0)

        for subplot in subplots:
            ax.plot(ax_data, subplot)
        ax.set_ylim(bottom=0)
        ax.yaxis.set_major_locator(mticker.MaxNLocator(4, integer=True))
        ax.set_title(plot["name"])

    plt.subplots_adjust(left=0.07, bottom=0.11, right=0.96, top=0.93, wspace=0.2, hspace=0.57)
    plt.show()
    
