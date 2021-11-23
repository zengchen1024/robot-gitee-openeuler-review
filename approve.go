package main

import (
	"fmt"
	"regexp"

	sdk "gitee.com/openeuler/go-gitee/gitee"
	"github.com/opensourceways/community-robot-lib/giteeclient"
	"github.com/sirupsen/logrus"
)

const approvedLabel = "approved"

var (
	regAddApprove    = regexp.MustCompile(`(?mi)^/approve\s*$`)
	regRemoveApprove = regexp.MustCompile(`(?mi)^/approve cancel\s*$`)
)

func (bot *robot) handleApprove(e *sdk.NoteEvent, cfg *botConfig, log *logrus.Entry) error {
	ne := giteeclient.NewPRNoteEvent(e)

	if !ne.IsPullRequest() || !ne.IsPROpen() || !ne.IsCreatingCommentEvent() {
		return nil
	}

	if regAddApprove.MatchString(ne.GetComment()) {
		return bot.AddApprove(cfg, ne, log)
	}

	if regRemoveApprove.MatchString(ne.GetComment()) {
		return bot.removeApprove(cfg, ne, log)
	}

	return nil
}

func (bot *robot) AddApprove(cfg *botConfig, e giteeclient.PRNoteEvent, log *logrus.Entry) error {
	pr := e.GetPRInfo()
	commenter := e.GetCommenter()

	v, err := bot.hasPermission(commenter, pr, cfg, log)
	if err != nil {
		return err
	}

	if !v {
		return bot.cli.CreatePRComment(pr.Org, pr.Repo, pr.Number, fmt.Sprintf(
			commentNoPermissionForLabel, commenter, "add", approvedLabel,
		))
	}

	return bot.cli.AddPRLabel(pr.Org, pr.Repo, pr.Number, approvedLabel)
}

func (bot *robot) removeApprove(cfg *botConfig, e giteeclient.PRNoteEvent, log *logrus.Entry) error {
	pr := e.GetPRInfo()
	commenter := e.GetCommenter()

	v, err := bot.hasPermission(commenter, pr, cfg, log)
	if err != nil {
		return err
	}

	if !v {
		return bot.cli.CreatePRComment(pr.Org, pr.Repo, pr.Number, fmt.Sprintf(
			commentNoPermissionForLabel, commenter, "remove", approvedLabel,
		))
	}

	return bot.cli.RemovePRLabel(pr.Org, pr.Repo, pr.Number, approvedLabel)
}
