package domain

import (
	"errors"
	"strings"
)

var (
	ErrRunNotFound = errors.New("run not found")
	ErrNoCases     = UserError("Create at least one case before starting a run")
)

type Run struct {
	ID   int64
	Name string
}

type RunName struct {
	value string
}

func ParseRunName(raw string) (RunName, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return RunName{}, UserError("Name is required")
	}
	return RunName{value: trimmed}, nil
}

func (n RunName) String() string {
	return n.value
}
