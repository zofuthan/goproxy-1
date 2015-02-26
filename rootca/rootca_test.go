package rootca

import (
	"testing"
	"time"
)

func TestRootCA(t *testing.T) {
	r, err := NewCA("GoAgent", 3*365*24*time.Hour, 2048)
	if err != nil {
		t.Errorf("create root rootca failed: %s", err)
	}
	err = r.Dump("CA.crt")
	if err != nil {
		t.Errorf("create root rootca failed: %s", err)
	}
	r, err = NewCAFromFile("CA.crt")
	if err != nil {
		t.Errorf("issue host failed: %s", err)
	}
	_, err = r.Issue("www.google.com", 365*24*time.Hour, 2048)
	if err != nil {
		t.Errorf("issue host failed: %s", err)
	}
	_, err = r.IssueFile("test.client4.google.com", 365*24*time.Hour, 2048)
	if err != nil {
		t.Errorf("IssueToFile host failed: %s", err)
	}
}
