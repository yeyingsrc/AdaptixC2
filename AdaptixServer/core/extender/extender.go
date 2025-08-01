package extender

import (
	"AdaptixServer/core/utils/logs"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
)

func NewExtender(teamserver Teamserver) *AdaptixExtender {
	return &AdaptixExtender{
		ts:              teamserver,
		listenerModules: make(map[string]ExtListener),
		agentModules:    make(map[string]ExtAgent),
	}
}

func (ex *AdaptixExtender) LoadPlugins(extenderFiles []string) {

	for _, path := range extenderFiles {

		_, err := os.Stat(path)
		if err != nil {
			logs.Error("", "Config %s not found", path)
			continue
		}
		config_data, err := os.ReadFile(path)
		if err != nil {
			logs.Error("", "Read file %s error: %s", path, err.Error())
			continue
		}

		var config_map map[string]any
		err = json.Unmarshal(config_data, &config_map)
		if err != nil {
			logs.Error("", "Error config %s parse: %s", path, err.Error())
			continue
		}
		extender_type, ok_type := config_map["extender_type"].(string)
		if !ok_type {
			logs.Error("", "Error config %s parse: extender_type not found", path)
			continue
		}

		if extender_type == "listener" {
			ex.LoadPluginListener(path, config_data)
		} else if extender_type == "agent" {
			ex.LoadPluginAgent(path, config_data)
		} else {
			logs.Error("", "Unknown extender_type in %s", path)
		}
	}
}

func (ex *AdaptixExtender) LoadPluginListener(config_path string, config_data []byte) {
	var configListener ExConfigListener
	err := json.Unmarshal(config_data, &configListener)
	if err != nil {
		logs.Error("", "Error config parse: %s", err.Error())
		return
	}

	plugin_path := filepath.Dir(config_path) + "/" + configListener.ExtenderFile
	plug, err := plugin.Open(plugin_path)
	if err != nil {
		logs.Error("", "failed to open plugin %s: %s", plugin_path, err.Error())
		return
	}

	sym, err := plug.Lookup("InitPlugin")
	if err != nil {
		logs.Error("", "failed to find InitPlugin in %s: %s", plugin_path, err.Error())
		return
	}

	pl_InitPlugin, ok := sym.(func(ts any, moduleDir string, listenerDir string) any)
	if !ok {
		logs.Error("", "unexpected signature from InitPlugin in %s", plugin_path)
		return
	}

	module, ok := pl_InitPlugin(ex.ts, filepath.Dir(plugin_path), logs.RepoLogsInstance.ListenerPath).(ExtListener)
	if !ok {
		logs.Error("", "plugin %s does not implement the ExtListener interface", plugin_path)
		return
	}

	ax_path := filepath.Dir(config_path) + "/" + configListener.AxFile
	ax_content, err := os.ReadFile(ax_path)
	if err != nil {
		logs.Error("", "failed to read ax file %s: %s", ax_path, err.Error())
		return
	}

	listenerInfo := ListenerInfo{
		Name:     configListener.ListenerName,
		Type:     configListener.ListenerType,
		Protocol: configListener.Protocol,
		AX:       string(ax_content),
	}

	err = ex.ts.TsListenerReg(listenerInfo)
	if err != nil {
		logs.Error("", "plugin %s does not registered: %s", plugin_path, err.Error())
		return
	}

	listenerFN := fmt.Sprintf("%v/%v/%v", listenerInfo.Type, listenerInfo.Protocol, listenerInfo.Name)
	ex.listenerModules[listenerFN] = module
}

func (ex *AdaptixExtender) LoadPluginAgent(config_path string, config_data []byte) {
	var configAgent ExConfigAgent
	err := json.Unmarshal(config_data, &configAgent)
	if err != nil {
		logs.Error("", "Error config parse: %s", err.Error())
		return
	}

	ax_path := filepath.Dir(config_path) + "/" + configAgent.AxFile
	ax_content, err := os.ReadFile(ax_path)
	if err != nil {
		logs.Error("", "failed to read ax file %s: %s", ax_path, err.Error())
		return
	}

	agentInfo := AgentInfo{
		Name:      configAgent.AgentName,
		Watermark: configAgent.AgentWatermark,
		AX:        string(ax_content),
		Listeners: configAgent.Listeners,
	}

	plugin_path := filepath.Dir(config_path) + "/" + configAgent.ExtenderFile
	plug, err := plugin.Open(plugin_path)
	if err != nil {
		logs.Error("", "failed to open plugin %s: %s", plugin_path, err.Error())
		return
	}

	sym, err := plug.Lookup("InitPlugin")
	if err != nil {
		logs.Error("", "failed to find InitPlugin in %s: %s", plugin_path, err.Error())
		return
	}

	pl_InitPlugin, ok := sym.(func(ts any, moduleDir string, watermark string) any)
	if !ok {
		logs.Error("", "unexpected signature from InitPlugin in %s", plugin_path)
		return
	}

	module, ok := pl_InitPlugin(ex.ts, filepath.Dir(plugin_path), agentInfo.Watermark).(ExtAgent)
	if !ok {
		logs.Error("", "plugin %s does not implement the ExtAgent interface", plugin_path)
		return
	}

	err = ex.ts.TsAgentReg(agentInfo)
	if err != nil {
		logs.Error("", "plugin %s does not registered: %s", plugin_path, err.Error())
		return
	}

	ex.agentModules[agentInfo.Name] = module
}
