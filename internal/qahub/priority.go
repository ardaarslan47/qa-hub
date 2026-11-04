package qahub

import "errors"

var ErrPriorityInvalid = errors.New("priority must be Low, Medium or High")

type Priority struct {
	name string
}

var (
	priorityLow    = Priority{"Low"}
	priorityMedium = Priority{"Medium"}
	priorityHigh   = Priority{"High"}
)

func Priorities() []Priority {
	return []Priority{priorityLow, priorityMedium, priorityHigh}
}

func DefaultPriority() Priority {
	return priorityMedium
}

func ParsePriority(raw string) (Priority, error) {
	for _, p := range Priorities() {
		if p.name == raw {
			return p, nil
		}
	}
	return Priority{}, ErrPriorityInvalid
}

func (p Priority) String() string {
	return p.name
}
