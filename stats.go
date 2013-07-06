package dskvs

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

type durationList struct {
	durations []time.Duration
}

func newDurationList(list []time.Duration) durationList {
	return durationList{list}
}

func (l durationList) Len() int {
	return len(l.durations)
}

func (l durationList) Swap(i, j int) {
	l.durations[i], l.durations[j] = l.durations[j], l.durations[i]
}

func (l durationList) Less(i, j int) bool {
	return l.durations[i].Nanoseconds() < l.durations[j].Nanoseconds()
}

type stats struct {
	n      int
	size   int
	median time.Duration
	avg    time.Duration
	min    time.Duration
	max    time.Duration
	p75    time.Duration
	p90    time.Duration
	p99    time.Duration
	p999   time.Duration
	p9999  time.Duration
}

func newStats(duration []time.Duration, size int) stats {

	N := len(duration)
	if N == 0 {
		return stats{}
	}

	sortable := newDurationList(duration)

	sort.Sort(sort.Reverse(sortable))

	list := sortable.durations

	return stats{
		n:      N,
		size:   size,
		median: list[N/2],
		avg:    avg(list),
		min:    list[N-1],
		max:    list[0],
		p75:    list[N/4],
		p90:    list[N/10],
		p99:    list[N/100],
		p999:   list[N/1000],
		p9999:  list[N/10000],
	}
}

func sum(list []time.Duration) time.Duration {
	var total time.Duration
	for _, val := range list {
		total += val
	}
	return total
}

func avg(list []time.Duration) time.Duration {
	if len(list) == 0 {
		return time.Duration(0)
	}
	avg := sum(list).Nanoseconds() / int64(len(list))
	return time.Duration(avg)
}

func (s *stats) String() string {

	total := float64(s.n) * s.avg.Seconds()
	totalMem := s.n * s.size
	persec := float64(s.n) / total
	persecMem := float64(totalMem) / total

	return fmt.Sprintf(
		"N=%d,\n"+
			"\t bandwidth : %6s/s\t rate : %9s qps\n"+
			"\t min   = %11s\t max   = %11s\n"+
			"\t avg   = %11s\t med   = %11s\n"+
			"\t p75   = %11s\t p90   = %11s\n"+
			"\t p99   = %11s\t p999  = %11s\n"+
			"\t p9999 = %11s",
		s.n,
		byteStr(uint64(persecMem)),
		comma(int64(persec)),
		s.min,
		s.max,
		s.avg,
		s.median,
		s.p75,
		s.p90,
		s.p99,
		s.p999,
		s.p9999)
}

/*
 * Stolen from "github.com/dustin/go-humanize"
 * Don't want to bring it as a depencendy since
 * it's only used for tests
 */

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

func humanateBytes(s uint64, base float64, sizes []string) string {
	if s < 10 {
		return fmt.Sprintf("%dB", s)
	}
	e := math.Floor(logn(float64(s), base))
	suffix := sizes[int(e)]
	val := float64(s) / math.Pow(base, math.Floor(e))
	f := "%.0f "
	if val < 10 {
		f = "%.1f "
	}

	return fmt.Sprintf(f+"%s", val, suffix)

}

func byteStr(s uint64) string {
	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	return humanateBytes(uint64(s), 1000, sizes)
}

func comma(v int64) string {
	sign := ""
	if v < 0 {
		sign = "-"
		v = 0 - v
	}

	parts := []string{"", "", "", "", "", "", "", "", ""}
	j := len(parts) - 1

	for v > 999 {
		parts[j] = strconv.FormatInt(v%1000, 10)
		switch len(parts[j]) {
		case 2:
			parts[j] = "0" + parts[j]
		case 1:
			parts[j] = "00" + parts[j]
		}
		v = v / 1000
		j--
	}
	parts[j] = strconv.Itoa(int(v))
	return sign + strings.Join(parts[j:len(parts)], ",")
}
