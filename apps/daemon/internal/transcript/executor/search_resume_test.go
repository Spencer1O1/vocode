package executor

import "testing"

func TestSearchResumeQuery_conversationalAnswer(t *testing.T) {
	t.Parallel()
	q := SearchResumeQuery("find stuff", "I'm looking for the stuff function")
	if q != "stuff function" {
		t.Fatalf("expected %q, got %q", "stuff function", q)
	}
}

func TestSearchResumeQuery_fallbackToOriginalSearchLike(t *testing.T) {
	t.Parallel()
	q := SearchResumeQuery("find widgets", "")
	if q != "widgets" {
		t.Fatalf("expected %q, got %q", "widgets", q)
	}
}

func TestSearchResumeQuery_answerUsesSearchLike(t *testing.T) {
	t.Parallel()
	q := SearchResumeQuery("do something else", "find bananas")
	if q != "bananas" {
		t.Fatalf("expected %q, got %q", "bananas", q)
	}
}
