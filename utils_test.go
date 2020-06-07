package nelly

import (
	"testing"
)

func TestCodeToString(t *testing.T) {
	noContent := codeToString(204)
	if noContent != "204" {
		t.Errorf("expected status '204', got '%v'", noContent)
	}

	notModified := codeToString(304)
	if notModified != "304" {
		t.Errorf("expected status '304', got '%v'", notModified)
	}

	unknown := codeToString(707)
	if unknown != "707" {
		t.Errorf("expected status '707', got '%v'", unknown)
	}
}
