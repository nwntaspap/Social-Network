package ratelimiter

import (
	"sync"
	"time"

	"social-network/internal/platform/cache"
)

type clientInfo struct {
	currentWindow  windowInfo
	previousWindow windowInfo
}

type windowInfo struct {
	count     int
	startTime int64
}

type Window struct {
	cache      cache.Cache
	limit      int
	windowSize time.Duration
	mutex      sync.Mutex
}

func NewWindow(c cache.Cache, limit int, windowSize time.Duration) *Window {
	return &Window{
		cache:      c,
		limit:      limit,
		windowSize: windowSize,
	}
}

func (w *Window) Allow(ip string) (bool, int, int64) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	now := time.Now().Unix()
	windowSecs := int64(w.windowSize.Seconds())

	key := "ratelimiter:" + ip
	ci := w.loadClient(key)

	currentWindowStart := ci.currentWindow.startTime
	windowElapsed := now - currentWindowStart

	if time.Duration(windowElapsed)*time.Second >= w.windowSize {
		ci.previousWindow = windowInfo{}
		ci.currentWindow = windowInfo{count: 0, startTime: now}
		currentWindowStart = now
		windowElapsed = 0
	}

	previousWeight := float64(windowSecs-windowElapsed) / float64(windowSecs)
	weightedPrevious := float64(ci.previousWindow.count) * previousWeight
	totalCount := int(weightedPrevious) + ci.currentWindow.count

	allowed := totalCount < w.limit
	if allowed {
		ci.currentWindow.count++
	}

	remaining := max(w.limit-totalCount-1, 0)
	resetTime := currentWindowStart + windowSecs

	w.saveClient(key, ci)

	return allowed, remaining, resetTime
}

func (w *Window) loadClient(key string) *clientInfo {
	val, err := w.cache.Get(key)
	if err != nil {
		return &clientInfo{}
	}
	ci, ok := val.(*clientInfo)
	if !ok {
		return &clientInfo{}
	}
	return ci
}

func (w *Window) saveClient(key string, ci *clientInfo) {
	w.cache.Set(key, ci, 2*w.windowSize)
}
