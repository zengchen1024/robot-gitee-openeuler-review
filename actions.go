package main

import (
	"fmt"

	sdk "gitee.com/openeuler/go-gitee/gitee"
	"github.com/opensourceways/community-robot-lib/giteeclient"
)

const (
	retestCommand     = "/retest"
	msgNotSetReviewer = "**@%s** Thank you for submitting a PullRequest. It is detected that you have not set a reviewer, please set a one."
)

func (bot *robot) doRetest(e *sdk.PullRequestEvent) error {
	if giteeclient.GetPullRequestAction(e) != giteeclient.PRActionChangedSourceBranch {
		return nil
	}

	pr := giteeclient.GetPRInfoByPREvent(e)

	return bot.cli.CreatePRComment(pr.Org, pr.Repo, pr.Number, retestCommand)
}

func (bot *robot) checkReviewer(e *sdk.PullRequestEvent, cfg *botConfig) error {
	if cfg.UnableCheckingReviewerForPR || giteeclient.GetPullRequestAction(e) != giteeclient.PRActionOpened {
		return nil
	}

	if e.GetPullRequest() != nil && len(e.GetPullRequest().Assignees) > 0 {
		return nil
	}

	pr := giteeclient.GetPRInfoByPREvent(e)

	return bot.cli.CreatePRComment(pr.Org, pr.Repo, pr.Number, fmt.Sprintf(msgNotSetReviewer, pr.Author))
}
