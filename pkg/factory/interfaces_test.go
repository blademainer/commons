package factory

import (
	"encoding/json"
	"fmt"
	"testing"
)

type fakeFactoryMethod interface {
	ExecuteFunc() string
}

type fakeExecutor struct {
}

type fakeConfig struct {
	Pattern string `json:"pattern"`
}

var factory = InitFactory()

func (f *fakeExecutor) ExecuteFunc() string {
	return "fake executed"
}

func init() {
	factory.RegisterExecutor("fake", instanceFakeExecutor, &fakeConfig{})
}

func instanceFakeExecutor(configInstance interface{}) (executor interface{}, e error) {
	executor = &fakeExecutor{}
	return
}

func TestFactory_InstanceExecutor(t *testing.T) {
	c := &Config{}
	jsonString := `{"name": "fake", "config": {"pattern": "hello,pattern"}}`
	if e := json.Unmarshal([]byte(jsonString), c); e != nil {
		panic(e)
	}

	executor2, err := factory.InstanceExecutor(*c)
	if err != nil {
		panic(err)
	}
	fmt.Println(executor2)
	executor := factory.GetExecutor("fake")
	//
	executeFunc := executor.(fakeFactoryMethod)
	i := executeFunc.ExecuteFunc()
	fmt.Println(i)
}
