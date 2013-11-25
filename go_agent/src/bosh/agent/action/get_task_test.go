package action

import (
	boshtask "bosh/agent/task"
	boshassert "bosh/assert"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetTaskRunReturnsAFailedTask(t *testing.T) {
	settings, platform, blobstore, taskService := getFakeFactoryDependencies()
	taskService.Tasks = map[string]boshtask.Task{
		"57": boshtask.Task{
			Id:    "found-57-id",
			State: boshtask.TaskStateFailed,
		},
	}

	factory := NewFactory(settings, platform, blobstore, taskService)
	getTask := factory.Create("get_task")

	taskValue, err := getTask.Run([]byte(`{"arguments":["57"]}`))
	assert.NoError(t, err)

	boshassert.MatchesJsonString(t, taskValue, `{"agent_task_id":"found-57-id","state":"failed"}`)
}

func TestGetTaskRunReturnsASuccessfulTask(t *testing.T) {
	settings, platform, blobstore, taskService := getFakeFactoryDependencies()
	taskService.Tasks = map[string]boshtask.Task{
		"57": boshtask.Task{
			Id:    "found-57-id",
			State: boshtask.TaskStateDone,
			Value: "some-task-value",
		},
	}

	factory := NewFactory(settings, platform, blobstore, taskService)
	getTask := factory.Create("get_task")

	taskValue, err := getTask.Run([]byte(`{"arguments":["57"]}`))
	assert.NoError(t, err)

	boshassert.MatchesJsonString(t, taskValue, `{"agent_task_id":"found-57-id","state":"done","value":"some-task-value"}`)
}

func TestGetTaskRunWhenTaskIsNotFound(t *testing.T) {
	settings, platform, blobstore, taskService := getFakeFactoryDependencies()
	taskService.Tasks = map[string]boshtask.Task{}

	factory := NewFactory(settings, platform, blobstore, taskService)
	getTask := factory.Create("get_task")

	_, err := getTask.Run([]byte(`{"arguments":["57"]}`))
	assert.Error(t, err)
	assert.Equal(t, "Task with id 57 could not be found", err.Error())
}

func TestGetTaskRunWhenPayloadDoesNotHaveTaskId(t *testing.T) {
	settings, platform, blobstore, taskService := getFakeFactoryDependencies()
	taskService.Tasks = map[string]boshtask.Task{}

	factory := NewFactory(settings, platform, blobstore, taskService)
	getTask := factory.Create("get_task")

	_, err := getTask.Run([]byte(`{"arguments":[]}`))
	assert.Error(t, err)
	assert.Equal(t, "Error finding task: not enough arguments", err.Error())
}
