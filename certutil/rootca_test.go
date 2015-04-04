package certutil

import (
	"testing"
	"time"
)

func TestCertUtil(t *testing.T) {
	ca, err := NewOpenCA("GoAgent", 3*365*24*time.Hour, 2048)
	if err != nil {
		t.Errorf("create root certutil failed: %s", err)
	}
	err = ca.Dump("CA.crt")
	if err != nil {
		t.Errorf("create root certutil failed: %s", err)
	}
	ca, err = NewOpenCAFromFile("CA.crt")
	if err != nil {
		t.Errorf("issue host failed: %s", err)
	}
	_, err = ca.Issue("www.google.com", 365*24*time.Hour, 2048)
	if err != nil {
		t.Errorf("issue host failed: %s", err)
	}
	_, err = ca.IssueFile("test.client4.google.com", 365*24*time.Hour, 2048)
	if err != nil {
		t.Errorf("IssueToFile host failed: %s", err)
	}
}
