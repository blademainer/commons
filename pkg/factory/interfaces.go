package factory

import (
	"encoding/json"
	"fmt"
	"github.com/blademainer/commons/pkg/logger"
	"reflect"
)

type Factory struct {
	executorMap                      map[string]InstanceExecutorFunc
	executorTypeAndExecutorMap       map[string]interface{}
	executorTypeAndConfigInstanceMap map[string]interface{}
}

func InitFactory() *Factory {
	factory := &Factory{}
	factory.executorMap = make(map[string]InstanceExecutorFunc)
	factory.executorTypeAndExecutorMap = make(map[string]interface{})
	factory.executorTypeAndConfigInstanceMap = make(map[string]interface{})
	return factory
}

type Config struct {
	Name               string      `json:"name"`
	FactoryConfigValue interface{} `json:"config"`
}

//type Executor interface {
//	ExecuteFunc() func() interface{}
//}

// 初始化executor的函数
type InstanceExecutorFunc func(configInstance interface{}) (interface{}, error)

// 注册executor类型
//
// executorType: 队列类型
//
// instanceFunc: 初始化executor的函数
//
// configInstance: 需要反序列化配置为struct的对象。管理器在调用`instanceFunc`之前会将config字符串反序列化为`configInstance`
//
func (f *Factory) RegisterExecutor(executorType string, instanceFunc InstanceExecutorFunc, configInstance interface{}) {
	if configInstance != nil && reflect.TypeOf(configInstance).Kind() != reflect.Ptr {
		logger.Log.Errorf("Config instance is not ptr kind!")
		panic("Config instance is not ptr kind!")
	}
	logger.Log.Infof("Register executor: %v with instanceFunc: %v", executorType, instanceFunc)
	f.executorMap[executorType] = instanceFunc
	f.executorTypeAndConfigInstanceMap[executorType] = configInstance
}

func (f *Factory) GetExecutors() []interface{} {
	executors := make([]interface{}, 0)
	for _, executor := range f.executorTypeAndExecutorMap {
		executors = append(executors, executor)
	}
	return executors
}

func (f *Factory) GetExecutor(executorType string) interface{} {
	if executor, exists := f.executorTypeAndExecutorMap[executorType]; !exists {
		return nil
	} else {
		return executor
	}
}

func (f *Factory) InstanceExecutor(config Config) (executor interface{}, err error) {
	executorType := config.Name
	configInstance, exists := f.executorTypeAndConfigInstanceMap[executorType]
	if !exists {
		err = fmt.Errorf("could'nt found executor type: %v", executorType)
		return
	}
	if configInstance != nil && config.FactoryConfigValue != "" {
		err := config.ConvertInterfaceTypeToConfigInstance(configInstance)
		if err != nil {
			logger.Log.Errorf("Failed to unmarshal json: %v to instance type: %v, error: %v", config.FactoryConfigValue, configInstance, err.Error())
		} else {
			logger.Log.Infof("Convert interface type to configInstance: %v", configInstance)
		}
	}
	executorFunc := f.executorMap[executorType]
	executor, err = executorFunc(configInstance)
	f.executorTypeAndExecutorMap[executorType] = executor
	logger.Log.Infof("Init executorType: %v executor: %v error: %v", executorType, executor, err)
	return
}

func (m *Config) ConvertInterfaceTypeToConfigInstance(configInstance interface{}) (e error) {
	if bytes, e := json.Marshal(m.FactoryConfigValue); e != nil {
		return e
	} else {
		return json.Unmarshal(bytes, configInstance)
	}
}
