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

func Log(eventName, description, actor string) {
	e := Event{
		EventName:        eventName,
		EventDescription: description,
		EventActor:       actor,
		EventTimestamp:   time.Now().UTC(),
	}

	select {
	case eventChan <- e:
		// sent successfully
	default:
		// Nothing
	}
}
