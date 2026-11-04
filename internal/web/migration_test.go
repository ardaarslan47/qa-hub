package web_test

import (
	"database/sql"
	"fmt"
	"slices"
	"testing"
)

const schemaV3 = `
CREATE TABLE cases (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL CHECK (title <> '')
);
CREATE TABLE runs (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL CHECK (name <> '')
);
CREATE TABLE executions (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL REFERENCES runs (id),
    case_id INTEGER NOT NULL REFERENCES cases (id),
    title TEXT NOT NULL CHECK (title <> ''),
    status TEXT CHECK (status IN ('Passed', 'Failed'))
);
CREATE INDEX executions_run_id ON executions (run_id);
`

const schemaV4 = schemaV3 + `
CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL COLLATE NOCASE UNIQUE CHECK (name <> '')
);
`

const sampleData = `
INSERT INTO cases (id, title) VALUES (4, 'Login works'), (9, 'Logout works');
INSERT INTO runs (id, name) VALUES (6, 'Sprint 1'), (8, 'Sprint 2');
INSERT INTO executions (id, run_id, case_id, title, status) VALUES
    (20, 6, 4, 'Login', 'Passed'),
    (21, 6, 9, 'Logout', 'Failed'),
    (30, 8, 4, 'Login works', NULL);
`

func databaseAt(t *testing.T, version int, script string) string {
	t.Helper()
	path := tempDB(t)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(script + fmt.Sprintf("PRAGMA user_version = %d;", version)); err != nil {
		t.Fatalf("seed version %d database: %v", version, err)
	}
	return path
}

func assertSampleData(t *testing.T, p project) {
	t.Helper()
	if got, want := p.caseTitles(), []string{"Login works", "Logout works"}; !slices.Equal(got, want) {
		t.Errorf("case titles = %q, want %q", got, want)
	}
	runs := p.runs()
	wantRuns := []link{
		{Text: "Sprint 1", Href: p.path + "/runs/6"},
		{Text: "Sprint 2", Href: p.path + "/runs/8"},
	}
	if !slices.Equal(runs, wantRuns) {
		t.Fatalf("runs = %v, want %v", runs, wantRuns)
	}
	if got, want := p.executions(runs[0]), []string{"Login Passed", "Logout Failed"}; !slices.Equal(got, want) {
		t.Errorf("%s executions = %q, want %q", runs[0].Text, got, want)
	}
	if got, want := p.executions(runs[1]), []string{"Login works Untested"}; !slices.Equal(got, want) {
		t.Errorf("%s executions = %q, want %q", runs[1].Text, got, want)
	}
	for title, execution := range map[string]string{"Login": "20", "Logout": "21"} {
		want := runs[0].Href + "/executions/" + execution + "/status"
		if got := p.statusAction(runs[0], title); got != want {
			t.Errorf("%s status action = %q, want %q", title, got, want)
		}
	}
}

func TestVersion3DataMovesToTheDefaultProject(t *testing.T) {
	a := startApp(t, databaseAt(t, 3, schemaV3+sampleData))

	if got, want := a.projectNames(), []string{"Default"}; !slices.Equal(got, want) {
		t.Fatalf("project names = %q, want %q", got, want)
	}
	assertSampleData(t, a.inProject("Default"))
}

func TestEmptyVersion3DatabaseHasNoProjects(t *testing.T) {
	a := startApp(t, databaseAt(t, 3, schemaV3))

	if got := a.projectNames(); len(got) != 0 {
		t.Errorf("project names = %q, want none", got)
	}
}

func TestEitherACaseOrARunIsEnoughToCreateTheDefaultProject(t *testing.T) {
	seeds := map[string]string{
		"case only": "INSERT INTO cases (id, title) VALUES (1, 'Login works');",
		"run only":  "INSERT INTO runs (id, name) VALUES (1, 'Sprint 1');",
	}
	for name, seed := range seeds {
		t.Run(name, func(t *testing.T) {
			a := startApp(t, databaseAt(t, 3, schemaV3+seed))

			if got, want := a.projectNames(), []string{"Default"}; !slices.Equal(got, want) {
				t.Errorf("project names = %q, want %q", got, want)
			}
		})
	}
}

func TestMovedDataKeepsWorkingAcrossRestarts(t *testing.T) {
	dbPath := databaseAt(t, 3, schemaV3+sampleData)
	first := startApp(t, dbPath).inProject("Default")
	first.mustCreateCase("Reset password works")
	first.mustStartRun("Sprint 3")
	first.mustEnterStatus(first.runs()[2], "Logout works", "Passed")
	first.mustEnterStatus(first.runs()[1], "Login works", "Failed")
	first.stop()

	second := startApp(t, dbPath)

	if got, want := second.projectNames(), []string{"Default"}; !slices.Equal(got, want) {
		t.Fatalf("project names = %q, want %q", got, want)
	}
	p := second.inProject("Default")
	if got, want := second.caseTitles(), []string{"Login works", "Logout works", "Reset password works"}; !slices.Equal(got, want) {
		t.Errorf("case titles = %q, want %q", got, want)
	}
	if got, want := p.executions(p.runs()[1]), []string{"Login works Failed"}; !slices.Equal(got, want) {
		t.Errorf("Sprint 2 executions = %q, want %q", got, want)
	}
	want := []string{"Login works Untested", "Logout works Passed", "Reset password works Untested"}
	if got := p.executions(p.runs()[2]); !slices.Equal(got, want) {
		t.Errorf("Sprint 3 executions = %q, want %q", got, want)
	}
}

func TestVersion4DataMovesToDefaultBesideExistingProjects(t *testing.T) {
	a := startApp(t, databaseAt(t, 4, schemaV4+sampleData+"INSERT INTO projects (name) VALUES ('Mobile');"))

	if got, want := a.projectNames(), []string{"Mobile", "Default"}; !slices.Equal(got, want) {
		t.Fatalf("project names = %q, want %q", got, want)
	}
	assertSampleData(t, a.inProject("Default"))
	if got := a.inProject("Mobile").runNames(); len(got) != 0 {
		t.Errorf("Mobile run names = %q, want none", got)
	}
}

func TestVersion4ProjectAlreadyNamedDefaultReceivesTheData(t *testing.T) {
	a := startApp(t, databaseAt(t, 4, schemaV4+sampleData+"INSERT INTO projects (name) VALUES ('Mobile'), ('default');"))

	if got, want := a.projectNames(), []string{"Mobile", "default"}; !slices.Equal(got, want) {
		t.Fatalf("project names = %q, want %q", got, want)
	}
	assertSampleData(t, a.inProject("default"))
}

func TestVersion4DatabaseWithoutCasesOrRunsGetsNoDefaultProject(t *testing.T) {
	a := startApp(t, databaseAt(t, 4, schemaV4+"INSERT INTO projects (name) VALUES ('Mobile');"))

	if got, want := a.projectNames(), []string{"Mobile"}; !slices.Equal(got, want) {
		t.Errorf("project names = %q, want %q", got, want)
	}
}

const schemaV5 = `
CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL COLLATE NOCASE UNIQUE CHECK (name <> '')
);
CREATE TABLE cases (
    id INTEGER PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects (id),
    title TEXT NOT NULL CHECK (title <> '')
);
CREATE TABLE runs (
    id INTEGER PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects (id),
    name TEXT NOT NULL CHECK (name <> '')
);
CREATE TABLE executions (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL REFERENCES runs (id),
    case_id INTEGER NOT NULL REFERENCES cases (id),
    title TEXT NOT NULL CHECK (title <> ''),
    status TEXT CHECK (status IN ('Passed', 'Failed'))
);
CREATE INDEX cases_project_id ON cases (project_id);
CREATE INDEX runs_project_id ON runs (project_id);
CREATE INDEX executions_run_id ON executions (run_id);
`

const sampleDataInTwoProjects = `
INSERT INTO projects (id, name) VALUES (1, 'Mobile'), (2, 'Web');
INSERT INTO cases (id, project_id, title) VALUES
    (4, 1, 'Login works'),
    (5, 2, 'Login works'),
    (7, 2, 'Checkout works'),
    (9, 1, 'Logout works');
INSERT INTO runs (id, project_id, name) VALUES (6, 1, 'Sprint 1'), (8, 2, 'Sprint 2');
INSERT INTO executions (id, run_id, case_id, title, status) VALUES
    (20, 6, 4, 'Login works', 'Passed'),
    (21, 6, 9, 'Logout works', 'Failed'),
    (30, 8, 5, 'Login works', NULL),
    (31, 8, 7, 'Checkout works', 'Passed');
`

func queryString(t *testing.T, dbPath, query string) string {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var got string
	if err := db.QueryRow(query).Scan(&got); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return got
}

func TestVersion5CasesOfBothProjectsMoveToTheLibraryWithRunsAndStatusesIntact(t *testing.T) {
	dbPath := databaseAt(t, 5, schemaV5+sampleDataInTwoProjects)
	a := startApp(t, dbPath)

	if got, want := a.projectNames(), []string{"Mobile", "Web"}; !slices.Equal(got, want) {
		t.Fatalf("project names = %q, want %q", got, want)
	}
	wantCases := []string{"Login works", "Login works", "Checkout works", "Logout works"}
	if got := a.caseTitles(); !slices.Equal(got, wantCases) {
		t.Errorf("case titles = %q, want %q", got, wantCases)
	}
	mobile, web := a.inProject("Mobile"), a.inProject("Web")
	wantMobileRuns := []link{{Text: "Sprint 1", Href: mobile.path + "/runs/6"}}
	if got := mobile.runs(); !slices.Equal(got, wantMobileRuns) {
		t.Fatalf("Mobile runs = %v, want %v", got, wantMobileRuns)
	}
	wantWebRuns := []link{{Text: "Sprint 2", Href: web.path + "/runs/8"}}
	if got := web.runs(); !slices.Equal(got, wantWebRuns) {
		t.Fatalf("Web runs = %v, want %v", got, wantWebRuns)
	}
	if got, want := mobile.executions(wantMobileRuns[0]), []string{"Login works Passed", "Logout works Failed"}; !slices.Equal(got, want) {
		t.Errorf("Sprint 1 executions = %q, want %q", got, want)
	}
	if got, want := web.executions(wantWebRuns[0]), []string{"Login works Untested", "Checkout works Passed"}; !slices.Equal(got, want) {
		t.Errorf("Sprint 2 executions = %q, want %q", got, want)
	}
	if got, want := mobile.statusAction(wantMobileRuns[0], "Logout works"), wantMobileRuns[0].Href+"/executions/21/status"; got != want {
		t.Errorf("status action = %q, want %q", got, want)
	}
	a.stop()

	if got, want := queryString(t, dbPath, "SELECT group_concat(id || ':' || title, ',') FROM (SELECT id, title FROM cases ORDER BY id)"),
		"4:Login works,5:Login works,7:Checkout works,9:Logout works"; got != want {
		t.Errorf("case ids = %q, want %q", got, want)
	}
	if got, want := queryString(t, dbPath, "SELECT group_concat(id || ':' || case_id, ',') FROM (SELECT id, case_id FROM executions ORDER BY id)"),
		"20:4,21:9,30:5,31:7"; got != want {
		t.Errorf("execution cases = %q, want %q", got, want)
	}
}

func TestRunsStartedAfterAVersion5MigrationIncludeEveryCaseOfTheLibrary(t *testing.T) {
	a := startApp(t, databaseAt(t, 5, schemaV5+sampleDataInTwoProjects))
	a.mustCreateCase("Reset password works")

	mobile := a.inProject("Mobile")
	mobile.mustStartRun("Sprint 3")

	want := []string{
		"Login works Untested",
		"Login works Untested",
		"Checkout works Untested",
		"Logout works Untested",
		"Reset password works Untested",
	}
	if got := mobile.executions(mobile.runs()[1]); !slices.Equal(got, want) {
		t.Errorf("Sprint 3 executions = %q, want %q", got, want)
	}
}

const schemaV6 = `
CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL COLLATE NOCASE UNIQUE CHECK (name <> '')
);
CREATE TABLE cases (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL CHECK (title <> '')
);
CREATE TABLE runs (
    id INTEGER PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects (id),
    name TEXT NOT NULL CHECK (name <> '')
);
CREATE TABLE executions (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL REFERENCES runs (id),
    case_id INTEGER NOT NULL REFERENCES cases (id),
    title TEXT NOT NULL CHECK (title <> ''),
    status TEXT CHECK (status IN ('Passed', 'Failed'))
);
CREATE INDEX runs_project_id ON runs (project_id);
CREATE INDEX executions_run_id ON executions (run_id);
`

const sampleLibraryData = `
INSERT INTO projects (id, name) VALUES (1, 'Mobile'), (2, 'Web');
INSERT INTO cases (id, title) VALUES
    (4, 'Login works'),
    (5, 'Login works'),
    (7, 'Checkout works'),
    (9, 'Logout works');
INSERT INTO runs (id, project_id, name) VALUES (6, 1, 'Sprint 1'), (8, 2, 'Sprint 2');
INSERT INTO executions (id, run_id, case_id, title, status) VALUES
    (20, 6, 4, 'Login works', 'Passed'),
    (21, 6, 9, 'Logout works', 'Failed'),
    (30, 8, 5, 'Login works', NULL),
    (31, 8, 7, 'Checkout works', 'Passed');
`

func TestVersion6CasesGetMediumPriorityAndExecutionsAreKept(t *testing.T) {
	a := startApp(t, databaseAt(t, 6, schemaV6+sampleLibraryData))

	wantCases := []libraryCase{
		{"Login works", "Medium"},
		{"Login works", "Medium"},
		{"Checkout works", "Medium"},
		{"Logout works", "Medium"},
	}
	if got := a.libraryCases(); !slices.Equal(got, wantCases) {
		t.Errorf("library cases = %v, want %v", got, wantCases)
	}
	mobile, web := a.inProject("Mobile"), a.inProject("Web")
	sprint1, sprint2 := mobile.runs()[0], web.runs()[0]
	if got, want := a.executions(sprint1), []string{"Login works Passed", "Logout works Failed"}; !slices.Equal(got, want) {
		t.Errorf("Sprint 1 executions = %q, want %q", got, want)
	}
	if got, want := a.executions(sprint2), []string{"Login works Untested", "Checkout works Passed"}; !slices.Equal(got, want) {
		t.Errorf("Sprint 2 executions = %q, want %q", got, want)
	}
	for _, run := range []link{sprint1, sprint2} {
		want := slices.Repeat([]executionFields{{Priority: "Medium"}}, 2)
		if got := a.executionFields(run); !slices.Equal(got, want) {
			t.Errorf("%s execution fields = %q, want %q", run.Text, got, want)
		}
	}
	if got, want := a.statusAction(sprint1, "Logout works"), sprint1.Href+"/executions/21/status"; got != want {
		t.Errorf("status action = %q, want %q", got, want)
	}
}

func TestStatusCanBeEnteredOnAnExecutionOfAVersion6Run(t *testing.T) {
	a := startApp(t, databaseAt(t, 6, schemaV6+sampleLibraryData))
	sprint2 := a.inProject("Web").runs()[0]

	a.mustEnterStatus(sprint2, "Login works", "Failed")

	if got, want := a.executions(sprint2), []string{"Login works Failed", "Checkout works Passed"}; !slices.Equal(got, want) {
		t.Errorf("Sprint 2 executions = %q, want %q", got, want)
	}
}

func TestRunsStartedAfterAVersion6MigrationCopyTheNewFieldsOfEveryCase(t *testing.T) {
	a := startApp(t, databaseAt(t, 6, schemaV6+sampleLibraryData))
	a.mustCreateCaseWith("Reset password works", "Request a reset link.", "High")
	mobile := a.inProject("Mobile")

	mobile.mustStartRun("Sprint 3")

	want := append(slices.Repeat([]executionFields{{Priority: "Medium"}}, 4),
		executionFields{Priority: "High", Description: "Request a reset link."})
	if got := a.executionFields(mobile.runs()[1]); !slices.Equal(got, want) {
		t.Errorf("Sprint 3 execution fields = %q, want %q", got, want)
	}
}
