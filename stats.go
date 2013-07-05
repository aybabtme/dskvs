package dskvs

import (
	"fmt"
	"sort"
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

func newStats(duration []time.Duration) stats {

	N := len(duration)
	if N == 0 {
		return stats{}
	}

	sortable := newDurationList(duration)

	sort.Sort(sort.Reverse(sortable))

	list := sortable.durations

	return stats{
		n:      N,
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
	persec := float64(s.n) / total

	return fmt.Sprintf(
		"N=%d,\n"+
			"\t %1.2f op/sec\n"+
			"\t min   = %s\n"+
			"\t max   = %s\n"+
			"\t avg   = %s\n"+
			"\t med   = %s\n"+
			"\t p75   = %s\n"+
			"\t p90   = %s\n"+
			"\t p99   = %s\n"+
			"\t p999  = %s\n"+
			"\t p9999 = %s",
		s.n,
		persec,
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
