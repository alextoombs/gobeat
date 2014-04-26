package main

import (
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
