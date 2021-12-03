package main

import (
	"fmt"
	"regexp"
	"strings"

	sdk "gitee.com/openeuler/go-gitee/gitee"
	"github.com/opensourceways/community-robot-lib/giteeclient"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	msgPRConflicts        = "PR conflicts to the target branch."
	msgMissingLabels      = "PR does not have these lables: %s"
	msgInvalidLabels      = "PR should remove these labels: %s"
	msgNotEnoughLGTMLabel = "PR needs %d lgtm labels and now gets %d"
	msgFrozenWithOwner    = "PR merge target has been frozen, and can merge only by branch owners: %s"
)

var regCheckPr = regexp.MustCompile(`(?mi)^/check-pr\s*$`)

func (bot *robot) handleCheckPR(e *sdk.NoteEvent, cfg *botConfig) error {
	ne := giteeclient.NewPRNoteEvent(e)

	if !ne.IsPullRequest() ||
		!ne.IsPROpen() ||
		!ne.IsCreatingCommentEvent() ||
		!regCheckPr.MatchString(ne.GetComment()) {
		return nil
	}

	org, repo := ne.GetOrgRep()
	pr := e.GetPullRequest()

	freeze, err := bot.getFreezeInfo(org, pr.GetBase().GetRef(), cfg.FreezeFile)
	if err != nil {
		return err
	}

	if r := canMerge(pr.Mergeable, ne.GetPRLabels(), cfg, freeze.getFrozenMsg(ne.GetCommenter())); len(r) > 0 {
		return bot.cli.CreatePRComment(
			org, repo, ne.GetPRNumber(),
			fmt.Sprintf(
				"@%s , this pr is not mergeable and the reasons are below:\n%s",
				ne.GetCommenter(), strings.Join(r, "\n"),
			),
		)
	}

	return bot.mergePR(
		pr.NeedReview || pr.NeedTest,
		org, repo, pr.Number, string(cfg.MergeMethod),
	)
}

func (bot *robot) handleLabelUpdate(e *sdk.PullRequestEvent, cfg *botConfig) error {
	if giteeclient.GetPullRequestAction(e) != giteeclient.PRActionUpdatedLabel {
		return nil
	}

	org, repo := giteeclient.GetOwnerAndRepoByPREvent(e)
	pr := e.GetPullRequest()

	freeze, err := bot.getFreezeInfo(org, pr.GetBase().GetRef(), cfg.FreezeFile)
	if err != nil {
		return err
	}

	return bot.tryMerge(org, repo, pr, cfg, freeze.getFrozenMsg())
}

func (bot *robot) tryMerge(org, repo string, pr *sdk.PullRequestHook, cfg *botConfig, isFreeze func() string) error {
	if r := canMerge(pr.Mergeable, nil, cfg, isFreeze); len(r) > 0 {
		return nil
	}

	return bot.mergePR(
		pr.NeedReview || pr.NeedTest,
		org, repo, pr.Number, string(cfg.MergeMethod),
	)
}

func (bot *robot) mergePR(needReviewOrTest bool, org, repo string, number int32, method string) error {
	if needReviewOrTest {
		v := int32(0)
		p := sdk.PullRequestUpdateParam{
			AssigneesNumber: &v,
			TestersNumber:   &v,
		}

		if _, err := bot.cli.UpdatePullRequest(org, repo, number, p); err != nil {
			return err
		}
	}

	return bot.cli.MergePR(
		org, repo, number,
		sdk.PullRequestMergePutParam{
			MergeMethod: method,
		},
	)
}

func canMerge(mergeable bool, labels sets.String, cfg *botConfig, isFreeze func() string) []string {
	var reasons []string

	if !mergeable {
		reasons = append(reasons, msgPRConflicts)
	}

	if r := isLabelMatched(labels, cfg); len(r) > 0 {
		reasons = append(reasons, r...)
	}

	if r := isFreeze(); r != "" {
		reasons = append(reasons, r)
	}

	return reasons
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
