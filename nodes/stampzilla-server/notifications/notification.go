package notifications

type NotificationLevel uint8

const (
	InfoLevel = iota
	WarnLevel
	ErrorLevel
	CriticalLevel
	UnknownLevel
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

func NewNotificationLevel(level string) NotificationLevel {
	for index, l := range levelToStringMap {
		if l == level {
			return index
		}
	}

	return UnknownLevel
}
