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
	Source     string
	SourceUuid string
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
