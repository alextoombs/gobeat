package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestSetupCliApp(t *testing.T) {
	app := setupCliApp()
	if app.Name != "gobeat" {
		t.Fatal("Expected setup to set name.")
	}

	if len(app.Commands) != 3 {
		t.Fatal("Expected setup to initialize three commands.")
	}
}

func TestPostResult(t *testing.T) {
	opponent := "oleg"
	score := "9001-0"

	// Mock the result server.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Expected body to read cleanly: %s", err)
		}
		if string(b) != fmt.Sprintf("alex beat %s at ping pong with score %s",
			opponent, score) {
			t.Fatalf("Oleg definitely didn't beat Alex.")
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	u, err := url.Parse("http://" + ts.Listener.Addr().String())
	if err != nil {
		t.Fatalf("Could not parse URL http://%s: %s", ts.Listener.Addr().String(),
			err)
	}

	mockSettingsFile(t, u.String())

	if err := postResult(u, opponent, score); err != nil {
		t.Fatalf("Expected a clean post: %s", err)
	}
}

func TestRetrieveSettings(t *testing.T) {
	uStr := "foo.gov"
	mockSettingsFile(t, uStr)
	settings, err := retrieveSettings()
	if err != nil {
		t.Fatalf("Could not retrieve settings: %s", err)
	}

	if settings.TargetURL != uStr {
		t.Fatal("Did not retrieve correct settings.")
	}
}

func mockSettingsFile(t *testing.T, url string) {
	gobeatPath = filepath.Join(os.TempDir(), "mockgobeatsettings")
	settings = &gobeatSettings{
		User:      "alex",
		TargetURL: url,
		Game:      "ping pong",
	}
	if err := settings.save(); err != nil {
		t.Fatalf("Could not save settings: %s", err)
	}
}
