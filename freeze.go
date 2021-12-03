package main

import "k8s.io/apimachinery/pkg/util/sets"

type freezeContent struct {
	Release []freezeItem `json:"release"`
}

func (fc freezeContent) getFreezeItem(org, branch string) *freezeItem {
	for i := range fc.Release {
		v := &fc.Release[i]
		if v.Branch == branch && v.hasOrg(org) {
			return v
		}
	}

	return nil
}

type freezeItem struct {
	Branch    string   `json:"branch"`
	Community []string `json:"community"`
	Frozen    bool     `json:"frozen"`
	Owner     []string `json:"owner"`
}

func (fi *freezeItem) isFrozen() bool {
	return fi.Frozen
}

func (fi *freezeItem) hasOrg(org string) bool {
	return sets.NewString(fi.Community...).Has(org)
}

func (fi *freezeItem) isOwner(owner string) bool {
	return sets.NewString(fi.Owner...).Has(owner)
}
