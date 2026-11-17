package google

import "testing"

func TestLoadbalancerIdFromForwardingRuleUsesUrlMapNotNumericId(t *testing.T) {
	target := "https://www.googleapis.com/compute/v1/projects/ffalcon-hw-01/global/targetHttpsProxies/us-central-1-https-msb-cms-01-target-proxy-4"
	urlMap := "https://www.googleapis.com/compute/v1/projects/ffalcon-hw-01/global/urlMaps/us-central-1-https-msb-cms-01"
	rule := &SForwardingRule{Target: target}
	rule.Id = "7478322352263255800"
	get := func(id string, out interface{}) error {
		proxy := out.(*STargetHttpProxy)
		if id != target {
			t.Fatalf("fetched %s", id)
		}
		proxy.URLMap = urlMap
		return nil
	}
	got := loadbalancerIdFromForwardingRule(get, rule)
	want := "projects/ffalcon-hw-01/global/urlMaps/us-central-1-https-msb-cms-01"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestLoadbalancerIdFromForwardingRuleUsesBackendService(t *testing.T) {
	backend := "https://www.googleapis.com/compute/v1/projects/p/regions/us-central1/backendServices/cms"
	rule := &SForwardingRule{BackendService: backend}
	got := loadbalancerIdFromForwardingRule(func(string, interface{}) error { return nil }, rule)
	want := "projects/p/regions/us-central1/backendServices/cms"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestLoadbalancerIdFromUsersSkipsFailedRule(t *testing.T) {
	bad := "https://www.googleapis.com/compute/v1/projects/p/regions/us-central1/forwardingRules/bad"
	good := "https://www.googleapis.com/compute/v1/projects/p/regions/us-central1/forwardingRules/good"
	target := "https://www.googleapis.com/compute/v1/projects/p/global/targetHttpsProxies/proxy"
	urlMap := "https://www.googleapis.com/compute/v1/projects/p/global/urlMaps/lb"
	get := func(id string, out interface{}) error {
		switch id {
		case bad:
			return errSkip
		case good:
			out.(*SForwardingRule).Target = target
		case target:
			out.(*STargetHttpProxy).URLMap = urlMap
		default:
			t.Fatalf("fetched %s", id)
		}
		return nil
	}
	got := loadbalancerIdFromUsers(get, []string{bad, good})
	want := "projects/p/global/urlMaps/lb"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

var errSkip = errorString("skip")

type errorString string

func (e errorString) Error() string { return string(e) }
