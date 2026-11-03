package domain

type Status struct {
	value string
}

var (
	Passed = Status{value: "Passed"}
	Failed = Status{value: "Failed"}
)

func Statuses() []Status {
	return []Status{Passed, Failed}
}

func ParseStatus(raw string) (Status, error) {
	for _, status := range Statuses() {
		if status.value == raw {
			return status, nil
		}
	}
	return Status{}, UserError("Status must be Passed or Failed")
}

func (s Status) String() string {
	return s.value
}
