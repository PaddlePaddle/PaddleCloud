## Plotter usage

### Preprocess

```bash
$ cd $REPO/doc/autoscale/experiment/result
$ ./plot/preprocess.sh `ls */*.log`
```

### Plot Experiment Result Graphs

1. [Install Go](https://golang.org/doc/install)

1. Run the command below:
   ```bash
   go run plot/plot.go -pattern '*/*.log.csv'
   ```

   The experiment result graphs will be generated in the current folder.

### Calculate Result Report

Run the command below:
```bash
./calculate_result.sh
```
