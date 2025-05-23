package extender

import (
	"errors"
	"github.com/Adaptix-Framework/axc2"
)

func (ex *AdaptixExtender) ExAgentGenerate(agentName string, config string, operatingSystem string, listenerWM string, listenerProfile []byte) ([]byte, string, error) {
	module, ok := ex.agentModules[agentName]
	if !ok {
		return nil, "", errors.New("module not found")
	}
	return module.F.AgentGenerate(config, operatingSystem, listenerWM, listenerProfile)
}

func (ex *AdaptixExtender) ExAgentCreate(agentName string, beat []byte) (adaptix.AgentData, error) {
	module, ok := ex.agentModules[agentName]
	if !ok {
		return adaptix.AgentData{}, errors.New("module not found")
	}
	return module.F.AgentCreate(beat)
}

func (ex *AdaptixExtender) ExAgentCommand(client string, cmdline string, agentName string, agentData adaptix.AgentData, args map[string]any) error {
	module, ok := ex.agentModules[agentName]
	if !ok {
		return errors.New("module not found")
	}
	return module.F.AgentCommand(client, cmdline, agentData, args)
}

func (ex *AdaptixExtender) ExAgentProcessData(agentData adaptix.AgentData, packedData []byte) ([]byte, error) {
	module, ok := ex.agentModules[agentData.Name]
	if !ok {
		return nil, errors.New("module not found")
	}
	return module.F.AgentProcessData(agentData, packedData)
}

func (ex *AdaptixExtender) ExAgentPackData(agentData adaptix.AgentData, maxDataSize int) ([]byte, error) {
	module, ok := ex.agentModules[agentData.Name]
	if !ok {
		return nil, errors.New("module not found")
	}
	return module.F.AgentPackData(agentData, maxDataSize)
}

func (ex *AdaptixExtender) ExAgentPivotPackData(agentName string, pivotId string, data []byte) (adaptix.TaskData, error) {
	module, ok := ex.agentModules[agentName]
	if !ok {
		return adaptix.TaskData{}, errors.New("module not found")
	}
	return module.F.AgentPivotPackData(pivotId, data)
}

func (ex *AdaptixExtender) ExAgentDownloadChangeState(agentData adaptix.AgentData, newState int, fileId string) (adaptix.TaskData, error) {
	module, ok := ex.agentModules[agentData.Name]
	if !ok {
		return adaptix.TaskData{}, errors.New("module not found")
	}

	lName, err := ex.ts.TsListenerTypeByName(agentData.Listener)
	if err != nil {
		return adaptix.TaskData{}, err
	}

	supports, ok := module.Supports[lName]
	if !ok {
		return adaptix.TaskData{}, errors.New("function DownloadChangeState is not supported")
	}
	supportConf, ok := supports[agentData.Os]
	if !ok {
		return adaptix.TaskData{}, errors.New("function DownloadChangeState is not supported")
	}
	if !supportConf.DownloadsState {
		return adaptix.TaskData{}, errors.New("function DownloadChangeState is not supported")
	}

	return module.F.AgentDownloadChangeState(agentData, newState, fileId)
}

func (ex *AdaptixExtender) ExAgentBrowserDisks(agentData adaptix.AgentData) (adaptix.TaskData, error) {
	module, ok := ex.agentModules[agentData.Name]
	if !ok {
		return adaptix.TaskData{}, errors.New("module not found")
	}

	lName, err := ex.ts.TsListenerTypeByName(agentData.Listener)
	if err != nil {
		return adaptix.TaskData{}, err
	}
	supports, ok := module.Supports[lName]
	if !ok {
		return adaptix.TaskData{}, errors.New("function BrowserDisks is not supported")
	}
	supportConf, ok := supports[agentData.Os]
	if !ok {
		return adaptix.TaskData{}, errors.New("function BrowserDisks is not supported")
	}
	if !supportConf.FileBrowserDisks {
		return adaptix.TaskData{}, errors.New("function BrowserDisks is not supported")
	}

	return module.F.AgentBrowserDisks(agentData)
}

func (ex *AdaptixExtender) ExAgentBrowserProcess(agentData adaptix.AgentData) (adaptix.TaskData, error) {
	module, ok := ex.agentModules[agentData.Name]
	if !ok {
		return adaptix.TaskData{}, errors.New("module not found")
	}

	lName, err := ex.ts.TsListenerTypeByName(agentData.Listener)
	if err != nil {
		return adaptix.TaskData{}, err
	}
	supports, ok := module.Supports[lName]
	if !ok {
		return adaptix.TaskData{}, errors.New("function ProcessBrowser is not supported")
	}
	supportConf, ok := supports[agentData.Os]
	if !ok {
		return adaptix.TaskData{}, errors.New("function ProcessBrowser is not supported")
	}
	if !supportConf.ProcessBrowser {
		return adaptix.TaskData{}, errors.New("function ProcessBrowser is not supported")
	}

	return module.F.AgentBrowserProcess(agentData)
}

func (ex *AdaptixExtender) ExAgentBrowserFiles(agentData adaptix.AgentData, path string) (adaptix.TaskData, error) {
	module, ok := ex.agentModules[agentData.Name]
	if !ok {
		return adaptix.TaskData{}, errors.New("module not found")
	}

	lName, err := ex.ts.TsListenerTypeByName(agentData.Listener)
	if err != nil {
		return adaptix.TaskData{}, err
	}
	supports, ok := module.Supports[lName]
	if !ok {
		return adaptix.TaskData{}, errors.New("function FileBrowser is not supported")
	}
	supportConf, ok := supports[agentData.Os]
	if !ok {
		return adaptix.TaskData{}, errors.New("function FileBrowser is not supported")
	}
	if !supportConf.FileBrowser {
		return adaptix.TaskData{}, errors.New("function FileBrowser is not supported")
	}

	return module.F.AgentBrowserFiles(path, agentData)
}

func (ex *AdaptixExtender) ExAgentBrowserUpload(agentData adaptix.AgentData, path string, content []byte) (adaptix.TaskData, error) {
	module, ok := ex.agentModules[agentData.Name]
	if !ok {
		return adaptix.TaskData{}, errors.New("module not found")
	}

	lName, err := ex.ts.TsListenerTypeByName(agentData.Listener)
	if err != nil {
		return adaptix.TaskData{}, err
	}
	supports, ok := module.Supports[lName]
	if !ok {
		return adaptix.TaskData{}, errors.New("function FileBrowserUpload is not supported")
	}
	supportConf, ok := supports[agentData.Os]
	if !ok {
		return adaptix.TaskData{}, errors.New("function FileBrowserUpload is not supported")
	}
	if !supportConf.FileBrowserUpload {
		return adaptix.TaskData{}, errors.New("function FileBrowserUpload is not supported")
	}

	return module.F.AgentBrowserUpload(path, content, agentData)
}

func (ex *AdaptixExtender) ExAgentBrowserDownload(agentData adaptix.AgentData, path string) (adaptix.TaskData, error) {
	module, ok := ex.agentModules[agentData.Name]
	if !ok {
		return adaptix.TaskData{}, errors.New("module not found")
	}

	lName, err := ex.ts.TsListenerTypeByName(agentData.Listener)
	if err != nil {
		return adaptix.TaskData{}, err
	}
	supports, ok := module.Supports[lName]
	if !ok {
		return adaptix.TaskData{}, errors.New("function FileBrowserDownload is not supported")
	}
	supportConf, ok := supports[agentData.Os]
	if !ok {
		return adaptix.TaskData{}, errors.New("function FileBrowserDownload is not supported")
	}
	if !supportConf.FileBrowserDownload {
		return adaptix.TaskData{}, errors.New("function FileBrowserDownload is not supported")
	}

	return module.F.AgentBrowserDownload(path, agentData)
}

func (ex *AdaptixExtender) ExAgentCtxExit(agentData adaptix.AgentData) (adaptix.TaskData, error) {
	module, ok := ex.agentModules[agentData.Name]
	if !ok {
		return adaptix.TaskData{}, errors.New("module not found")
	}

	lName, err := ex.ts.TsListenerTypeByName(agentData.Listener)
	if err != nil {
		return adaptix.TaskData{}, err
	}
	supports, ok := module.Supports[lName]
	if !ok {
		return adaptix.TaskData{}, errors.New("function SessionsMenuExit is not supported")
	}
	supportConf, ok := supports[agentData.Os]
	if !ok {
		return adaptix.TaskData{}, errors.New("function SessionsMenuExit is not supported")
	}
	if !supportConf.SessionsMenuExit {
		return adaptix.TaskData{}, errors.New("function SessionsMenuExit is not supported")
	}

	return module.F.AgentBrowserExit(agentData)
}

func (ex *AdaptixExtender) ExAgentBrowserJobKill(agentData adaptix.AgentData, jobId string) (adaptix.TaskData, error) {
	module, ok := ex.agentModules[agentData.Name]
	if !ok {
		return adaptix.TaskData{}, errors.New("module not found")
	}

	lName, err := ex.ts.TsListenerTypeByName(agentData.Listener)
	if err != nil {
		return adaptix.TaskData{}, err
	}
	supports, ok := module.Supports[lName]
	if !ok {
		return adaptix.TaskData{}, errors.New("function TasksJobKill is not supported")
	}
	supportConf, ok := supports[agentData.Os]
	if !ok {
		return adaptix.TaskData{}, errors.New("function TasksJobKill is not supported")
	}
	if !supportConf.TasksJobKill {
		return adaptix.TaskData{}, errors.New("function TasksJobKill is not supported")
	}

	return module.F.AgentBrowserJobKill(jobId)
}
