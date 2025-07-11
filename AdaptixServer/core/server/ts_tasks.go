package server

import (
	"AdaptixServer/core/utils/krypt"
	"AdaptixServer/core/utils/logs"
	"fmt"
	"github.com/Adaptix-Framework/axc2"
	"time"
)

func (ts *Teamserver) TsTaskRunningExists(agentId string, taskId string) bool {
	value, ok := ts.agents.Get(agentId)
	if !ok {
		logs.Error("", "TsTaskUpdate: agent %v not found", agentId)
		return false
	}
	agent, _ := value.(*Agent)

	return agent.RunningTasks.Contains(taskId)
}

func (ts *Teamserver) TsTaskCreate(agentId string, cmdline string, client string, taskData adaptix.TaskData) {
	value, ok := ts.agents.Get(agentId)
	if !ok {
		logs.Error("", "TsTaskCreate: agent %v not found", agentId)
		return
	}

	agent, _ := value.(*Agent)
	if agent.Active == false {
		return
	}

	if taskData.TaskId == "" {
		taskData.TaskId, _ = krypt.GenerateUID(8)
	}
	taskData.AgentId = agentId
	taskData.CommandLine = cmdline
	taskData.Client = client
	taskData.Computer = agent.Data.Computer
	taskData.StartDate = time.Now().Unix()
	if taskData.Completed {
		taskData.FinishDate = taskData.StartDate
	}

	taskData.User = agent.Data.Username
	if agent.Data.Impersonated != "" {
		taskData.User += fmt.Sprintf(" [%s]", agent.Data.Impersonated)
	}

	switch taskData.Type {

	case TYPE_TASK:
		if taskData.Sync {
			packet := CreateSpAgentTaskSync(taskData)
			ts.TsSyncAllClients(packet)

			packet2 := CreateSpAgentConsoleTaskSync(taskData)
			ts.TsSyncAllClients(packet2)

			agent.OutConsole.Put(packet2)
			_ = ts.DBMS.DbConsoleInsert(agentId, packet2)
		}
		agent.TasksQueue.Put(taskData)

	case TYPE_BROWSER:
		agent.TasksQueue.Put(taskData)

	case TYPE_JOB:
		if taskData.Sync {
			packet := CreateSpAgentTaskSync(taskData)
			ts.TsSyncAllClients(packet)

			packet2 := CreateSpAgentConsoleTaskSync(taskData)
			ts.TsSyncAllClients(packet2)

			agent.OutConsole.Put(packet2)
			_ = ts.DBMS.DbConsoleInsert(agentId, packet2)
		}
		agent.TasksQueue.Put(taskData)

	case TYPE_TUNNEL:
		if taskData.Sync {
			if taskData.Completed {
				agent.CompletedTasks.Put(taskData.TaskId, taskData)
			} else {
				agent.RunningTasks.Put(taskData.TaskId, taskData)
			}

			packet := CreateSpAgentTaskSync(taskData)
			ts.TsSyncAllClients(packet)

			packet2 := CreateSpAgentConsoleTaskSync(taskData)
			ts.TsSyncAllClients(packet2)

			agent.OutConsole.Put(packet2)
			_ = ts.DBMS.DbConsoleInsert(agentId, packet2)

			if taskData.Completed {
				_ = ts.DBMS.DbTaskInsert(taskData)
			}
		}

	case TYPE_PROXY_DATA:
		fmt.Println("----TYPE_PROXY_DATA----")
	//agent.TunnelQueue.Put(taskData)

	default:
		break
	}
}

func (ts *Teamserver) TsTaskUpdate(agentId string, taskData adaptix.TaskData) {
	value, ok := ts.agents.Get(agentId)
	if !ok {
		logs.Error("", "TsTaskUpdate: agent %v not found", agentId)
		return
	}
	agent, _ := value.(*Agent)

	value, ok = agent.RunningTasks.GetDelete(taskData.TaskId)
	if !ok {
		return
	}
	task, _ := value.(adaptix.TaskData)

	task.Data = []byte("")
	task.FinishDate = taskData.FinishDate
	task.Completed = taskData.Completed

	if task.Type == TYPE_JOB {
		if task.MessageType != CONSOLE_OUT_ERROR {
			task.MessageType = taskData.MessageType
		}

		var oldMessage string
		if task.Message == "" {
			oldMessage = taskData.Message
		} else {
			oldMessage = task.Message
		}

		oldText := task.ClearText

		task.Message = taskData.Message
		task.ClearText = taskData.ClearText

		packet := CreateSpAgentTaskUpdate(task)
		packet2 := CreateSpAgentConsoleTaskUpd(task)

		task.Message = oldMessage
		task.ClearText = oldText + task.ClearText

		if task.Sync {
			if task.Completed {
				agent.CompletedTasks.Put(task.TaskId, task)
			} else {
				agent.RunningTasks.Put(task.TaskId, task)
			}

			if task.Completed {
				_ = ts.DBMS.DbTaskInsert(task)
			}

			ts.TsSyncAllClients(packet)
			ts.TsSyncAllClients(packet2)

			agent.OutConsole.Put(packet2)
			_ = ts.DBMS.DbConsoleInsert(agentId, packet2)
		}

	} else if task.Type == TYPE_TUNNEL {
		var oldMessage string
		if task.Message == "" {
			oldMessage = taskData.Message
		} else {
			oldMessage = task.Message
		}
		oldText := task.ClearText

		task.MessageType = taskData.MessageType
		task.Message = taskData.Message
		task.ClearText = taskData.ClearText

		packet := CreateSpAgentTaskUpdate(task)
		packet2 := CreateSpAgentConsoleTaskUpd(task)

		task.Message = oldMessage
		task.ClearText = oldText + task.ClearText

		if task.Sync {
			if task.Completed {
				agent.CompletedTasks.Put(task.TaskId, task)
			} else {
				agent.RunningTasks.Put(task.TaskId, task)
			}

			if task.Completed {
				_ = ts.DBMS.DbTaskInsert(task)
			}

			ts.TsSyncAllClients(packet)
			ts.TsSyncAllClients(packet2)

			agent.OutConsole.Put(packet2)
			_ = ts.DBMS.DbConsoleInsert(agentId, packet2)
		}

	} else if task.Type == TYPE_TASK || task.Type == TYPE_BROWSER {
		task.MessageType = taskData.MessageType
		task.Message = taskData.Message
		task.ClearText = taskData.ClearText

		if task.Sync {
			if task.Completed {
				agent.CompletedTasks.Put(task.TaskId, task)
			} else {
				agent.RunningTasks.Put(task.TaskId, task)
			}

			if task.Completed {
				_ = ts.DBMS.DbTaskInsert(task)
			}

			packet := CreateSpAgentTaskUpdate(task)
			ts.TsSyncAllClients(packet)

			packet2 := CreateSpAgentConsoleTaskUpd(task)
			ts.TsSyncAllClients(packet2)

			agent.OutConsole.Put(packet2)
			_ = ts.DBMS.DbConsoleInsert(agentId, packet2)
		}
	}
}

func (ts *Teamserver) TsTaskDelete(agentId string, taskId string) error {
	value, ok := ts.agents.Get(agentId)
	if !ok {
		return fmt.Errorf("agent %v not found", agentId)
	}
	agent, _ := value.(*Agent)

	var task adaptix.TaskData
	for i := uint(0); i < agent.TasksQueue.Len(); i++ {
		if value, ok = agent.TasksQueue.Get(i); ok {
			task = value.(adaptix.TaskData)
			if task.TaskId == taskId {
				return fmt.Errorf("task %v in process", taskId)
			}
		}
	}

	value, ok = agent.RunningTasks.Get(taskId)
	if ok {
		return fmt.Errorf("task %v in process", taskId)
	}
	value, ok = agent.CompletedTasks.GetDelete(taskId)
	if !ok {
		return fmt.Errorf("task %v not found", taskId)
	}

	task = value.(adaptix.TaskData)
	_ = ts.DBMS.DbTaskDelete(task.TaskId, "")

	packet := CreateSpAgentTaskRemove(task)
	ts.TsSyncAllClients(packet)
	return nil
}

func (ts *Teamserver) TsTaskStop(agentId string, taskId string) error {
	value, ok := ts.agents.Get(agentId)
	if !ok {
		return fmt.Errorf("agent %v not found", agentId)
	}
	agent, _ := value.(*Agent)

	var task adaptix.TaskData
	found := false
	for i := uint(0); i < agent.TasksQueue.Len(); i++ {
		if value, ok = agent.TasksQueue.Get(i); ok {
			task = value.(adaptix.TaskData)
			if task.TaskId == taskId {
				agent.TasksQueue.Delete(i)
				found = true
				break
			}
		}
	}

	if found {
		packet := CreateSpAgentTaskRemove(task)
		ts.TsSyncAllClients(packet)
		return nil
	}

	value, ok = agent.RunningTasks.Get(taskId)
	if !ok {
		return nil
	}
	if value.(adaptix.TaskData).Type != TYPE_JOB {
		return fmt.Errorf("task %v in process", taskId)
	}

	taskData, err := ts.Extender.ExAgentBrowserJobKill(agent.Data, taskId)
	if err != nil {
		return err
	}
	ts.TsTaskCreate(agent.Data.Id, "job kill "+taskId, "", taskData)
	return nil
}

///// Get Tasks

func (ts *Teamserver) TsTaskGetAvailableAll(agentId string, availableSize int) ([]adaptix.TaskData, error) {
	value, ok := ts.agents.Get(agentId)
	if !ok {
		return nil, fmt.Errorf("TsTaskQueueGetAvailable: agent %v not found", agentId)
	}
	agent, _ := value.(*Agent)

	var tasks []adaptix.TaskData
	tasksSize := 0

	/// TASKS QUEUE

	var sendTasks []string
	for i := uint(0); i < agent.TasksQueue.Len(); i++ {
		value, ok = agent.TasksQueue.Get(i)
		if ok {
			taskData := value.(adaptix.TaskData)
			if tasksSize+len(taskData.Data) < availableSize {
				tasks = append(tasks, taskData)
				if taskData.Sync || taskData.Type == TYPE_BROWSER {
					agent.RunningTasks.Put(taskData.TaskId, taskData)
				}
				agent.TasksQueue.Delete(i)
				i--
				sendTasks = append(sendTasks, taskData.TaskId)
				tasksSize += len(taskData.Data)
			} else {
				break
			}
		} else {
			break
		}
	}
	if len(sendTasks) > 0 {
		packet := CreateSpAgentTaskSend(sendTasks)
		ts.TsSyncAllClients(packet)
	}

	for i := uint(0); i < agent.TunnelConnectTasks.Len(); i++ {
		value, ok = agent.TunnelConnectTasks.Get(i)
		if ok {
			taskData := value.(adaptix.TaskData)
			if tasksSize+len(taskData.Data) < availableSize {
				tasks = append(tasks, taskData)
				agent.TunnelConnectTasks.Delete(i)
				i--
				tasksSize += len(taskData.Data)
			} else {
				break
			}
		} else {
			break
		}
	}

	/// TUNNELS QUEUE

	for i := uint(0); i < agent.TunnelQueue.Len(); i++ {
		value, ok = agent.TunnelQueue.Get(i)
		if ok {
			taskDataTunnel := value.(adaptix.TaskDataTunnel)
			if tasksSize+len(taskDataTunnel.Data.Data) < availableSize {
				tasks = append(tasks, taskDataTunnel.Data)
				agent.TunnelQueue.Delete(i)
				i--
				tasksSize += len(taskDataTunnel.Data.Data)
			} else {
				break
			}
		} else {
			break
		}
	}

	/// PIVOTS QUEUE

	for i := uint(0); i < agent.PivotChilds.Len(); i++ {
		value, ok = agent.PivotChilds.Get(i)
		if ok {
			pivotData := value.(*adaptix.PivotData)
			lostSize := availableSize - tasksSize
			if availableSize > 0 {
				data, err := ts.TsAgentGetHostedTasksAll(pivotData.ChildAgentId, lostSize)
				if err != nil {
					continue
				}
				pivotTaskData, err := ts.Extender.ExAgentPivotPackData(agent.Data.Name, pivotData.PivotId, data)
				if err != nil {
					continue
				}
				tasks = append(tasks, pivotTaskData)
				tasksSize += len(pivotTaskData.Data)
			}
		} else {
			break
		}
	}

	return tasks, nil
}

func (ts *Teamserver) TsTaskGetAvailableTasks(agentId string, availableSize int) ([]adaptix.TaskData, int, error) {
	value, ok := ts.agents.Get(agentId)
	if !ok {
		return nil, 0, fmt.Errorf("TsTaskQueueGetAvailable: agent %v not found", agentId)
	}
	agent, _ := value.(*Agent)

	var tasks []adaptix.TaskData
	tasksSize := 0

	/// TASKS QUEUE

	var sendTasks []string
	for i := uint(0); i < agent.TasksQueue.Len(); i++ {
		value, ok = agent.TasksQueue.Get(i)
		if ok {
			taskData := value.(adaptix.TaskData)
			if tasksSize+len(taskData.Data) < availableSize {
				tasks = append(tasks, taskData)
				if taskData.Sync || taskData.Type == TYPE_BROWSER {
					agent.RunningTasks.Put(taskData.TaskId, taskData)
				}
				agent.TasksQueue.Delete(i)
				i--
				sendTasks = append(sendTasks, taskData.TaskId)
				tasksSize += len(taskData.Data)
			} else {
				break
			}
		} else {
			break
		}
	}

	if len(sendTasks) > 0 {
		packet := CreateSpAgentTaskSend(sendTasks)
		ts.TsSyncAllClients(packet)
	}

	for i := uint(0); i < agent.TunnelConnectTasks.Len(); i++ {
		value, ok = agent.TunnelConnectTasks.Get(i)
		if ok {
			taskData := value.(adaptix.TaskData)
			if tasksSize+len(taskData.Data) < availableSize {
				tasks = append(tasks, taskData)
				agent.TunnelConnectTasks.Delete(i)
				i--
				//sendTasks = append(sendTasks, taskData.TaskId)
				tasksSize += len(taskData.Data)
			} else {
				break
			}
		} else {
			break
		}
	}

	return tasks, tasksSize, nil
}

/// Get Pivot Tasks

func (ts *Teamserver) TsTasksPivotExists(agentId string, first bool) bool {
	value, ok := ts.agents.Get(agentId)
	if !ok {
		return false
	}
	agent := value.(*Agent)

	if !first {
		if agent.TasksQueue.Len() > 0 || agent.TunnelQueue.Len() > 0 {
			return true
		}
	}

	for i := uint(0); i < agent.PivotChilds.Len(); i++ {
		value, ok = agent.PivotChilds.Get(i)
		if ok {
			pivotData := value.(*adaptix.PivotData)
			if ts.TsTasksPivotExists(pivotData.ChildAgentId, false) {
				return true
			}
		}
	}
	return false
}

func (ts *Teamserver) TsTaskGetAvailablePivotAll(agentId string, availableSize int) ([]adaptix.TaskData, error) {
	value, ok := ts.agents.Get(agentId)
	if !ok {
		return nil, fmt.Errorf("TsTaskQueueGetAvailable: agent %v not found", agentId)
	}
	agent, _ := value.(*Agent)

	var tasks []adaptix.TaskData
	tasksSize := 0

	/// PIVOTS QUEUE

	for i := uint(0); i < agent.PivotChilds.Len(); i++ {
		value, ok = agent.PivotChilds.Get(i)
		if ok {
			pivotData := value.(*adaptix.PivotData)
			lostSize := availableSize - tasksSize
			if availableSize > 0 {
				data, err := ts.TsAgentGetHostedTasksAll(pivotData.ChildAgentId, lostSize)
				if err != nil {
					continue
				}
				pivotTaskData, err := ts.Extender.ExAgentPivotPackData(agent.Data.Name, pivotData.PivotId, data)
				if err != nil {
					continue
				}
				tasks = append(tasks, pivotTaskData)
				tasksSize += len(pivotTaskData.Data)
			}
		} else {
			break
		}
	}

	return tasks, nil
}
