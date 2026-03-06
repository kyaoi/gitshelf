package cli

import "testing"

func TestParseGitHubIssueURL(t *testing.T) {
	ref, err := parseGitHubIssueURL("https://github.com/acme/roadmap/issues/42?x=1#frag")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if ref.Owner != "acme" || ref.Repo != "roadmap" || ref.Number != 42 {
		t.Fatalf("unexpected ref: %+v", ref)
	}
	if ref.URL != "https://github.com/acme/roadmap/issues/42" {
		t.Fatalf("unexpected normalized url: %s", ref.URL)
	}
}

func TestParseGitHubIssueURLRejectsUnsupportedURL(t *testing.T) {
	if _, err := parseGitHubIssueURL("https://example.com/acme/roadmap/issues/42"); err == nil {
		t.Fatal("expected host validation error")
	}
	if _, err := parseGitHubIssueURL("https://github.com/acme/roadmap/discussions/42"); err == nil {
		t.Fatal("expected unsupported path error")
	}
}
