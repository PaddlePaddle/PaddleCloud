package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"

	"github.com/lucasb-eyer/go-colorful"
)

type job struct {
	caseNumber   int
	autoscaling  bool
	trainerCount int
}

type row struct {
	timestamp               int
	cpuUtil                 float64
	runningTrainerCount     int
	notExistJobCount        int
	pendingJobCount         int
	runningJobCount         int
	doneJobCount            int
	nginxCount              int
	jobRunningTrainerCounts []int
	jobCPUUtils             []float64
}

type jobCase []row

func parseJobCase(path string) (j jobCase) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	idx := 0
	s := bufio.NewScanner(bytes.NewReader(b))
	ts := 0
	for s.Scan() {
		var r row
		ss := strings.Split(s.Text(), ",")
		r.timestamp, err = strconv.Atoi(ss[0])
		if err != nil {
			panic(err)
		}

		if r.timestamp < ts {
			continue
		}
		ts = r.timestamp

		r.cpuUtil, err = strconv.ParseFloat(ss[1], 64)
		if err != nil {
			panic(err)
		}

		r.runningTrainerCount, err = strconv.Atoi(ss[2])
		if err != nil {
			panic(err)
		}

		r.notExistJobCount, err = strconv.Atoi(ss[3])
		if err != nil {
			panic(err)
		}

		r.pendingJobCount, err = strconv.Atoi(ss[4])
		if err != nil {
			panic(err)
		}

		r.runningJobCount, err = strconv.Atoi(ss[5])
		if err != nil {
			panic(err)
		}

		r.doneJobCount, err = strconv.Atoi(ss[6])
		if err != nil {
			panic(err)
		}

		r.nginxCount, err = strconv.Atoi(ss[7])
		if err != nil {
			panic(err)
		}

		cur := 8
		remain := len(ss) - cur
		trainerCount := remain / 2
		if remain != trainerCount*2 {
			panic(fmt.Errorf("unrecognized row at %s:%d", path, idx))
		}

		r.jobRunningTrainerCounts = make([]int, trainerCount)
		r.jobCPUUtils = make([]float64, trainerCount)

		for i := range r.jobRunningTrainerCounts {
			c, err := strconv.ParseFloat(ss[cur], 64)
			if err != nil {
				panic(err)
			}
			r.jobRunningTrainerCounts[i] = int(c)

			cur++
		}

		for i := range r.jobCPUUtils {
			r.jobCPUUtils[i], err = strconv.ParseFloat(ss[cur], 64)
			if err != nil {
				panic(err)
			}

			cur++
		}

		j = append(j, r)
		idx++
	}

	return
}

func parseJob(path string) job {
	j := job{}
	s := strings.Split(path, "/")
	folder := s[len(s)-2]

	s = strings.Split(folder, "-")
	switch s[0] {
	case "case1":
		j.caseNumber = 1
	case "case2":
		j.caseNumber = 2
	default:
		panic(fmt.Errorf("could not recognize file path: %s", path))
	}

	switch s[2] {
	case "OFF":
		j.autoscaling = false
	case "ON":
		j.autoscaling = true
	default:
		panic(fmt.Errorf("could not recognize file path: %s", path))
	}

	count, err := strconv.Atoi(s[3])
	if err != nil {
		panic(err)
	}

	j.trainerCount = count
	return j
}

type present func(c jobCase) plotter.XYs

func casesToPoints(p present, c []jobCase) []plotter.XYs {
	r := make([]plotter.XYs, len(c))
	for i := range r {
		r[i] = p(c[i])
	}
	return r
}

func clusterUtil(c jobCase) plotter.XYs {
	r := make(plotter.XYs, len(c))
	for i, row := range c {
		r[i].X = float64(row.timestamp)
		r[i].Y = row.cpuUtil
	}

	return r
}

func pendingJobs(c jobCase) plotter.XYs {
	r := make(plotter.XYs, len(c))
	for i, row := range c {
		r[i].X = float64(row.timestamp)
		r[i].Y = float64(row.pendingJobCount)
	}

	return r
}

func nginxCount(c jobCase) plotter.XYs {
	r := make(plotter.XYs, len(c))
	for i, row := range c {
		r[i].X = float64(row.timestamp)
		r[i].Y = float64(row.nginxCount)
	}

	return r
}

func draw(p *plot.Plot, output string, caseOn, caseOff []jobCase, pre present) {
	pts := casesToPoints(pre, caseOn)
	for i := range pts {
		l, err := plotter.NewLine(pts[i])
		if err != nil {
			panic(err)
		}
		l.LineStyle.Width = vg.Points(1)
		l.LineStyle.Color = colorful.HappyColor()

		p.Add(l)
		p.Legend.Add(fmt.Sprintf("on-%d", i), l)
		if err != nil {
			panic(err)
		}
	}

	pts = casesToPoints(pre, caseOff)
	for i := range pts {
		l, err := plotter.NewLine(pts[i])
		if err != nil {
			panic(err)
		}
		l.LineStyle.Width = vg.Points(1)
		l.LineStyle.Color = colorful.WarmColor()
		l.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}

		p.Add(l)
		p.Legend.Add(fmt.Sprintf("off-%d", i), l)
		if err != nil {
			panic(err)
		}
	}

	// Save the plot to a PNG file.
	if err := p.Save(8*vg.Inch, 4*vg.Inch, output); err != nil {
		panic(err)
	}
}

var (
	case1On  = job{caseNumber: 1, autoscaling: true, trainerCount: 20}
	case1Off = job{caseNumber: 1, autoscaling: false, trainerCount: 20}
	case2On  = job{caseNumber: 2, autoscaling: true, trainerCount: 6}
	case2Off = job{caseNumber: 2, autoscaling: false, trainerCount: 6}
)

func main() {
	pattern := flag.String("pattern", "", "input files")
	flag.Parse()

	matches, err := filepath.Glob(*pattern)
	if err != nil {
		panic(err)
	}

	if len(matches) == 0 {
		panic("no file matched from pattern")
	}

	cases := make(map[job][]jobCase)

	for _, path := range matches {
		j := parseJob(path)
		c := parseJobCase(path)
		cases[j] = append(cases[j], c)
	}

	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "Case 1 Cluster Utilization"
	p.X.Label.Text = "Timestamp"
	p.X.Min = 0
	p.X.Max = 600
	p.Y.Label.Text = "Utilization Percentage"
	p.Y.Min = 0
	p.Y.Max = 100
	p.Add(plotter.NewGrid())
	draw(p, "case1_util.png", cases[case1On], cases[case1Off], clusterUtil)

	p, err = plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "Case 2 Cluster Utilization"
	p.X.Label.Text = "Timestamp"
	p.X.Min = 0
	p.X.Max = 600
	p.Y.Label.Text = "Utilization Percentage"
	p.Y.Min = 0
	p.Y.Max = 100
	p.Add(plotter.NewGrid())
	draw(p, "case2_util.png", cases[case2On], cases[case2Off], clusterUtil)

	p, err = plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "Case 1 Pending Job Count"
	p.X.Label.Text = "Timestamp"
	p.X.Min = 0
	p.X.Max = 600
	p.Y.Label.Text = "Number of Pending Jobs"
	p.Y.Min = 0
	p.Y.Max = 16
	p.Add(plotter.NewGrid())
	draw(p, "case1_pending.png", cases[case1On], cases[case1Off], pendingJobs)

	p, err = plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "Case 2 Pending Job Count"
	p.X.Label.Text = "Timestamp"
	p.X.Min = 0
	p.X.Max = 600
	p.Y.Label.Text = "Number of Pending Jobs"
	p.Y.Min = 0
	p.Y.Max = 4
	p.Add(plotter.NewGrid())
	draw(p, "case2_pending.png", cases[case2On], cases[case2Off], pendingJobs)
	p, err = plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Case 2 Nginx Count"
	p.X.Label.Text = "Timestamp"
	p.X.Min = 0
	p.X.Max = 600
	p.Y.Label.Text = "Number of Nginx Pods"
	p.Y.Min = 0
	p.Y.Max = 420
	p.Add(plotter.NewGrid())
	draw(p, "case2_nginx.png", cases[case2On], cases[case2Off], nginxCount)
}
