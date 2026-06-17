package deploy

import (
	"sync"
	"time"
)

var (
	WorkerTimings = make(map[string]time.Duration)
	mu            sync.RWMutex
)

func RecordTiming(name string, duration time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	WorkerTimings[name] = duration
}

func GetTimings() map[string]time.Duration {
	mu.RLock()
	defer mu.RUnlock()
	res := make(map[string]time.Duration)
	for k, v := range WorkerTimings { res[k] = v }
	return res
}
