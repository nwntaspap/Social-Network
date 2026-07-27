package ratelimiter

import (
	"sync"
	"time"
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
	clients    map[string]*clientInfo
	limit      int
	windowSize time.Duration
	mu         sync.Mutex
}

func NewWindow(limit int, windowSize time.Duration) *Window {
	return &Window{
		clients:    make(map[string]*clientInfo),
		limit:      limit,
		windowSize: windowSize,
	}
}

func (w *Window) Allow(ip string) (bool, int, int64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now().Unix()
	windowSecs := int64(w.windowSize.Seconds())

	ci, exists := w.clients[ip]
	if !exists {
		ci = &clientInfo{}
		w.clients[ip] = ci
	}

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

	return allowed, remaining, resetTime
}
