package main

import (
	"strings"
	"testing"
)

func TestParseInput(t *testing.T) {
	in, err := parseInput([]string{"--email", " Admin@Oportunidades.Social ", "--name", "Equipa"}, strings.NewReader("a-strong-password\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.Email != "admin@oportunidades.social" || in.Name != "Equipa" || in.Password != "a-strong-password" {
		t.Fatalf("unexpected input: %+v", in)
	}
}

func TestParseInputRejects(t *testing.T) {
	cases := map[string]struct {
		args  []string
		stdin string
	}{
		"missing email":  {args: []string{"--name", "Equipa"}, stdin: "a-strong-password\n"},
		"invalid email":  {args: []string{"--email", "not-an-email", "--name", "Equipa"}, stdin: "a-strong-password\n"},
		"missing name":   {args: []string{"--email", "a@b.co"}, stdin: "a-strong-password\n"},
		"short password": {args: []string{"--email", "a@b.co", "--name", "Equipa"}, stdin: "short\n"},
		"no password":    {args: []string{"--email", "a@b.co", "--name", "Equipa"}, stdin: ""},
		"unknown flag":   {args: []string{"--role", "admin"}, stdin: "a-strong-password\n"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parseInput(tc.args, strings.NewReader(tc.stdin)); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
