package domain

import (
	"errors"
	"strings"
)

type Case struct {
	ID    int64
	Title string
}

type CaseTitle struct {
	value string
}

func ParseCaseTitle(raw string) (CaseTitle, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return CaseTitle{}, errors.New("Title is required")
	}
	return CaseTitle{value: trimmed}, nil
}

func (t CaseTitle) String() string {
	return t.value
}
