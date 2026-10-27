package database

import (
	"strings"
	"testing"
)

func TestWithNeonPoolerCompatAddsBinaryParameters(t *testing.T) {
	in := "postgres://u:p@ep-x-pooler.c-2.us-east-2.aws.neon.tech/app?sslmode=require"
	out, err := withNeonPoolerCompat(in)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "binary_parameters=yes") {
		t.Fatalf("missing binary_parameters=yes in %q", out)
	}
}

func TestWithNeonPoolerCompatKeepsExisting(t *testing.T) {
	in := "postgres://u:p@host/db?binary_parameters=yes&sslmode=require"
	out, err := withNeonPoolerCompat(in)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "binary_parameters=yes") {
		t.Fatalf("lost binary_parameters in %q", out)
	}
}
