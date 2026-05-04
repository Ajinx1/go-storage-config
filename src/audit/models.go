package audit

import "time"

type Event struct {
	EventName        string    `json:"eventName"`
	EventDescription string    `json:"eventDescription"`
	EventActor       string    `json:"eventActor"`
	EventTimestamp   time.Time `json:"eventTimestamp"`
}
