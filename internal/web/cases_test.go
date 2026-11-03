package web_test

import (
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestRootShowsCases(t *testing.T) {
	a := startApp(t, tempDB(t))

	resp := a.get("/")

	if resp.Status != http.StatusOK || !strings.Contains(resp.Body, "<h1>Cases</h1>") {
		t.Fatalf("GET / = %d, want the Cases page; body: %s", resp.Status, resp.Body)
	}
}

func TestCreatedCasesAreListedInCreationOrder(t *testing.T) {
	a := startApp(t, tempDB(t))

	for _, title := range []string{"Login works", "Logout works", "Password reset"} {
		if resp := a.createCase(title); resp.Status != http.StatusOK {
			t.Fatalf("create %q = %d", title, resp.Status)
		}
	}

	want := []string{"Login works", "Logout works", "Password reset"}
	if got := a.caseTitles(); !reflect.DeepEqual(got, want) {
		t.Fatalf("cases = %q, want %q", got, want)
	}
}

func TestBlankTitleIsRejectedAndNothingIsAdded(t *testing.T) {
	for name, title := range map[string]string{"empty": "", "spaces": "   ", "whitespace": " \t\n "} {
		t.Run(name, func(t *testing.T) {
			a := startApp(t, tempDB(t))
			a.createCase("Existing case")

			resp := a.createCase(title)

			if resp.Status != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422", resp.Status)
			}
			if !strings.Contains(resp.Body, "Title is required.") {
				t.Fatalf("body lacks the error message: %s", resp.Body)
			}
			if got := a.caseTitles(); !reflect.DeepEqual(got, []string{"Existing case"}) {
				t.Fatalf("cases = %q, want only the existing case", got)
			}
		})
	}
}

func TestTwoCasesMayShareATitle(t *testing.T) {
	a := startApp(t, tempDB(t))

	a.createCase("Checkout")
	resp := a.createCase("Checkout")

	if resp.Status != http.StatusOK {
		t.Fatalf("second create = %d, want 200", resp.Status)
	}
	if got := a.caseTitles(); !reflect.DeepEqual(got, []string{"Checkout", "Checkout"}) {
		t.Fatalf("cases = %q, want two Checkout cases", got)
	}
}

func TestCaseTitleIsTrimmed(t *testing.T) {
	a := startApp(t, tempDB(t))

	a.createCase("  Login works \n")

	if got := a.caseTitles(); !reflect.DeepEqual(got, []string{"Login works"}) {
		t.Fatalf("cases = %q", got)
	}
}

func TestCaseTitleIsEscaped(t *testing.T) {
	a := startApp(t, tempDB(t))

	a.createCase("<script>alert(1)</script>")

	body := a.get("/cases").Body
	if strings.Contains(body, "<script>") {
		t.Fatalf("title was rendered unescaped: %s", body)
	}
	if got := a.caseTitles(); !reflect.DeepEqual(got, []string{"<script>alert(1)</script>"}) {
		t.Fatalf("cases = %q", got)
	}
}

func TestCasesSurviveRestart(t *testing.T) {
	dbPath := tempDB(t)
	first := startApp(t, dbPath)
	first.createCase("Login works")
	first.createCase("Login works")
	first.createCase("Logout works")
	first.stop()

	second := startApp(t, dbPath)

	want := []string{"Login works", "Login works", "Logout works"}
	if got := second.caseTitles(); !reflect.DeepEqual(got, want) {
		t.Fatalf("cases after restart = %q, want %q", got, want)
	}
}
