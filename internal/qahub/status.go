package qahub

import "errors"

var ErrStatusInvalid = errors.New("status must be Passed or Failed")

type Status struct {
	name string
}

func Statuses() []Status {
	return []Status{{"Passed"}, {"Failed"}}
}

func ParseStatus(raw string) (Status, error) {
	for _, s := range Statuses() {
		if s.name == raw {
			return s, nil
		}
	}
	return Status{}, ErrStatusInvalid
}

func (s Status) String() string {
	return s.name
}
