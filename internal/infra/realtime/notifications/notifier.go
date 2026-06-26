package notifications

import (
	"sync"

	"social-network/internal/domain/notification"
)

const ChannelCapacity int = 10

type Notifier struct {
	clients map[string][]chan *notification.Notification
	mu      sync.RWMutex
}

func NewNotifier() *Notifier {
	return &Notifier{clients: make(map[string][]chan *notification.Notification)}
}

func (s *Notifier) RegisterClient(userID string) chan *notification.Notification {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan *notification.Notification, ChannelCapacity)
	s.clients[userID] = append(s.clients[userID], ch)

	return ch
}

func (s *Notifier) UnregisterClient(userID string, ch chan *notification.Notification) {
	s.mu.Lock()
	defer s.mu.Unlock()

	clients := s.clients[userID]
	for i, client := range clients {
		if client == ch {
			s.clients[userID] = append(clients[:i], clients[i+1:]...)
			close(ch)
			break
		}
	}

	if len(s.clients[userID]) == 0 {
		delete(s.clients, userID)
	}
}

func (s *Notifier) BroadcastToUser(userID string, notification *notification.Notification) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := s.clients[userID]
	for _, ch := range clients {
		select {
		case ch <- notification:
			// send
		default:
			// channel full, skip
		}
	}
}
