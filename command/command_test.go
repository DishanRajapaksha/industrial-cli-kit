package command

import (
	"reflect"
	"testing"
)

var globalFlags = []Flag{
	{Name: "config", TakesValue: true},
	{Name: "profile", TakesValue: true},
	{Name: "verbose"},
}

func TestNormalizeGlobalFlags(t *testing.T) {
	got, err := NormalizeGlobalFlags([]string{"--config", "site.yaml", "--verbose", "read", "--item", "power"}, globalFlags)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"read", "--config", "site.yaml", "--verbose", "--item", "power"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("NormalizeGlobalFlags() = %#v, want %#v", got, want)
	}
}

func TestNormalizeGlobalFlagsRejectsUnknownPrefixFlag(t *testing.T) {
	if _, err := NormalizeGlobalFlags([]string{"--endpoint", "x", "read"}, globalFlags); err == nil {
		t.Fatal("unknown global flag accepted")
	}
}
