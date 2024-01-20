package hashrate

import (
	"sync/atomic"
	"time"
)

type HashRateReporter struct {
	Counter  uint64
	LastTime time.Time
}

func NewReporter() *HashRateReporter {
	r := &HashRateReporter{
		Counter:  0,
		LastTime: time.Now(),
	}

	go r.PrintLoop()
	return r
}

func (r *HashRateReporter) Report(hashDone uint64) {
	atomic.AddUint64(&r.Counter, hashDone)
}

func (r *HashRateReporter) PrintLoop() {
	//for {
	//	time.Sleep(time.Millisecond * 100)
	//	now := time.Now()
	//	diff := now.Sub(r.LastTime)
	//	if diff > time.Second {
	//		done := atomic.SwapUint64(&r.Counter, 0)
	//		log.Printf("hashrate: %d/s", uint64(float64(done)/diff.Seconds()))
	//		r.LastTime = now
	//	}
	//}
}
