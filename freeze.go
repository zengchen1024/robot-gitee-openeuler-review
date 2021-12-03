package main

import (
	"encoding/base64"
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"
)

type freezeContent struct {
	Release []freezeItem `json:"release"`
}

func (fc freezeContent) getFreezeItem(org, branch string) *freezeItem {
	for _, v := range fc.Release {
		if v.Branch == branch && v.hasOrg(org) {
			return &v
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

func (fi freezeItem) isFrozen() bool {
	return fi.Frozen
}

func (fi freezeItem) hasOrg(org string) bool {
	for _, v := range fi.Community {
		if v == org {
			return true
		}
	}

	return false
}

func (fi freezeItem) isFrozenForOwner(owner string) bool {
	for _, v := range fi.Owner {
		if v == owner {
			return false
		}
	}

	return fi.isFrozen()
}

func (fi freezeItem) getFrozenMsg(owner ...string) func() string {
	return func() string {
		if len(owner) == 0 && !fi.isFrozen() {
			return ""
		}

		if len(owner) > 0 && !fi.isFrozenForOwner(owner[0]) {
			return ""
		}

		return fmt.Sprintf(msgFrozenWithOwner, strings.Join(fi.Owner, ", "))
	}
}

func (bot *robot) getFreezeInfo(org, branch string, cfg []freezeFile) (freezeItem, error) {
	for _, v := range cfg {
		fc, err := bot.getFreezeContent(v)
		if err != nil {
			return freezeItem{}, err
		}

		if v := fc.getFreezeItem(org, branch); v != nil {
			return *v, nil
		}
	}

	return freezeItem{}, nil
}

func (bot *robot) getFreezeContent(f freezeFile) (freezeContent, error) {
	var fc freezeContent

	c, err := bot.cli.GetPathContent(f.Owner, f.Repo, f.Branch, f.Path)
	if err != nil {
		return fc, err
	}

	b, err := base64.StdEncoding.DecodeString(c.Content)
	if err != nil {
		return fc, err
	}

	err = yaml.Unmarshal(b, &fc)

	return fc, err
}
