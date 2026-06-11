package audit

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

var (
	eventChan chan Event
	once      sync.Once
	closeOnce sync.Once
	mu        sync.RWMutex
	closed    bool
)

func InitAudit(bufferSize int) {
	once.Do(func() {
		eventChan = make(chan Event, bufferSize)

		go func() {
			encoder := json.NewEncoder(os.Stdout)

			for e := range eventChan {
				_ = encoder.Encode(e)
			}
		}()
	})
}

func Close() {
	closeOnce.Do(func() {
		mu.Lock()
		defer mu.Unlock()

		closed = true

		if eventChan != nil {
			close(eventChan)
		}
	})
}

func Log(eventName, description, actor string) {
	e := Event{
		EventName:        eventName,
		EventDescription: description,
		EventActor:       actor,
		EventTimestamp:   time.Now().UTC(),
	}

	mu.RLock()
	defer mu.RUnlock()

	if closed {
		return
	}

	select {
	case eventChan <- e:
	default:
	}
}
