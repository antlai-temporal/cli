package temporalcli_test

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

func (s *SharedServerSuite) TestTaskQueue_Rules_BuildId() {
	type assignmentRowType struct {
		TargetBuildID string    `json:"targetBuildID"`
		Percentage    float32   `json:"percentage"`
		CreateTime    time.Time `json:"-"`
	}

	type redirectRowType struct {
		SourceBuildID string    `json:"sourceBuildID"`
		TargetBuildID string    `json:"targetBuildID"`
		CreateTime    time.Time `json:"-"`
	}

	type formattedRulesType struct {
		AssignmentRules []assignmentRowType `json:"assignmentRules"`
		RedirectRules   []redirectRowType   `json:"redirectRules"`
	}

	buildIdTaskQueue := uuid.NewString()

	res := s.Execute(
		"task-queue", "get-build-id-rules",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--output", "json",
	)
	s.NoError(res.Err)

	var jsonOut formattedRulesType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal(formattedRulesType{}, jsonOut)

	res = s.Execute(
		"task-queue", "update-build-id-rules", "insert-assignment-rule",
		"--build-id", "id1",
		"-y",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--output", "json",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"task-queue", "update-build-id-rules", "insert-assignment-rule",
		"--build-id", "id2",
		"--percentage", "10",
		"--rule-index", "0",
		"-y",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--output", "json",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"task-queue", "update-build-id-rules", "replace-assignment-rule",
		"--build-id", "id2",
		"--percentage", "40",
		"--rule-index", "0",
		"-y",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--output", "json",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"task-queue", "update-build-id-rules", "insert-assignment-rule",
		"--build-id", "id3",
		"--percentage", "10",
		"--rule-index", "100",
		"-y",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--output", "json",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"task-queue", "update-build-id-rules", "delete-assignment-rule",
		"--rule-index", "2",
		"-y",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"task-queue", "update-build-id-rules", "add-redirect-rule",
		"--source-build-id", "id1",
		"--target-build-id", "id3",
		"-y",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--output", "json",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"task-queue", "update-build-id-rules", "add-redirect-rule",
		"--source-build-id", "id3",
		"--target-build-id", "id4",
		"-y",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"task-queue", "update-build-id-rules", "replace-redirect-rule",
		"--source-build-id", "id3",
		"--target-build-id", "id5",
		"-y",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--output", "json",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"task-queue", "update-build-id-rules", "delete-redirect-rule",
		"--source-build-id", "id1",
		"-y",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--output", "json",
	)
	s.NoError(res.Err)

	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal(formattedRulesType{
		AssignmentRules: []assignmentRowType{
			{
				TargetBuildID: "id2",
				Percentage:    40.0,
			},
			{
				TargetBuildID: "id1",
				Percentage:    100.0,
			},
		},
		RedirectRules: []redirectRowType{
			{
				SourceBuildID: "id3",
				TargetBuildID: "id5",
			},
		},
	}, jsonOut)

	// Plain output

	res = s.Execute(
		"task-queue", "get-build-id-rules",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
	)
	s.NoError(res.Err)

	s.ContainsOnSameLine(res.Stdout.String(), "0", "id2", "40", "now")
	s.ContainsOnSameLine(res.Stdout.String(), "1", "id1", "100", "now")
	s.ContainsOnSameLine(res.Stdout.String(), "id3", "id5", "now")

	// Safe mode

	s.CommandHarness.Stdin.WriteString("y\n")
	res = s.Execute(
		"task-queue", "update-build-id-rules", "replace-redirect-rule",
		"--source-build-id", "id3",
		"--target-build-id", "id9",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--output", "json",
	)
	s.Error(res.Err) // json output needs needs autoconfirm

	s.CommandHarness.Stdin.WriteString("y\n")
	res = s.Execute(
		"task-queue", "update-build-id-rules", "replace-redirect-rule",
		"--source-build-id", "id3",
		"--target-build-id", "id9",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
	)
	s.NoError(res.Err)
	// Shown before replacing
	s.ContainsOnSameLine(res.Stdout.String(), "id3", "id5", "now")
	// Shown after replacing
	s.ContainsOnSameLine(res.Stdout.String(), "id3", "id9", "now")

	// Commit

	res = s.Execute(
		"task-queue", "update-build-id-rules", "commit-build-id",
		"--build-id", "id2",
		"--force",
		"-y",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--output", "json",
	)

	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal(formattedRulesType{
		AssignmentRules: []assignmentRowType{
			{
				TargetBuildID: "id2",
				Percentage:    100.0,
			},
		},
		RedirectRules: []redirectRowType{
			{
				SourceBuildID: "id3",
				TargetBuildID: "id9",
			},
		},
	}, jsonOut)
}
