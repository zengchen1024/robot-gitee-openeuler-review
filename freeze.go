package main

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

func (fi freezeItem) isOwner(owner string) bool {
	for _, v := range fi.Owner {
		if v == owner {
			return false
		}
	}

	return fi.isFrozen()
}
