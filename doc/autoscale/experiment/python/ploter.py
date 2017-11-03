import matplotlib.pyplot as plt
import numpy as np
import matplotlib.ticker as mticker
import os
import csv

DATAFILEPATH = os.getenv("DATAPATH", "./ts.txt ./ts1.txt")
CASEID = os.getenv("CASE", "1")

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

    def smooth(y, box_pts):
        box = np.ones(box_pts)/box_pts
        y_smooth = np.convolve(y, box, mode='same')
        return y_smooth

    # read csv files
    data_csvs = []
    datafiles = DATAFILEPATH.split(",")
    for filepath in datafiles:
        with open(filepath, "rb") as csvfile:
            csv_reader = csv.reader(csvfile, delimiter=',')
            data_csvs.append({
                'data': list(csv_reader),
                'time_of_event': 0
            })
    
    '''
        start aligning different batch of data by time of event
        case one's event is when # of not existing job become 0
        case tow's event is when # of ngix pods start to change
    '''
    # find time of event
    for index, data_csv in enumerate(data_csvs):
        previous_row = data_csv["data"][0]
        for row in data_csv["data"]:
            if CASEID == "1":
                if row[3] == "0":
                    print row
                    data_csv['time_of_event'] = row[0]
                    break
            if CASEID == "2":
                if row[7] != previous_row[7]:
                    data_csv['time_of_event'] = row[0]
                    break
            previous_row = row
    
    # add offset to csv file and merge
    data_merged = data_csvs[0]["data"]
    standard_time_of_event = int(data_csvs[0]['time_of_event'])
    print "standard time of event", standard_time_of_event
    for index, data_csv in enumerate(data_csvs):
        if index == 0:
            continue
        time_offset = int(data_csv['time_of_event']) - standard_time_of_event
        for row in data_csv["data"]:
            if abs(int(row[0]) - (standard_time_of_event + time_offset))<2:
                print row[0]
                print (standard_time_of_event + time_offset)
                print row
            row[0] = int(row[0]) + time_offset
            if row[0] >= 0:
                data_merged.append(row)

    # deal with dupliated timestamps
    sorted_csv = sorted(data_merged, key=lambda row: int(row[0]), reverse=False)
    for i in sorted_csv:
        pass
        #print i
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
            if isinstance(item, basestring) and "|" in item:
                item = item.split("|")
            else:
                item = [item]
            average_data_store[index].append(np.array(item).astype(np.float))
    storeAverage(average_data_store, plots_data)

    #timestamp as x axes
    ax_data = np.array(plots_data[0]["data"]).flatten()
    plots_data = plots_data[1:]
    fig, axes = plt.subplots(len(plots_data), sharex=True)
    #create charts
    for index, plot in enumerate(plots_data):
        print plot["name"]
        ax = axes[index]
        subplots = np.rot90(plot["data"])
        #add average line for cpu util chart
        if plot["name"] == "cpu util":
            average_arr = np.empty(len(subplots[0]))
            average_arr.fill(np.average(subplots[0]))
            subplots = np.append(subplots, [average_arr], axis=0)

        for subplot in subplots:
            #print subplot
            y = smooth(subplot, 12)
            ax.plot(ax_data, y)
        ax.set_ylim(bottom=0)
        ax.yaxis.set_major_locator(mticker.MaxNLocator(4, integer=True))
        ax.set_title(plot["name"])

    plt.subplots_adjust(left=0.07, bottom=0.11, right=0.96, top=0.93, wspace=0.2, hspace=0.57)
    plt.show()