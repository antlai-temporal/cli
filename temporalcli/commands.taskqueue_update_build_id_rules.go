package temporalcli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/sdk/client"
)

type getConflictTokenOptions struct {
	safeMode        bool
	safeModeMessage string
	taskQueue       string
	showAssignment  bool
}

func (c *TemporalTaskQueueUpdateBuildIdRulesCommand) getConflictToken(cctx *CommandContext, options *getConflictTokenOptions) (client.VersioningConflictToken, error) {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return client.VersioningConflictToken{}, err
	}
	defer cl.Close()

	rules, err := cl.GetWorkerVersioningRules(cctx, &client.GetWorkerVersioningOptions{
		TaskQueue: options.taskQueue,
	})
	if err != nil {
		return client.VersioningConflictToken{}, fmt.Errorf("unable to get versioning conflict token: %w", err)
	}

	if options.safeMode {
		// duplicate `cctx.promptYes` check to avoid printing current rules with json
		if cctx.JSONOutput {
			return client.VersioningConflictToken{}, fmt.Errorf("must bypass prompts when using JSON output")
		}
		fRules := versioningRulesToRows(rules)

		if options.showAssignment {
			cctx.Printer.Println(color.MagentaString("Current Assignment Rules:"))
			err = cctx.Printer.PrintStructured(fRules.AssignmentRules, printer.StructuredOptions{Table: &printer.TableOptions{}})
		} else {
			//!showAssigment == showRedirect
			cctx.Printer.Println(color.MagentaString("Current Redirect Rules:"))
			err = cctx.Printer.PrintStructured(fRules.RedirectRules, printer.StructuredOptions{Table: &printer.TableOptions{}})
		}
		if err != nil {
			return client.VersioningConflictToken{}, fmt.Errorf("displaying rules failed: %w", err)
		}

		yes, err := cctx.promptYes(
			fmt.Sprintf("Continue with rules update %v? y/N", options.safeModeMessage), false)
		if err != nil {
			return client.VersioningConflictToken{}, err
		} else if !yes {
			return client.VersioningConflictToken{}, fmt.Errorf("user denied confirmation")
		}
	}

	return rules.ConflictToken, nil
}

func (c *TemporalTaskQueueUpdateBuildIdRulesCommand) updateBuildIdRules(cctx *CommandContext, options *client.UpdateWorkerVersioningRulesOptions) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	rules, err := cl.UpdateWorkerVersioningRules(cctx, options)
	if err != nil {
		return fmt.Errorf("error updating task queue build ID rules: %w", err)
	}

	err = printBuildIdRules(cctx, rules)
	if err != nil {
		return err
	}

	cctx.Printer.Println("Successfully updated task queue build ID rules")
	return nil
}

func (c *TemporalTaskQueueUpdateBuildIdRulesAddRedirectRuleCommand) run(cctx *CommandContext, args []string) error {
	token, err := c.Parent.getConflictToken(cctx, &getConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "adding a redirect rule",
		taskQueue:       c.TaskQueue,
		showAssignment:  false,
	})
	if err != nil {
		return err
	}

	return c.Parent.updateBuildIdRules(cctx, &client.UpdateWorkerVersioningRulesOptions{
		TaskQueue:     c.TaskQueue,
		ConflictToken: token,
		Operation: &client.VersioningOpAddRedirectRule{
			Rule: client.VersioningRedirectRule{
				SourceBuildID: c.SourceBuildId,
				TargetBuildID: c.TargetBuildId,
			},
		},
	})
}

func (c *TemporalTaskQueueUpdateBuildIdRulesCommitBuildIdCommand) run(cctx *CommandContext, args []string) error {
	token, err := c.Parent.getConflictToken(cctx, &getConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "commiting a redirect rule",
		taskQueue:       c.TaskQueue,
		showAssignment:  true,
	})
	if err != nil {
		return err
	}

	return c.Parent.updateBuildIdRules(cctx, &client.UpdateWorkerVersioningRulesOptions{
		TaskQueue:     c.TaskQueue,
		ConflictToken: token,
		Operation: &client.VersioningOpCommitBuildID{
			TargetBuildID: c.BuildId,
			Force:         c.Force,
		},
	})
}

func (c *TemporalTaskQueueUpdateBuildIdRulesDeleteAssignmentRuleCommand) run(cctx *CommandContext, args []string) error {
	token, err := c.Parent.getConflictToken(cctx, &getConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "deleting an assignment rule",
		taskQueue:       c.TaskQueue,
		showAssignment:  true,
	})
	if err != nil {
		return err
	}

	return c.Parent.updateBuildIdRules(cctx, &client.UpdateWorkerVersioningRulesOptions{
		TaskQueue:     c.TaskQueue,
		ConflictToken: token,
		Operation: &client.VersioningOpDeleteAssignmentRule{
			RuleIndex: int32(c.RuleIndex),
			Force:     c.Force,
		},
	})
}

func (c *TemporalTaskQueueUpdateBuildIdRulesDeleteRedirectRuleCommand) run(cctx *CommandContext, args []string) error {
	token, err := c.Parent.getConflictToken(cctx, &getConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "deleting a redirect rule",
		taskQueue:       c.TaskQueue,
		showAssignment:  false,
	})
	if err != nil {
		return err
	}

	return c.Parent.updateBuildIdRules(cctx, &client.UpdateWorkerVersioningRulesOptions{
		TaskQueue:     c.TaskQueue,
		ConflictToken: token,
		Operation: &client.VersioningOpDeleteRedirectRule{
			SourceBuildID: c.SourceBuildId,
		},
	})
}

func (c *TemporalTaskQueueUpdateBuildIdRulesInsertAssignmentRuleCommand) run(cctx *CommandContext, args []string) error {
	token, err := c.Parent.getConflictToken(cctx, &getConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "inserting an assignment rule",
		taskQueue:       c.TaskQueue,
		showAssignment:  true,
	})
	if err != nil {
		return err
	}

	rule := client.VersioningAssignmentRule{
		TargetBuildID: c.BuildId,
	}
	if c.Percentage != 100 {
		rule.Ramp = &client.VersioningRampByPercentage{
			Percentage: float32(c.Percentage),
		}
	}

	return c.Parent.updateBuildIdRules(cctx, &client.UpdateWorkerVersioningRulesOptions{
		TaskQueue:     c.TaskQueue,
		ConflictToken: token,
		Operation: &client.VersioningOpInsertAssignmentRule{
			RuleIndex: int32(c.RuleIndex),
			Rule:      rule,
		},
	})
}

func (c *TemporalTaskQueueUpdateBuildIdRulesReplaceAssignmentRuleCommand) run(cctx *CommandContext, args []string) error {
	token, err := c.Parent.getConflictToken(cctx, &getConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "replacing an assignment rule",
		taskQueue:       c.TaskQueue,
		showAssignment:  true,
	})
	if err != nil {
		return err
	}

	rule := client.VersioningAssignmentRule{
		TargetBuildID: c.BuildId,
	}
	if c.Percentage != 100 {
		rule.Ramp = &client.VersioningRampByPercentage{
			Percentage: float32(c.Percentage),
		}
	}

	return c.Parent.updateBuildIdRules(cctx, &client.UpdateWorkerVersioningRulesOptions{
		TaskQueue:     c.TaskQueue,
		ConflictToken: token,
		Operation: &client.VersioningOpReplaceAssignmentRule{
			RuleIndex: int32(c.RuleIndex),
			Rule:      rule,
			Force:     c.Force,
		},
	})
}

func (c *TemporalTaskQueueUpdateBuildIdRulesReplaceRedirectRuleCommand) run(cctx *CommandContext, args []string) error {
	token, err := c.Parent.getConflictToken(cctx, &getConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "replacing a redirect rule",
		taskQueue:       c.TaskQueue,
		showAssignment:  false,
	})
	if err != nil {
		return err
	}

	return c.Parent.updateBuildIdRules(cctx, &client.UpdateWorkerVersioningRulesOptions{
		TaskQueue:     c.TaskQueue,
		ConflictToken: token,
		Operation: &client.VersioningOpReplaceRedirectRule{
			Rule: client.VersioningRedirectRule{
				SourceBuildID: c.SourceBuildId,
				TargetBuildID: c.TargetBuildId,
			},
		},
	})
}
