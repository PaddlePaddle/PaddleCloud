import matplotlib.pyplot as plt
import numpy as np
import matplotlib.ticker as mticker
import os
import csv

DATAFILEPATH = os.getenv("DATAPATH", "./ts.txt,./ts1.txt")
CASEID = os.getenv("CASE", "1")
DATA_MAX = int(os.getenv("DATA_MAX", "9999999"))

def clean_data(data):
    new_data = []
    for ins in data:
        new_ins = []
        for step in ins:
            ts = int(step[0])
            if ts >= 0 and ts < DATA_MAX:
                new_ins.append(step)
        new_data.append(new_ins)
    return new_data

def nearest(l, val):
    r = l[0]
    cur_dis = -1
    for v in l:
        dis = abs(int(val[0]) - int(v[0]))
        if cur_dis < 0 or dis < cur_dis:
            cur_dis = dis
            r = v
    return r

def avg(a, weight_a, b, weight_b):
    if isinstance(a, list):
        r = []
        for i, item in enumerate(a):
            v = avg(item, weight_a, b[i], weight_b)
            r.append(v)
            out = r
    elif isinstance(a, str):
        if "|" in a:
            r = []
            va = a.split("|")
            vb = b.split("|")
            total = len(va)
            if len(vb) > total:
                total = len(vb)

            for i in range(total):
                if len(a) > i:
                    fa = float(va[i])
                else:
                    fa = float(va[-1])

                if len(b) > i:
                    fb = float(vb[i])
                else:
                    fb = float(vb[-1])

                f = (fa * weight_a + fb * weight_b) / (weight_a + weight_b)
                r.append(str(f))

            out = "|".join(r)
        else:
            a = float(a)
            b = float(b)
            out = (a * weight_a + b * weight_b) / (weight_a + weight_b)
    else:
        a = float(a)
        b = float(b)
        out = (a * weight_a + b * weight_b) / (weight_a + weight_b)

    return out

def merge_two(a, weight_a, b, weight_b):
    result = []

    for v in a:
        near = nearest(b, v)
        result.append(avg(v, weight_a, near, weight_b))

    for v in b:
        near = nearest(a, v)
        result.append(avg(near, weight_a, v, weight_b))

    sorted_result = sorted(result, key=lambda v: int(v[0]))
    return sorted_result

def merge_data(data):
    data = clean_data(data)
    result = data[0]
    for i in range(len(data)-1):
        result = merge_two(result, i+1, data[i+1], 1)
    return result

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
    
    # correct data offest
    data_merged = data_csvs[0]["data"]
    data_corrected = [None] * len(data_csvs)
    standard_time_of_event = int(data_csvs[0]['time_of_event'])
    print "standard time of event", standard_time_of_event
    for index, data_csv in enumerate(data_csvs):
        time_offset = int(data_csv['time_of_event']) - standard_time_of_event
        print "time offset", time_offset
        data_corrected[index] = [[str(int(x[0])-time_offset)] + x[1:] for x in data_csv['data']]

    data_plot = merge_data(data_corrected)
    plot_data = [[] for _ in range(len(column_names))]

    for row_idx, row in enumerate(data_plot):
        for col_idx, item in enumerate(row):
            if isinstance(item, str):
                v = item.split("|")
            else:
                v = [item]
            v = np.array(v).astype(np.float)
            plot_data[col_idx].append(v)

    ax_data = np.array(plot_data[0])
    fig, axes = plt.subplots(len(plot_data)-1, sharex=True)

    #create charts
    for index, plot in enumerate(plot_data):
        if index == 0:
            continue

        plot = np.array(plot)
        name = column_names[index]
        ax = axes[index-1]
        ax.plot(ax_data, plot)
        
        ax.set_ylim(bottom=0)
        ax.yaxis.set_major_locator(mticker.MaxNLocator(4, integer=True))
        ax.set_title(name)

    plt.subplots_adjust(left=0.07, bottom=0.11, right=0.96, top=0.93, wspace=0.2, hspace=0.57)
    plt.show()
