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

	pr := ne.PullRequest
	org, repo := ne.GetOrgRep()

	if r := canMerge(pr.Mergeable, ne.GetPRLabels(), cfg); len(r) > 0 {
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
		org, repo, ne.GetPRNumber(), string(cfg.MergeMethod),
	)
}

func (bot *robot) tryMerge(e *sdk.PullRequestEvent, cfg *botConfig) error {
	if giteeclient.GetPullRequestAction(e) != giteeclient.PRActionUpdatedLabel {
		return nil
	}

	pr := e.PullRequest
	info := giteeclient.GetPRInfoByPREvent(e)

	if r := canMerge(pr.Mergeable, info.Labels, cfg); len(r) > 0 {
		return nil
	}

	return bot.mergePR(
		pr.NeedReview || pr.NeedTest,
		info.Org, info.Repo, info.Number, string(cfg.MergeMethod),
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

func canMerge(mergeable bool, labels sets.String, cfg *botConfig) []string {
	if !mergeable {
		return []string{msgPRConflicts}
	}

	reasons := []string{}

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
