package web_test

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
)

const usedByList = "Cases using this Precondition"

func (a *app) casesUsing(c link) []string {
	a.t.Helper()
	status, body := a.do(http.MethodGet, c.Href+"/delete", nil)
	if status != http.StatusOK {
		a.t.Fatalf("GET %s/delete: status %d\n%s", c.Href, status, body)
	}
	return titles(listLinks(a.t, body, usedByList))
}

func (a *app) confirmDeletion(c link) (int, string) {
	a.t.Helper()
	return a.do(http.MethodPost, c.Href+"/delete", url.Values{"confirm": {"yes"}})
}

func (a *app) mustDeleteCase(c link) {
	a.t.Helper()
	if status, body := a.confirmDeletion(c); status != http.StatusOK {
		a.t.Fatalf("delete case %q: status %d\n%s", c.Text, status, body)
	}
}

func TestDeleteConfirmationListsTheCasesUsingThisCaseAsAPrecondition(t *testing.T) {
	a := startApp(t, tempDB(t))
	own := a.newCase("Own Precondition")
	shared := a.newCase("Login works", own)
	a.newCase("Unrelated")
	a.newCase("Checkout works", shared)
	a.newCase("Profile loads", shared)

	if got, want := a.casesUsing(shared), []string{"Checkout works", "Profile loads"}; !slices.Equal(got, want) {
		t.Errorf("cases using %q = %q, want %q", shared.Text, got, want)
	}
	if got, want := a.casesUsing(a.caseLink("Unrelated")), []string{}; !slices.Equal(got, want) {
		t.Errorf("cases using Unrelated = %q, want %q", got, want)
	}
}

func TestConfirmedDeletionRemovesTheCaseFromTheLibrary(t *testing.T) {
	a := startApp(t, tempDB(t))
	login := a.newCase("Login works")
	a.newCase("Logout works")

	status, body := a.confirmDeletion(login)

	if status != http.StatusOK {
		t.Fatalf("status %d\n%s", status, body)
	}
	if got, want := titles(listLinks(t, body, "Cases")), []string{"Logout works"}; !slices.Equal(got, want) {
		t.Errorf("library after deletion = %q, want %q", got, want)
	}
	if status, _ := a.do(http.MethodGet, login.Href, nil); status != http.StatusNotFound {
		t.Errorf("GET %s: status %d, want %d", login.Href, status, http.StatusNotFound)
	}
}

func TestCaseIsNotDeletedWithoutConfirmation(t *testing.T) {
	a := startApp(t, tempDB(t))
	precondition := a.newCase("Login works")
	checkout := a.newCase("Checkout works", precondition)
	a.newCase("Profile loads", checkout)

	a.casesUsing(checkout)
	forms := map[string]url.Values{
		"no form":       nil,
		"empty form":    {},
		"blank confirm": {"confirm": {""}},
		"other value":   {"confirm": {"no"}},
	}
	for name, form := range forms {
		status, body := a.do(http.MethodPost, checkout.Href+"/delete", form)
		if status != http.StatusUnprocessableEntity {
			t.Errorf("%s: status %d, want %d", name, status, http.StatusUnprocessableEntity)
		}
		if got, want := titles(listLinks(t, body, usedByList)), []string{"Profile loads"}; !slices.Equal(got, want) {
			t.Errorf("%s: page does not repeat the confirmation step, lists %q, want %q", name, got, want)
		}
	}

	want := []string{"Login works", "Checkout works", "Profile loads"}
	if got := a.caseTitles(); !slices.Equal(got, want) {
		t.Errorf("case titles = %q, want %q", got, want)
	}
	a.assertPreconditions(checkout, "Login works")
	a.assertPreconditions(a.caseLink("Profile loads"), "Checkout works")
}

func TestDeletedCaseLeavesThePreconditionListsInBothDirections(t *testing.T) {
	a := startApp(t, tempDB(t))
	login := a.newCase("Login works")
	shared := a.newCase("Profile loads", login)
	checkout := a.newCase("Checkout works", shared)

	a.mustDeleteCase(shared)

	a.assertPreconditions(checkout)
	_, page := a.do(http.MethodGet, checkout.Href, nil)
	want := []preconditionChoice{{Title: "Login works", Value: caseIDs(login)[0]}}
	if got := preconditionChoicesOf(t, page); !slices.Equal(got, want) {
		t.Errorf("choices of %q = %+v, want %+v", checkout.Text, got, want)
	}
	if got, want := a.casesUsing(login), []string{}; !slices.Equal(got, want) {
		t.Errorf("cases using %q = %q, want %q", login.Text, got, want)
	}
}

func TestEditsAndRunsWorkAfterAPreconditionIsDeleted(t *testing.T) {
	a := startApp(t, tempDB(t)).inNewProject("Mobile")
	login := a.newCase("Login works")
	shared := a.newCase("Profile loads", login)
	checkout := a.newCase("Checkout works", shared, login)

	a.mustDeleteCase(shared)

	a.assertPreconditions(checkout, "Login works")
	a.mustSetPreconditions(checkout, login)
	a.assertPreconditions(checkout, "Login works")
	a.mustStartRun("Sprint 1")
	run := a.runs()[0]
	want := []string{"Login works Untested", "Checkout works Untested"}
	if got := a.executions(run); !slices.Equal(got, want) {
		t.Errorf("executions = %q, want %q", got, want)
	}
	a.assertExecutionPreconditions(run, "Checkout works", "Login works")
}

func TestDeletingAPreconditionBreaksTheCycleCheckChainThroughIt(t *testing.T) {
	a := startApp(t, tempDB(t))
	end := a.newCase("Case C")
	middle := a.newCase("Case B", end)
	start := a.newCase("Case A", middle)

	a.mustDeleteCase(middle)

	a.mustSetPreconditions(end, start)
	a.assertPreconditions(end, "Case A")
}

func TestAnExecutionOfADeletedCaseKeepsItsCopiesAndItsStatus(t *testing.T) {
	a := startApp(t, tempDB(t)).inNewProject("Mobile")
	login := a.newCase("Login works")
	a.mustCreateCaseWith("Checkout works", "Pay with a card.", "High")
	checkout := a.caseLink("Checkout works")
	a.mustSetPreconditions(checkout, login)
	a.newCase("Logout works")
	a.mustStartRun("Sprint 1")
	run := a.runs()[0]
	a.mustEnterStatus(run, "Checkout works", "Failed")

	a.mustDeleteCase(checkout)
	a.mustDeleteCase(login)

	wantExecutions := []string{"Login works Untested", "Checkout works Failed", "Logout works Untested"}
	if got := a.executions(run); !slices.Equal(got, wantExecutions) {
		t.Errorf("executions = %q, want %q", got, wantExecutions)
	}
	wantFields := []executionFields{
		{Priority: "Medium"},
		{Priority: "High", Description: "Pay with a card."},
		{Priority: "Medium"},
	}
	if got := a.executionFields(run); !slices.Equal(got, wantFields) {
		t.Errorf("execution fields = %q, want %q", got, wantFields)
	}
	a.assertExecutionPreconditions(run, "Checkout works", "Login works")
}

func TestStatusCanBeEnteredOnAnExecutionOfADeletedCase(t *testing.T) {
	a := startApp(t, tempDB(t)).inNewProject("Mobile")
	login := a.newCase("Login works")
	logout := a.newCase("Logout works")
	a.mustStartRun("Sprint 1")
	run := a.runs()[0]
	a.mustEnterStatus(run, "Logout works", "Failed")

	a.mustDeleteCase(login)
	a.mustDeleteCase(logout)
	a.mustEnterStatus(run, "Login works", "Passed")
	a.mustEnterStatus(run, "Logout works", "Passed")

	want := []string{"Login works Passed", "Logout works Passed"}
	if got := a.executions(run); !slices.Equal(got, want) {
		t.Errorf("executions = %q, want %q", got, want)
	}
}

func TestADeletedCaseStaysOutOfRunsStartedLater(t *testing.T) {
	a := startApp(t, tempDB(t)).inNewProject("Mobile")
	login := a.newCase("Login works")
	a.newCase("Logout works")
	a.mustStartRun("Sprint 1")

	a.mustDeleteCase(login)
	a.mustStartRun("Sprint 2")

	runs := a.runs()
	if got, want := a.executions(runs[0]), []string{"Login works Untested", "Logout works Untested"}; !slices.Equal(got, want) {
		t.Errorf("%s executions = %q, want %q", runs[0].Text, got, want)
	}
	if got, want := a.executions(runs[1]), []string{"Logout works Untested"}; !slices.Equal(got, want) {
		t.Errorf("%s executions = %q, want %q", runs[1].Text, got, want)
	}
}

func TestDeletingTheLastCaseLeavesNothingToRun(t *testing.T) {
	a := startApp(t, tempDB(t)).inNewProject("Mobile")
	a.mustDeleteCase(a.newCase("Login works"))

	status, body := a.startRun("Sprint 1")

	if status != http.StatusUnprocessableEntity {
		t.Errorf("status %d, want %d", status, http.StatusUnprocessableEntity)
	}
	if !strings.Contains(body, "Create a case before starting a run.") {
		t.Errorf("page does not show the error\n%s", body)
	}
}

func TestDeletingAMissingCaseIs404(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.newCase("Login works")

	for _, path := range []string{"/library/0/delete", "/library/-1/delete", "/library/99999/delete", "/library/abc/delete", "/library/9223372036854775808/delete"} {
		if status, _ := a.do(http.MethodGet, path, nil); status != http.StatusNotFound {
			t.Errorf("GET %s: status %d, want %d", path, status, http.StatusNotFound)
		}
		if status, _ := a.do(http.MethodPost, path, url.Values{"confirm": {"yes"}}); status != http.StatusNotFound {
			t.Errorf("POST %s: status %d, want %d", path, status, http.StatusNotFound)
		}
	}

	if got, want := a.caseTitles(), []string{"Login works"}; !slices.Equal(got, want) {
		t.Errorf("case titles = %q, want %q", got, want)
	}
}

func TestDeletingACaseTwiceIs404(t *testing.T) {
	a := startApp(t, tempDB(t))
	login := a.newCase("Login works")
	a.newCase("Logout works")
	a.mustDeleteCase(login)

	if status, _ := a.do(http.MethodGet, login.Href+"/delete", nil); status != http.StatusNotFound {
		t.Errorf("GET confirmation: status %d, want %d", status, http.StatusNotFound)
	}
	if status, _ := a.confirmDeletion(login); status != http.StatusNotFound {
		t.Errorf("second deletion: status %d, want %d", status, http.StatusNotFound)
	}

	if got, want := a.caseTitles(), []string{"Logout works"}; !slices.Equal(got, want) {
		t.Errorf("case titles = %q, want %q", got, want)
	}
}

func TestAStaleDeletionCannotHitACaseCreatedAfterIt(t *testing.T) {
	a := startApp(t, tempDB(t))
	login := a.newCase("Login works")
	logout := a.newCase("Logout works")
	a.mustDeleteCase(logout)

	reset := a.newCase("Reset password works")

	if reset.Href == logout.Href {
		t.Errorf("new case took the address of the deleted one: %s", reset.Href)
	}
	if status, _ := a.confirmDeletion(logout); status != http.StatusNotFound {
		t.Errorf("stale deletion: status %d, want %d", status, http.StatusNotFound)
	}
	if got, want := a.caseTitles(), []string{login.Text, reset.Text}; !slices.Equal(got, want) {
		t.Errorf("case titles = %q, want %q", got, want)
	}
}
