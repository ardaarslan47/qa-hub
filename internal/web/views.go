package web

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/ardaarslan47/qa-hub/internal/qahub"
)

type formState struct {
	Value string
	Error string
}

func runPath(id int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf("/runs/%d", id))
}

func statusLabel(status *qahub.Status) string {
	if status == nil {
		return "Untested"
	}
	return string(*status)
}
