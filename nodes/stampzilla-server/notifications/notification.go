package notifications

import (
	"encoding/json"
	"fmt"
)

type NotificationLevel uint8

const (
	UnknownLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	CriticalLevel
)

type Notification struct {
	Source     string `json:",omitempty"`
	SourceUuid string `json:",omitempty"`
	Level      NotificationLevel
	Message    string
}

var levelToStringMap = map[NotificationLevel]string{
	InfoLevel:     "Information",
	WarnLevel:     "Warning",
	ErrorLevel:    "Error",
	CriticalLevel: "Critical",
	UnknownLevel:  "Unknown",
}

func NewNotification(level NotificationLevel, message string) *Notification {
	return &Notification{
		Level:   level,
		Message: message,
	}
}

func (level NotificationLevel) String() string {
	levelStr, ok := levelToStringMap[level]
	if ok {
		return levelStr
	}

	return ""
}

func (l NotificationLevel) MarshalJSON() ([]byte, error) {
	if s, ok := interface{}(l).(fmt.Stringer); ok {
		return json.Marshal(s.String())
	}
	return json.Marshal(l)
}

func (l *NotificationLevel) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("NotificationLevel should be a string, got %s", data)
	}
	*l = NewNotificationLevel(s)
	return nil
}

func NewNotificationLevel(level string) NotificationLevel {
	for index, l := range levelToStringMap {
		if l == level {
			return index
		}
	}

	return UnknownLevel
}
