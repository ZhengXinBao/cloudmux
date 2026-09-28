package azure

import "testing"

func TestResourceSkuLocationParams(t *testing.T) {
	cases := []struct {
		location string
		want     string
	}{
		{"westus3", "location eq 'westus3'"},
		{"australiacentral", "location eq 'australiacentral'"},
		{"bad'region", "location eq 'bad''region'"},
	}
	for _, c := range cases {
		if got := resourceSkuLocationParams(c.location).Get("$filter"); got != c.want {
			t.Errorf("location %q: got %q, want %q", c.location, got, c.want)
		}
	}
}
