package consulnotifications

import (
	"time"
)

type Notification struct {
	Id      string
	Body    string
	Status  string
	Created time.Time
	Sent    bool
}

func NewNotification(id string, status string, body string) *Notification {
	notification := &Notification{Id: id, Status: status, Body: body, Created: time.Now(), Sent: false}
	return notification
}
