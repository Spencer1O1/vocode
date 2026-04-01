package executor

import "testing"

func TestSearchLikeQueryFromText(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"find the main function", "the main function", true},
		{"Find MAIN", "MAIN", true},
		{"search for foo", "foo", true},
		{"search foo", "foo", true},
		{"where is bar", "bar", true},
		{"where's baz", "baz", true},
		{"locate x", "x", true},
		{"find ", "", false},
		{"rename foo to bar", "", false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			got, ok := searchLikeQueryFromText(tc.in)
			if ok != tc.ok || got != tc.want {
				t.Fatalf("got (%q, %v); want (%q, %v)", got, ok, tc.want, tc.ok)
			}
		})
	}
}
