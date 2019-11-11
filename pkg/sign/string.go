package sign

import (
	"fmt"
	"github.com/blademainer/commons/pkg/logger"
	"github.com/fatih/structs"
	"sort"
	"strings"
)

type ParamsCompacter struct {
	IgnoreKeys                  []string
	IgnoreEmptyValue            bool
	entityFieldNames            []string
	SortedKeyFieldNames         []string
	PairsDelimiter              string
	KeyValueDelimiter           string
	FieldTag                    string
	fieldTagNameAndFieldNameMap map[string]string
}

func NewParamsCompacter(entityDemoInstance interface{}, fieldTag string, ignoreKeys []string, ignoreEmptyValue bool, pairsDelimiter string, keyValueDelimiter string) ParamsCompacter {
	if fieldTag == "" {
		fieldTag = "json"
	}

	fieldNames := make([]string, 0)
	fieldTagNameAndFieldNameMap := make(map[string]string)
	instance := structs.New(entityDemoInstance)
	entityFieldNames := instance.Names()
l:
	for _, k := range entityFieldNames {
		field := instance.Field(k)
		tag := field.Tag(fieldTag)
		name, tagOptions := parseTag(tag)
		logger.Infof("field: %v with options: %v", field, tagOptions)
		if tagOptions.Contains("-") || name == "-" {
			logger.Infof("ignore field: %v by tag: %v", k, name)
			continue
		}
		fieldTagNameAndFieldNameMap[name] = field.Name()
		for _, ignore := range ignoreKeys {
			if ignore == name {
				continue l
			}
		}
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)

	p := ParamsCompacter{}
	p.entityFieldNames = entityFieldNames
	p.SortedKeyFieldNames = fieldNames
	p.IgnoreEmptyValue = ignoreEmptyValue
	p.PairsDelimiter = pairsDelimiter
	p.KeyValueDelimiter = keyValueDelimiter
	p.fieldTagNameAndFieldNameMap = fieldTagNameAndFieldNameMap

	return p
}

func (p ParamsCompacter) ParamsToString(instance interface{}) string {
	defer func() {
		if message := recover(); message != nil {
			fmt.Println(message)
		}
	}()
	params := make(map[string]string)
	s := structs.New(instance)
	for _, tagFieldName := range p.SortedKeyFieldNames {
		fieldName := p.fieldTagNameAndFieldNameMap[tagFieldName]
		field := s.Field(fieldName)
		value := field.Value()
		stringValue := fmt.Sprintf("%v", value)
		params[tagFieldName] = stringValue
	}

	return p.BuildMapToString(params)
}

func (p ParamsCompacter) BuildMapToString(params map[string]string) string {
	builder := strings.Builder{}
	delimiter := ""
	for _, key := range p.SortedKeyFieldNames {
		value := params[key]
		if p.IgnoreEmptyValue && value == "" {
			continue
		}
		builder.WriteString(delimiter)
		builder.WriteString(key)
		builder.WriteString(p.KeyValueDelimiter)
		builder.WriteString(value)
		delimiter = p.PairsDelimiter
	}
	return builder.String()
}
