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
	p50    time.Duration
	p9     time.Duration
	p99    time.Duration
	p999   time.Duration
	p9999  time.Duration
}

func newStats(duration []time.Duration) stats {
	sortable := newDurationList(duration)

	sort.Sort(sort.Reverse(sortable))

	list := sortable.durations

	N := len(list)
	return stats{
		n:      N,
		total:  sum(list),
		median: list[N/2],
		avg:    avg(list),
		min:    list[N-1],
		max:    list[0],
		p50:    avg(list[:N/2]),
		p9:     avg(list[:N/10]),
		p99:    avg(list[:N/100]),
		p999:   avg(list[:N/1000]),
		p9999:  avg(list[:N/10000]),
	}
}

func sum(list []time.Duration) time.Duration {
	total := int64(0)
	for _, val := range list {
		total += val.Nanoseconds()
	}
	return time.Duration(total)
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
			"\t avg   = %dns\n"+
			"\t med   = %dns\n"+
			"\t max   = %dns\n"+
			"\t p50   = %dns\n"+
			"\t p9    = %dns\n"+
			"\t p99   = %dns\n"+
			"\t p999  = %dns\n"+
			"\t p9999 = %dns",
		s.n,
		s.total.Seconds(),
		s.min.Nanoseconds(),
		s.avg.Nanoseconds(),
		s.median.Nanoseconds(),
		s.max.Nanoseconds(),
		s.p50.Nanoseconds(),
		s.p9.Nanoseconds(),
		s.p99.Nanoseconds(),
		s.p999.Nanoseconds(),
		s.p9999.Nanoseconds())
}
