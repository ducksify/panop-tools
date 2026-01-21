package main

import "regexp"

type matchConfig struct {
	Equals        string `yaml:"equals"`
	Contains      string `yaml:"contains"`
	Regex         string `yaml:"regex"`
	LoginContains string `yaml:"login_contains"`
	LoginRegex    string `yaml:"login_regex"`
}

type ruleConfig struct {
	ID          string      `yaml:"id"`
	Match       matchConfig `yaml:"match"`
	OS          string      `yaml:"os"`
	OSShortname string      `yaml:"os_shortname,omitempty"`
	// Version may be empty if unknown.
	Version string `yaml:"version,omitempty"`
}

type rulesFile struct {
	Rules []ruleConfig `yaml:"rules"`
}

type compiledRule struct {
	rule         ruleConfig
	bannerRegex  *regexp.Regexp
	loginRegex   *regexp.Regexp
	hasEquals    bool
	hasContains  bool
	hasLoginCont bool
}

type result struct {
	OS          string `json:"os"`
	OSShortname string `json:"os_shortname"`
	Source      string `json:"source"`
	RuleID      string `json:"rule_id"`
	Ver         string `json:"version"`
	Banner      string `json:"banner,omitempty"`
	Login       string `json:"login_banner,omitempty"`
}
