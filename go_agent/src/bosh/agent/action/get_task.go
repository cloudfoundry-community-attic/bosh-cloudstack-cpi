package action

import (
	boshtask "bosh/agent/task"
	bosherr "bosh/errors"
	"encoding/json"
	"errors"
	"fmt"
)

type getTaskAction struct {
	taskService boshtask.Service
}

func newGetTask(taskService boshtask.Service) (getTask getTaskAction) {
	getTask.taskService = taskService
	return
}

func (a getTaskAction) Run(payloadBytes []byte) (value interface{}, err error) {
	taskId, err := parseTaskId(payloadBytes)
	if err != nil {
		err = bosherr.WrapError(err, "Error finding task")
		return
	}

	task, found := a.taskService.FindTask(taskId)
	if !found {
		err = errors.New(fmt.Sprintf("Task with id %s could not be found", taskId))
		return
	}

	type valueType struct {
		AgentTaskId string      `json:"agent_task_id"`
		State       string      `json:"state"`
		Value       interface{} `json:"value,omitempty"`
	}

	value = valueType{
		AgentTaskId: task.Id,
		State:       string(task.State),
		Value:       task.Value,
	}
	return
}

func parseTaskId(payloadBytes []byte) (taskId string, err error) {
	var payload struct {
		Arguments []string
	}
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return
	}

	if len(payload.Arguments) == 0 {
		err = errors.New("not enough arguments")
		return
	}

	taskId = payload.Arguments[0]
	return
}
