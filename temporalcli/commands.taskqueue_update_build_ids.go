package temporalcli

import (
	"fmt"

	"go.temporal.io/sdk/client"
)

func updateBuildIds(cctx *CommandContext, parent *TemporalTaskQueueCommand, options *client.UpdateWorkerBuildIdCompatibilityOptions) error {
	cl, err := parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	if err := cl.UpdateWorkerBuildIdCompatibility(cctx, options); err != nil {
		return fmt.Errorf("error updating task queue build IDs: %w", err)
	}

	cctx.Printer.Println("Successfully updated task queue build IDs")
	return nil
}

func (c *TemporalTaskQueueUpdateBuildIdsAddNewDefaultCommand) run(cctx *CommandContext, args []string) error {
	options := &client.UpdateWorkerBuildIdCompatibilityOptions{
		TaskQueue: c.TaskQueue,
		Operation: &client.BuildIDOpAddNewIDInNewDefaultSet{
			BuildID: c.BuildId,
		},
	}
	err := updateBuildIds(cctx, c.Parent.Parent, options)
	if err != nil {
		return err
	}
	return nil
}

func (c *TemporalTaskQueueUpdateBuildIdsAddNewCompatibleCommand) run(cctx *CommandContext, args []string) error {
	options := &client.UpdateWorkerBuildIdCompatibilityOptions{
		TaskQueue: c.TaskQueue,
		Operation: &client.BuildIDOpAddNewCompatibleVersion{
			BuildID:                   c.BuildId,
			ExistingCompatibleBuildID: c.ExistingCompatibleBuildId,
			MakeSetDefault:            c.SetAsDefault,
		},
	}
	err := updateBuildIds(cctx, c.Parent.Parent, options)
	if err != nil {
		return err
	}
	return nil
}

func (c *TemporalTaskQueueUpdateBuildIdsPromoteSetCommand) run(cctx *CommandContext, args []string) error {
	options := &client.UpdateWorkerBuildIdCompatibilityOptions{
		TaskQueue: c.TaskQueue,
		Operation: &client.BuildIDOpPromoteSet{
			BuildID: c.BuildId,
		},
	}
	err := updateBuildIds(cctx, c.Parent.Parent, options)
	if err != nil {
		return err
	}
	return nil
}

func (c *TemporalTaskQueueUpdateBuildIdsPromoteIdInSetCommand) run(cctx *CommandContext, args []string) error {
	options := &client.UpdateWorkerBuildIdCompatibilityOptions{
		TaskQueue: c.TaskQueue,
		Operation: &client.BuildIDOpPromoteIDWithinSet{
			BuildID: c.BuildId,
		},
	}
	err := updateBuildIds(cctx, c.Parent.Parent, options)
	if err != nil {
		return err
	}
	return nil
}
