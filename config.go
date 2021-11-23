package main

import (
	"fmt"

	libconfig "github.com/opensourceways/community-robot-lib/config"
)

type pullRequestMergeMethod string

const (
	mergeMethodeMerge pullRequestMergeMethod = "merge"
	mergeMethodSquash pullRequestMergeMethod = "squash"
)

type configuration struct {
	ConfigItems []botConfig `json:"config_items,omitempty"`
}

func (c *configuration) configFor(org, repo string) *botConfig {
	if c == nil {
		return nil
	}

	items := c.ConfigItems

	v := make([]libconfig.IPluginForRepo, len(items))
	for i := range items {
		v[i] = &items[i]
	}

	if i := libconfig.FindConfig(org, repo, v); i >= 0 {
		return &items[i]
	}
	return nil
}

func (c *configuration) Validate() error {
	if c == nil {
		return nil
	}

	items := c.ConfigItems
	for i := range items {
		if err := items[i].validate(); err != nil {
			return err
		}
	}

	return nil
}

func (c *configuration) SetDefault() {
	if c == nil {
		return
	}

	Items := c.ConfigItems
	for i := range Items {
		Items[i].setDefault()
	}
}

type botConfig struct {
	libconfig.PluginForRepo

	// LgtmCountsRequired specifies the number of lgtm label which will be need for the pr.
	// When it is greater than 1, the lgtm label is composed of 'lgtm-login'.
	// The default value is 1 which means the lgtm label is itself.
	LgtmCountsRequired uint `json:"lgtm_counts_required,omitempty"`

	// ReposOfSig specifies the repos for which it should check the devepler's permission
	// besed on the owners file in sig directory when the developer comment /lgtm or /approve
	// command. The format is 'org/repo'.
	ReposOfSig []string `json:"repos_of_sig,omitempty"`

	// LabelsForMerge specifies the labels except approved and lgtm relevant labels
	// that must be available to merge pr
	LabelsForMerge []string `json:"labels_for_merge,omitempty"`

	// MissingLabelsForMerge specifies the ones which a PR must not have to be merged.
	MissingLabelsForMerge []string `json:"missing_labels_for_merge,omitempty"`

	// MergeMethod is the method to merge PR.
	// The default method of merge. Valid options are squash and merge.
	MergeMethod pullRequestMergeMethod `json:"merge_method,omitempty"`
}

func (c *botConfig) setDefault() {
	if c.LgtmCountsRequired == 0 {
		c.LgtmCountsRequired = 1
	}

	if c.MergeMethod == "" {
		c.MergeMethod = mergeMethodeMerge
	}
}

func (c *botConfig) validate() error {
	if m := c.MergeMethod; m != mergeMethodeMerge && m != mergeMethodSquash {
		return fmt.Errorf("unsupported merge method:%s", m)
	}

	return c.PluginForRepo.Validate()
}
