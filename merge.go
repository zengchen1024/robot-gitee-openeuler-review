package main

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	sdk "gitee.com/openeuler/go-gitee/gitee"
	"github.com/opensourceways/community-robot-lib/giteeclient"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"
)

const (
	msgPRConflicts        = "PR conflicts to the target branch."
	msgMissingLabels      = "PR does not have these lables: %s"
	msgInvalidLabels      = "PR should remove these labels: %s"
	msgNotEnoughLGTMLabel = "PR needs %d lgtm labels and now gets %d"
	msgFrozenWithOwner    = "The target branch of PR has been frozen and it can be merge only by branch owners: %s"
)

var regCheckPr = regexp.MustCompile(`(?mi)^/check-pr\s*$`)

func (bot *robot) handleCheckPR(e *sdk.NoteEvent, cfg *botConfig, log *logrus.Entry) error {
	ne := giteeclient.NewPRNoteEvent(e)

	if !ne.IsPullRequest() ||
		!ne.IsPROpen() ||
		!ne.IsCreatingCommentEvent() ||
		!regCheckPr.MatchString(ne.GetComment()) {
		return nil
	}

	return bot.tryMerge(ne, cfg, true, log)
}

func (bot *robot) tryMerge(e giteeclient.PRNoteEvent, cfg *botConfig, addComment bool, log *logrus.Entry) error {
	org, repo := e.GetOrgRep()

	h := mergeHelper{
		cfg:     cfg,
		org:     org,
		repo:    repo,
		cli:     bot.cli,
		pr:      e.GetPullRequest(),
		trigger: e.GetCommenter(),
	}

	if r, ok := h.canMerge(log); !ok {
		if len(r) > 0 && addComment {
			return bot.cli.CreatePRComment(
				org, repo, e.GetPRNumber(),
				fmt.Sprintf(
					"@%s , this pr is not mergeable and the reasons are below:\n%s",
					e.GetCommenter(), strings.Join(r, "\n"),
				),
			)
		}
	}

	return h.merge()
}

func (bot *robot) handleLabelUpdate(e *sdk.PullRequestEvent, cfg *botConfig, log *logrus.Entry) error {
	if giteeclient.GetPullRequestAction(e) != giteeclient.PRActionUpdatedLabel {
		return nil
	}

	org, repo := giteeclient.GetOwnerAndRepoByPREvent(e)

	h := mergeHelper{
		cfg:  cfg,
		org:  org,
		repo: repo,
		cli:  bot.cli,
		pr:   e.GetPullRequest(),
	}

	if _, ok := h.canMerge(log); ok {
		return h.merge()
	}

	return nil
}

type mergeHelper struct {
	pr  *sdk.PullRequestHook
	cfg *botConfig

	org     string
	repo    string
	trigger string

	cli iClient
}

func (m *mergeHelper) merge() error {
	number := m.pr.Number

	if m.pr.NeedReview || m.pr.NeedTest {
		v := int32(0)
		p := sdk.PullRequestUpdateParam{
			AssigneesNumber: &v,
			TestersNumber:   &v,
		}

		if _, err := m.cli.UpdatePullRequest(m.org, m.repo, number, p); err != nil {
			return err
		}
	}

	return m.cli.MergePR(
		m.org, m.repo, number,
		sdk.PullRequestMergePutParam{
			MergeMethod: string(m.cfg.MergeMethod),
		},
	)
}

func (m *mergeHelper) canMerge(log *logrus.Entry) ([]string, bool) {
	if !m.pr.GetMergeable() {
		return []string{msgPRConflicts}, false
	}

	labels := sets.NewString()
	for _, item := range m.pr.Labels {
		labels.Insert(item.Name)
	}

	if r := isLabelMatched(labels, m.cfg); len(r) > 0 {
		return r, false
	}

	freeze, err := m.getFreezeInfo(log)
	if err != nil {
		return nil, false
	}

	if freeze == nil || !freeze.isFrozen() {
		return nil, true
	}

	if m.trigger == "" {
		return nil, false
	}

	if freeze.isOwner(m.trigger) {
		return nil, true
	}

	return []string{
		fmt.Sprintf(msgFrozenWithOwner, strings.Join(freeze.Owner, ", ")),
	}, false
}

func (m *mergeHelper) getFreezeInfo(log *logrus.Entry) (*freezeItem, error) {
	branch := m.pr.GetBase().GetRef()
	for _, v := range m.cfg.FreezeFile {
		fc, err := m.getFreezeContent(v)
		if err != nil {
			log.Errorf("get freeze file:%s, err:%s", v.toString(), err.Error())
			return nil, err
		}

		if v := fc.getFreezeItem(m.org, branch); v != nil {
			return v, nil
		}
	}

	return nil, nil
}

func (m *mergeHelper) getFreezeContent(f freezeFile) (freezeContent, error) {
	var fc freezeContent

	c, err := m.cli.GetPathContent(f.Owner, f.Repo, f.Branch, f.Path)
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

func isLabelMatched(labels sets.String, cfg *botConfig) []string {
	var reasons []string

	needs := sets.NewString(approvedLabel)
	needs.Insert(cfg.LabelsForMerge...)

	if ln := cfg.LgtmCountsRequired; ln == 1 {
		needs.Insert(lgtmLabel)
	} else {
		v := getLGTMLabelsOnPR(labels)
		if n := uint(len(v)); n < ln {
			reasons = append(reasons, fmt.Sprintf(msgNotEnoughLGTMLabel, ln, n))
		}
	}

	if v := needs.Difference(labels); v.Len() > 0 {
		reasons = append(reasons, fmt.Sprintf(
			msgMissingLabels, strings.Join(v.UnsortedList(), ", "),
		))
	}

	if len(cfg.MissingLabelsForMerge) > 0 {
		missing := sets.NewString(cfg.MissingLabelsForMerge...)
		if v := missing.Intersection(labels); v.Len() > 0 {
			reasons = append(reasons, fmt.Sprintf(
				msgInvalidLabels, strings.Join(v.UnsortedList(), ", "),
			))
		}
	}

	return reasons
}
