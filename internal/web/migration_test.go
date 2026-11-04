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
	if got, want := p.caseTitles(), []string{"Login works", "Logout works", "Reset password works"}; !slices.Equal(got, want) {
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
	if got := a.inProject("Mobile").caseTitles(); len(got) != 0 {
		t.Errorf("Mobile case titles = %q, want none", got)
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
