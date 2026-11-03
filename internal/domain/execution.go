package domain

import "errors"

var ErrExecutionNotFound = errors.New("execution not found")

type Execution struct {
	ID     int64
	RunID  int64
	CaseID int64
	Title  string
	Status *Status
}
