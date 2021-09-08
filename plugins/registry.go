// Copyright (c) 2021 OceanBase
// obagent is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
// http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package plugins

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

var inputManager *InputManager
var inputManagerOnce sync.Once
var processorManager *ProcessorManager
var processorManagerOnce sync.Once
var outputManager *OutputManager
var outputManagerOnce sync.Once
var exporterManager *ExporterManager
var exporterManagerOnce sync.Once

//GetInputManager get input manager singleton
func GetInputManager() *InputManager {
	inputManagerOnce.Do(func() {
		inputManager = &InputManager{
			Registry: make(map[string]func() Input),
		}
	})
	return inputManager
}

//GetProcessorManager get processor manager singleton
func GetProcessorManager() *ProcessorManager {
	processorManagerOnce.Do(func() {
		processorManager = &ProcessorManager{
			Registry: make(map[string]func() Processor),
		}
	})

	return processorManager
}

//GetOutputManager get output manager singleton
func GetOutputManager() *OutputManager {
	outputManagerOnce.Do(func() {
		outputManager = &OutputManager{
			Registry: make(map[string]func() Output),
		}
	})
	return outputManager
}

//GetExporterManager get exporter manager singleton
func GetExporterManager() *ExporterManager {
	exporterManagerOnce.Do(func() {
		exporterManager = &ExporterManager{
			Registry: make(map[string]func() Exporter),
		}
	})

	return exporterManager
}

//InputManager responsible for managing and creating input plugin instances
type InputManager struct {
	Registry map[string]func() Input
}

//ProcessorManager responsible for managing and creating processor plugin instances
type ProcessorManager struct {
	Registry map[string]func() Processor
}

//OutputManager responsible for managing and creating output plugin instances
type OutputManager struct {
	Registry map[string]func() Output
}

//ExporterManager responsible for managing and creating exporter plugin instances
type ExporterManager struct {
	Registry map[string]func() Exporter
}

//Register add the input plugin with the manager
func (m *InputManager) Register(name string, f func() Input) {
	_, exist := m.Registry[name]
	if exist {
		panic(fmt.Sprintf("input plugin %s already registered", name))
	}
	m.Registry[name] = f
}

//Register add the processor plugin with the manager
func (m *ProcessorManager) Register(name string, f func() Processor) {
	_, exist := m.Registry[name]
	if exist {
		panic(fmt.Sprintf("processor plugin %s already registered", name))
	}
	m.Registry[name] = f
}

//Register add the output plugin with the manager
func (m *OutputManager) Register(name string, f func() Output) {
	_, exist := m.Registry[name]
	if exist {
		panic(fmt.Sprintf("output plugin %s already registered", name))
	}
	m.Registry[name] = f
}

//Register add the exporter plugin with the manager
func (m *ExporterManager) Register(name string, f func() Exporter) {
	_, exist := m.Registry[name]
	if exist {
		panic(fmt.Sprintf("exporter plugin %s already registered", name))
	}
	m.Registry[name] = f
}

//GetPlugin get input by name
func (m *InputManager) GetPlugin(name string) (Input, error) {
	f, exist := m.Registry[name]
	if !exist {
		return nil, errors.Errorf("input plugin %s not exist", name)
	}
	return f(), nil
}

//GetPlugin get processor by name
func (m *ProcessorManager) GetPlugin(name string) (Processor, error) {
	f, exist := m.Registry[name]
	if !exist {
		return nil, errors.Errorf("processor plugin %s not exist", name)
	}
	return f(), nil
}

//GetPlugin get output by name
func (m *OutputManager) GetPlugin(name string) (Output, error) {
	f, exist := m.Registry[name]
	if !exist {
		return nil, errors.Errorf("output plugin %s not exist", name)
	}
	return f(), nil
}

//GetPlugin get exporter by name
func (m *ExporterManager) GetPlugin(name string) (Exporter, error) {
	f, exist := m.Registry[name]
	if !exist {
		return nil, errors.Errorf("exporter plugin %s not exist", name)
	}
	return f(), nil
}
