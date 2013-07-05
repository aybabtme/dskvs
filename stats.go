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
	total  time.Duration
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
		total:  sum(list),
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
	return fmt.Sprintf(
		"N=%d\n"+
			"\t total = %fs\n"+
			"\t min   = %dns\n"+
			"\t max   = %dns\n"+
			"\t avg   = %dns\n"+
			"\t med   = %dns\n"+
			"\t p75   = %dns\n"+
			"\t p90   = %dns\n"+
			"\t p99   = %dns\n"+
			"\t p999  = %dns\n"+
			"\t p9999 = %dns",
		s.n,
		s.total.Seconds(),
		s.min.Nanoseconds(),
		s.max.Nanoseconds(),
		s.avg.Nanoseconds(),
		s.median.Nanoseconds(),
		s.p75.Nanoseconds(),
		s.p90.Nanoseconds(),
		s.p99.Nanoseconds(),
		s.p999.Nanoseconds(),
		s.p9999.Nanoseconds())
}
