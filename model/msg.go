package model

import "time"

type Message struct {
	ID       string
	Start    *time.Time
	End      *time.Time
	Duration *int
	Tag      string
	Data     string
}

func (m *Message) SetDuration() {
	duration := int(m.End.UnixNano() - m.Start.UnixNano())
	m.Duration = &duration
}
