package web_test

import (
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

const (
	cycleError   = "A case cannot be its own Precondition, directly or through other cases."
	unknownError = "A Precondition must be a case in the Library."
)

func caseIDs(cases ...link) []string {
	ids := []string{}
	for _, c := range cases {
		ids = append(ids, path.Base(c.Href))
	}
	return ids
}

func (a *app) newCase(title string, preconditions ...link) link {
	a.t.Helper()
	form := url.Values{"title": {title}, "description": {""}, "priority": {"Medium"}, "preconditions": caseIDs(preconditions...)}
	if status, body := a.do(http.MethodPost, "/library", form); status != http.StatusOK {
		a.t.Fatalf("create case %q: status %d\n%s", title, status, body)
	}
	return a.caseLink(title)
}

func (a *app) postPreconditions(c link, values ...string) (int, string) {
	a.t.Helper()
	fields := a.caseFields(c)
	form := url.Values{"title": {fields.Title}, "description": {fields.Description}, "priority": {fields.Priority}, "preconditions": values}
	return a.do(http.MethodPost, c.Href, form)
}

func (a *app) mustSetPreconditions(c link, preconditions ...link) {
	a.t.Helper()
	if status, body := a.postPreconditions(c, caseIDs(preconditions...)...); status != http.StatusOK {
		a.t.Fatalf("set preconditions of %q: status %d\n%s", c.Text, status, body)
	}
}

func (a *app) preconditionsOf(c link) []link {
	a.t.Helper()
	status, body := a.do(http.MethodGet, c.Href, nil)
	if status != http.StatusOK {
		a.t.Fatalf("GET %s: status %d\n%s", c.Href, status, body)
	}
	return listLinks(a.t, body, "Preconditions")
}

func titles(links []link) []string {
	names := []string{}
	for _, l := range links {
		names = append(names, l.Text)
	}
	return names
}

func (a *app) assertPreconditions(c link, want ...string) {
	a.t.Helper()
	if want == nil {
		want = []string{}
	}
	if got := titles(a.preconditionsOf(c)); !slices.Equal(got, want) {
		a.t.Errorf("preconditions of %q = %q, want %q", c.Text, got, want)
	}
}

func assertRejected(t *testing.T, label string, status int, body, message string) {
	t.Helper()
	if status != http.StatusUnprocessableEntity {
		t.Errorf("%s: status %d, want %d", label, status, http.StatusUnprocessableEntity)
	}
	if !strings.Contains(body, message) {
		t.Errorf("%s: page does not show %q\n%s", label, message, body)
	}
}

type preconditionChoice struct {
	Title   string
	Value   string
	Checked bool
}

func preconditionChoicesOf(t *testing.T, page string) []preconditionChoice {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	var choices []preconditionChoice
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "input" && attr(n, "name") == "preconditions" {
			choices = append(choices, preconditionChoice{Title: text(n.Parent), Value: attr(n, "value"), Checked: hasAttr(n, "checked")})
		}
	}
	return choices
}

func (a *app) executionPreconditions(run link, title string) []string {
	a.t.Helper()
	status, body := a.do(http.MethodGet, run.Href, nil)
	if status != http.StatusOK {
		a.t.Fatalf("GET %s: status %d\n%s", run.Href, status, body)
	}
	for _, li := range listRows(a.t, body, "Executions") {
		if testID(li, "execution-title") != title {
			continue
		}
		copies := []string{}
		for d := range li.Descendants() {
			if d.Type == html.ElementNode && attr(d, "data-testid") == "execution-precondition" {
				copies = append(copies, text(d))
			}
		}
		return copies
	}
	a.t.Fatalf("no execution titled %q on %s\n%s", title, run.Href, body)
	return nil
}

func (a *app) assertExecutionPreconditions(run link, title string, want ...string) {
	a.t.Helper()
	if want == nil {
		want = []string{}
	}
	if got := a.executionPreconditions(run, title); !slices.Equal(got, want) {
		a.t.Errorf("%s: preconditions of execution %q = %q, want %q", run.Text, title, got, want)
	}
}

func TestACaseIsCreatedWithItsPreconditions(t *testing.T) {
	a := startApp(t, tempDB(t))
	login := a.newCase("Login works")
	profile := a.newCase("Profile loads")

	checkout := a.newCase("Checkout works", login, profile)

	a.assertPreconditions(checkout, "Login works", "Profile loads")
	if got, want := a.preconditionsOf(checkout), []link{login, profile}; !slices.Equal(got, want) {
		t.Errorf("precondition links = %v, want %v", got, want)
	}
	a.assertPreconditions(login)
}

func TestAnEditReplacesThePreconditions(t *testing.T) {
	a := startApp(t, tempDB(t))
	login := a.newCase("Login works")
	profile := a.newCase("Profile loads")
	address := a.newCase("Address saved")
	checkout := a.newCase("Checkout works")

	a.mustSetPreconditions(checkout, login, profile)
	a.assertPreconditions(checkout, "Login works", "Profile loads")

	a.mustSetPreconditions(checkout, profile, address)
	a.assertPreconditions(checkout, "Profile loads", "Address saved")

	a.mustSetPreconditions(checkout)
	a.assertPreconditions(checkout)
}

func TestTheCaseFormOffersEveryOtherCaseWithTheCurrentPreconditionsChecked(t *testing.T) {
	a := startApp(t, tempDB(t))
	login := a.newCase("Login works")
	profile := a.newCase("Profile loads")
	checkout := a.newCase("Checkout works", login)

	_, page := a.do(http.MethodGet, checkout.Href, nil)

	want := []preconditionChoice{
		{Title: "Login works", Value: caseIDs(login)[0], Checked: true},
		{Title: "Profile loads", Value: caseIDs(profile)[0]},
	}
	if got := preconditionChoicesOf(t, page); !slices.Equal(got, want) {
		t.Errorf("choices = %+v, want %+v", got, want)
	}
}

func TestTheCreateFormOffersEveryCaseOfTheLibrary(t *testing.T) {
	a := startApp(t, tempDB(t))
	login := a.newCase("Login works")
	logout := a.newCase("Logout works")

	_, page := a.do(http.MethodGet, "/library", nil)

	want := []preconditionChoice{
		{Title: "Login works", Value: caseIDs(login)[0]},
		{Title: "Logout works", Value: caseIDs(logout)[0]},
	}
	if got := preconditionChoicesOf(t, page); !slices.Equal(got, want) {
		t.Errorf("choices = %+v, want %+v", got, want)
	}
}

func TestACaseCannotChooseItselfAsAPrecondition(t *testing.T) {
	a := startApp(t, tempDB(t))
	login := a.newCase("Login works")
	profile := a.newCase("Profile loads")
	checkout := a.newCase("Checkout works", login)

	status, body := a.postPreconditions(checkout, caseIDs(checkout, profile)...)

	assertRejected(t, "self", status, body, cycleError)
	a.assertPreconditions(checkout, "Login works")
}

func TestPreconditionCyclesAreRejected(t *testing.T) {
	a := startApp(t, tempDB(t))
	first := a.newCase("Case A")
	second := a.newCase("Case B")
	third := a.newCase("Case C")
	a.mustSetPreconditions(first, second)

	status, body := a.postPreconditions(second, caseIDs(first)...)

	assertRejected(t, "A→B→A", status, body, cycleError)
	a.assertPreconditions(second)

	a.mustSetPreconditions(second, third)

	status, body = a.postPreconditions(third, caseIDs(first)...)

	assertRejected(t, "A→B→C→A", status, body, cycleError)
	a.assertPreconditions(third)
	a.assertPreconditions(first, "Case B")
	a.assertPreconditions(second, "Case C")
}

func TestARejectedCycleSavesNoneOfTheEdit(t *testing.T) {
	a := startApp(t, tempDB(t))
	login := a.newCase("Login works")
	profile := a.newCase("Profile loads", login)
	address := a.newCase("Address saved")
	a.mustEditCase(login, "Login works", "Sign in.", "High")
	a.mustSetPreconditions(login, address)

	form := url.Values{"title": {"Changed"}, "description": {"Changed."}, "priority": {"Low"}, "preconditions": caseIDs(profile)}
	status, body := a.do(http.MethodPost, login.Href, form)

	assertRejected(t, "cycle", status, body, cycleError)
	if got, want := a.caseFields(login), (caseFields{Title: "Login works", Description: "Sign in.", Priority: "High"}); got != want {
		t.Errorf("case = %+v, want %+v", got, want)
	}
	a.assertPreconditions(login, "Address saved")
}

func TestPreconditionShapesWithoutACycleAreAccepted(t *testing.T) {
	a := startApp(t, tempDB(t))
	deepest := a.newCase("Deepest")
	left := a.newCase("Left", deepest)
	right := a.newCase("Right", deepest)
	top := a.newCase("Top", left, right)

	a.assertPreconditions(top, "Left", "Right")

	a.mustSetPreconditions(top, left, right, deepest)

	a.assertPreconditions(top, "Deepest", "Left", "Right")
}

func TestAnEditMayReverseALinkItRemovesFirst(t *testing.T) {
	a := startApp(t, tempDB(t))
	first := a.newCase("Case A")
	second := a.newCase("Case B", first)

	a.mustSetPreconditions(second)
	a.mustSetPreconditions(first, second)

	a.assertPreconditions(first, "Case B")
	a.assertPreconditions(second)
}

func TestTheCasePageShowsOnlyDirectPreconditions(t *testing.T) {
	a := startApp(t, tempDB(t))
	third := a.newCase("Case C")
	second := a.newCase("Case B", third)
	first := a.newCase("Case A", second)

	a.assertPreconditions(first, "Case B")
	a.assertPreconditions(second, "Case C")
	a.assertPreconditions(third)
}

func TestAnUnknownPreconditionIsRejectedOnCreateAndOnEdit(t *testing.T) {
	a := startApp(t, tempDB(t))
	login := a.newCase("Login works")
	checkout := a.newCase("Checkout works", login)

	for _, value := range []string{"99999", "0", "-1", "abc", "", " 1", "1.5", "9223372036854775808"} {
		form := url.Values{"title": {"Profile loads"}, "description": {""}, "priority": {"Medium"}, "preconditions": {caseIDs(login)[0], value}}
		status, body := a.do(http.MethodPost, "/library", form)
		assertRejected(t, "create with "+value, status, body, unknownError)

		status, body = a.postPreconditions(checkout, value)
		assertRejected(t, "edit with "+value, status, body, unknownError)
	}

	if got, want := a.caseTitles(), []string{"Login works", "Checkout works"}; !slices.Equal(got, want) {
		t.Errorf("case titles = %q, want %q", got, want)
	}
	a.assertPreconditions(checkout, "Login works")
}

func TestAPreconditionWithTheIDOfTheCaseBeingCreatedIsRejected(t *testing.T) {
	a := startApp(t, tempDB(t))
	login := a.newCase("Login works")
	last, err := strconv.Atoi(caseIDs(login)[0])
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{"title": {"Profile loads"}, "description": {""}, "priority": {"Medium"}, "preconditions": {strconv.Itoa(last + 1)}}
	status, _ := a.do(http.MethodPost, "/library", form)

	if status != http.StatusUnprocessableEntity {
		t.Errorf("status %d, want %d", status, http.StatusUnprocessableEntity)
	}
	if got, want := a.caseTitles(), []string{"Login works"}; !slices.Equal(got, want) {
		t.Errorf("case titles = %q, want %q", got, want)
	}
}

func TestARepeatedPreconditionCountsOnce(t *testing.T) {
	a := startApp(t, tempDB(t))
	login := a.newCase("Login works")
	checkout := a.newCase("Checkout works", login, login)
	p := a.inNewProject("Mobile")

	a.assertPreconditions(checkout, "Login works")

	p.mustStartRun("Sprint 1")

	a.assertExecutionPreconditions(p.runs()[0], "Checkout works", "Login works")
}

func TestARejectedFormKeepsTheChosenPreconditions(t *testing.T) {
	a := startApp(t, tempDB(t))
	login := a.newCase("Login works")
	profile := a.newCase("Profile loads")
	checkout := a.newCase("Checkout works", login)
	loginChoice := preconditionChoice{Title: "Login works", Value: caseIDs(login)[0]}
	profileChoice := preconditionChoice{Title: "Profile loads", Value: caseIDs(profile)[0], Checked: true}

	form := url.Values{"title": {""}, "description": {""}, "priority": {"Medium"}, "preconditions": caseIDs(profile)}
	status, page := a.do(http.MethodPost, checkout.Href, form)

	assertRejected(t, "blank title on edit", status, page, "Title is required.")
	if got, want := preconditionChoicesOf(t, page), []preconditionChoice{loginChoice, profileChoice}; !slices.Equal(got, want) {
		t.Errorf("edit choices = %+v, want %+v", got, want)
	}
	if got := titles(listLinks(t, page, "Preconditions")); !slices.Equal(got, []string{"Login works"}) {
		t.Errorf("saved preconditions on the rejected page = %q", got)
	}

	status, page = a.do(http.MethodPost, "/library", form)

	assertRejected(t, "blank title on create", status, page, "Title is required.")
	checkoutChoice := preconditionChoice{Title: "Checkout works", Value: caseIDs(checkout)[0]}
	if got, want := preconditionChoicesOf(t, page), []preconditionChoice{loginChoice, profileChoice, checkoutChoice}; !slices.Equal(got, want) {
		t.Errorf("create choices = %+v, want %+v", got, want)
	}
}

func TestAnExecutionCopiesTheDirectPreconditionTitles(t *testing.T) {
	a := startApp(t, tempDB(t))
	third := a.newCase("Case C")
	second := a.newCase("Case B", third)
	a.newCase("Case A", second)
	p := a.inNewProject("Mobile")

	p.mustStartRun("Sprint 1")

	run := p.runs()[0]
	a.assertExecutionPreconditions(run, "Case A", "Case B")
	a.assertExecutionPreconditions(run, "Case B", "Case C")
	a.assertExecutionPreconditions(run, "Case C")
}

func TestTwoPreconditionsWithTheSameTitleAreBothCopied(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")
	a.mustCreateCase("Login works")
	_, library := a.do(http.MethodGet, "/library", nil)
	a.newCase("Checkout works", listLinks(t, library, "Cases")...)
	p := a.inNewProject("Mobile")

	p.mustStartRun("Sprint 1")

	a.assertExecutionPreconditions(p.runs()[0], "Checkout works", "Login works", "Login works")
}

func TestEditingPreconditionsLeavesTheCopiesOfAStartedRunUnchanged(t *testing.T) {
	a := startApp(t, tempDB(t))
	login := a.newCase("Login works")
	profile := a.newCase("Profile loads")
	checkout := a.newCase("Checkout works", login)
	p := a.inNewProject("Mobile")
	p.mustStartRun("Sprint 1")

	a.mustEditCase(login, "Sign in works", "", "Medium")
	a.mustSetPreconditions(checkout, profile)
	p.mustStartRun("Sprint 2")

	runs := p.runs()
	a.assertExecutionPreconditions(runs[0], "Checkout works", "Login works")
	a.assertExecutionPreconditions(runs[1], "Checkout works", "Profile loads")

	a.mustSetPreconditions(checkout)

	a.assertExecutionPreconditions(runs[0], "Checkout works", "Login works")
	a.assertExecutionPreconditions(runs[1], "Checkout works", "Profile loads")
}

func TestAFailedPreconditionDoesNotBlockStatusOrAddExecutions(t *testing.T) {
	a := startApp(t, tempDB(t))
	login := a.newCase("Login works")
	a.newCase("Checkout works", login)
	p := a.inNewProject("Mobile")
	p.mustStartRun("Sprint 1")
	run := p.runs()[0]

	a.mustEnterStatus(run, "Login works", "Failed")
	a.mustEnterStatus(run, "Checkout works", "Passed")

	if got, want := a.executions(run), []string{"Login works Failed", "Checkout works Passed"}; !slices.Equal(got, want) {
		t.Errorf("executions = %q, want %q", got, want)
	}
}

func TestStartedRunsOfAVersion6DatabaseShowNoCopiesAndNewOnesDo(t *testing.T) {
	a := startApp(t, databaseAt(t, 6, schemaV6+sampleLibraryData))
	web := a.inProject("Web")
	sprint2 := web.runs()[0]
	checkout := a.caseLink("Checkout works")
	logout := a.caseLink("Logout works")

	a.mustSetPreconditions(checkout, logout)
	web.mustStartRun("Sprint 3")

	a.assertPreconditions(checkout, "Logout works")
	a.assertExecutionPreconditions(sprint2, "Checkout works")
	a.assertExecutionPreconditions(web.runs()[1], "Checkout works", "Logout works")
}
