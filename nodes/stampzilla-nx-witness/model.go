package main

import (
	"fmt"
	"regexp"
)

type Rule struct {
	ActionParams      string   `json:"actionParams"`
	ActionResourceIds []string `json:"actionResourceIds"`
	ActionType        string   `json:"actionType"`
	AggregationPeriod int      `json:"aggregationPeriod"`
	Comment           string   `json:"comment"`
	Disabled          bool     `json:"disabled"`
	EventCondition    string   `json:"eventCondition"`
	EventResourceIds  []string `json:"eventResourceIds"`
	EventState        string   `json:"eventState"`
	EventType         string   `json:"eventType"`
	ID                string   `json:"id"`
	Schedule          string   `json:"schedule"`
	System            bool     `json:"system"`
}

type EventRulesResponse []Rule

var ErrRuleNotFound = fmt.Errorf("rule not found")

func (rules EventRulesResponse) ByID(id string) (Rule, error) {
	for _, v := range rules {
		if v.StampzillaDeviceID() == id {
			return v, nil
		}
	}
	return Rule{}, ErrRuleNotFound
}

/*
those are mandatory when using POST
id uuid (required)
eventType enum (required)
eventState enum (required)
actionType enum (required)
disabled boolean (required)
*/

var re = regexp.MustCompile(`stampzilla:(.*)`)

func (r Rule) StampzillaDeviceID() string {
	m := re.FindAllStringSubmatch(r.Comment, 1)
	if len(m) > 0 && len(m[0]) > 1 {
		return m[0][1]
	}
	return ""
}
