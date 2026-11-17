package google

import "strings"

// loadbalancerIdFromUsers 把地址的 users（转发规则 selfLink）解析成负载均衡在 CMDB 里使用的云上 ID。
// HTTP(S)/gRPC 负载均衡是 URL Map，TCP/SSL 是后端服务。都取 selfLink 去掉 API 前缀后的路径，
// 不用转发规则自己的数字 id。
func loadbalancerIdFromUsers(get func(string, interface{}) error, users []string) string {
	if get == nil {
		return ""
	}
	for _, user := range users {
		if !strings.Contains(user, "/forwardingRules/") {
			continue
		}
		rule := &SForwardingRule{}
		if err := get(user, rule); err != nil {
			continue
		}
		if id := loadbalancerIdFromForwardingRule(get, rule); id != "" {
			return id
		}
	}
	return ""
}

func loadbalancerIdFromForwardingRule(get func(string, interface{}) error, rule *SForwardingRule) string {
	if rule == nil {
		return ""
	}
	target := strings.TrimSpace(rule.Target)
	if target == "" {
		return getGlobalId(rule.BackendService)
	}
	if strings.Contains(target, "/targetPools/") {
		return ""
	}
	if strings.Contains(target, "/targetTcpProxies/") || strings.Contains(target, "/targetSslProxies/") {
		proxy := &STargetTcpProxy{}
		if err := get(target, proxy); err != nil {
			return ""
		}
		return getGlobalId(proxy.Service)
	}
	proxy := &STargetHttpProxy{}
	if err := get(target, proxy); err != nil || proxy.URLMap == "" {
		return getGlobalId(rule.BackendService)
	}
	return getGlobalId(proxy.URLMap)
}
