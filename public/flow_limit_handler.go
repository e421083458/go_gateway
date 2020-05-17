package public

import (
	"golang.org/x/time/rate"
	"sync"
)

var FlowLimiterHandler *FlowLimiter

type FlowLimiter struct {
	FlowLmiterMap   map[string]*FlowLimiterItem
	FlowLmiterSlice []*FlowLimiterItem
	Locker          sync.RWMutex
}

type FlowLimiterItem struct {
	ServiceName string
	Limter      *rate.Limiter
}

func NewFlowLimiter() *FlowLimiter {
	return &FlowLimiter{
		FlowLmiterMap:   map[string]*FlowLimiterItem{},
		FlowLmiterSlice: []*FlowLimiterItem{},
		Locker:          sync.RWMutex{},
	}
}

func init() {
	FlowLimiterHandler = NewFlowLimiter()
}

func (counter *FlowLimiter) GetLimiter(serverName string, qps float64) (*rate.Limiter, error) {
	for _, item := range counter.FlowLmiterSlice {
		if item.ServiceName == serverName {
			return item.Limter, nil
		}
	}

	newLimiter := rate.NewLimiter(rate.Limit(qps), int(qps*3))
	item := &FlowLimiterItem{
		ServiceName: serverName,
		Limter:      newLimiter,
	}
	counter.FlowLmiterSlice = append(counter.FlowLmiterSlice, item)
	counter.Locker.Lock()
	defer counter.Locker.Unlock()
	counter.FlowLmiterMap[serverName] = item
	return newLimiter, nil
}
