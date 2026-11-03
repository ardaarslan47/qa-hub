package domain

import "strings"

type Case struct {
	ID    int64
	Title string
}

type CaseTitle struct {
	value string
}

func ParseCaseTitle(raw string) (CaseTitle, error) {
	trimmed := strings.TrimSpace(raw)
	if !hasVisibleRune(trimmed) {
		return CaseTitle{}, UserError("Title is required")
	}
	return CaseTitle{value: trimmed}, nil
}

func (t CaseTitle) String() string {
	return t.value
}
