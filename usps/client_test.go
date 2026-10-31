package usps

import (
	"strings"
	"testing"
)

func TestBaseURL(t *testing.T) {
	if got := BaseURL(EnvProduction); got != "https://apis.usps.com" {
		t.Fatal(got)
	}
	if got := BaseURL(EnvTEM); !strings.Contains(got, "apis-tem") {
		t.Fatal(got)
	}
	if got := BaseURL(EnvSandbox); got != "https://apis-tem.usps.com" {
		t.Fatal(got)
	}
}

func TestNewDefaults(t *testing.T) {
	c := New(Config{ClientID: "id", ClientSecret: "sec"})
	if c.Environment() != EnvTEM {
		t.Fatal(c.Environment())
	}
}
