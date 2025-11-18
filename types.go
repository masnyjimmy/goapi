package goapi

type Email string

func (Email) Format() string {
	return "email"
}

type Datetime string

func (Datetime) Format() string {
	return "date-time"
}
