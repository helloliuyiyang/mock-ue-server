package controller

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	fieldTypeInt   = "int"
	fieldTypeFloat = "float"

	fieldsStrSep = ","
	fieldSep     = "|"
	rangeSep     = "~"

	fieldNameIndex0      = 0
	fieldTypeIndex1      = 1
	fieldRangeIndex2     = 2
	fieldMaxChangeIndex3 = 3

	// defaultMaxChange 默认最大变化量，为 -1 时表示不限制
	defaultMaxChange = -1.0
)

var cache *InjectParamsCache

// InjectParamsCache 注入模板 params 的信息，包括 field schema信息 与 生成的值；非并发安全
type InjectParamsCache struct {
	// fieldSchemaMap 记录每个 field 的 schema 信息，key 为 fieldName
	// eg: {x: {fieldName: x, fieldType: int, minValue: 0, maxValue: 100},
	//		y: {fieldName: y, fieldType: float, minValue: 0, maxValue: 100}}
	fieldSchemaMap map[string]fieldSchema
	// serverParamsMap 记录各 server 的 params, key 为 serverName
	// eg: {server1: [{x: 1, y: 2}, {x: 3, y: 4}], server2: [{x: 1, y: 2}, {x: 3, y: 4}]}
	serverParamsMap map[string]params
}

type fieldSchema struct {
	fieldName string
	fieldType string
	minValue  float64
	maxValue  float64
	// 每个数据刷新周期的最大变化量，用于控制每次刷新时变化的幅度
	maxChange float64
}

// eg: [{x: 1, y: 2}, {x: 3, y: 4}]
type params []param

// eg: {x: 1, y: 2}
type param map[string]interface{}

func initParamsCache(serverNames []string, userCount int, fieldStr string) error {
	// 1. 初始化 fieldSchemaMap
	schemaMap, err := genFieldSchemaMap(fieldStr)
	if err != nil {
		return err
	}

	// 2. 初始化 serverParamsMap
	serverParamsMap, err := genServerParamsMap(serverNames, userCount, schemaMap)
	if err != nil {
		return err
	}

	// 3. 将 1、2 中的信息记录到 cache 中
	cache = &InjectParamsCache{
		fieldSchemaMap:  schemaMap,
		serverParamsMap: serverParamsMap,
	}
	return nil
}

func getParamsCache() *InjectParamsCache {
	return cache
}

func genFieldSchemaMap(fieldsStr string) (map[string]fieldSchema, error) {
	// 解析用户定义的 arr, 以逗号分割, eg: allFields = ["x|int|0~100", "y|float|0~100"]
	allFields := strings.Split(fieldsStr, fieldsStrSep)
	if len(allFields) == 0 {
		return nil, errors.Errorf("invalid inject.params fieldsStr: %s", fieldsStr)
	}

	// 遍历 allFields, 解析每个 field 的 schema 信息
	schemaMap := make(map[string]fieldSchema, len(allFields))
	for _, field := range allFields {
		splits := strings.Split(field, fieldSep)
		if len(splits) != 3 && len(splits) != 4 {
			return nil, errors.Errorf("invalid inject.params format: %s\n, field splits: %+v ",
				fieldsStr, splits)
		}
		valueRange := strings.Split(splits[fieldRangeIndex2], rangeSep)
		if len(valueRange) != 2 {
			return nil, errors.Errorf("invalid value range format: %s", splits[fieldRangeIndex2])
		}
		min, err := strconv.ParseFloat(valueRange[0], 64)
		if err != nil {
			return nil, errors.Errorf("invalid min value in range: %s", valueRange[0])
		}
		max, err := strconv.ParseFloat(valueRange[1], 64)
		if err != nil {
			return nil, errors.Errorf("invalid max value in range: %s", valueRange[1])
		}
		if min == max {
			return nil, errors.Errorf("invalid range: %s", splits[fieldRangeIndex2])
		}
		maxChange := defaultMaxChange
		if len(splits) > 3 {
			maxChange, err = strconv.ParseFloat(splits[fieldMaxChangeIndex3], 64)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid max change: %s", splits[fieldMaxChangeIndex3])
			}
			if maxChange <= 0 {
				return nil, errors.Errorf("invalid max change: %s", splits[fieldMaxChangeIndex3])
			}
		}

		fieldName := splits[fieldNameIndex0]
		schemaMap[fieldName] = fieldSchema{
			fieldName: fieldName,
			fieldType: splits[fieldTypeIndex1],
			minValue:  min,
			maxValue:  max,
			maxChange: maxChange,
		}
	}
	return schemaMap, nil
}

func genServerParamsMap(serverNames []string, userCount int, schemaMap map[string]fieldSchema) (map[string]params, error) {
	serverParamsMap := make(map[string]params, len(serverNames))
	// 字段个数
	fieldCount := len(schemaMap)
	// 生成模板中需要注入的 params 的值，记录到 paramsMap 中
	for _, serverName := range serverNames {
		// params eg: [{x: 1, y: 2}, {x: 3, y: 4}]
		ps := make(params, 0, userCount)

		// 循环 userCount 次，生成每个 server 的 param
		for i := 0; i < userCount; i++ {
			// param eg: {x: 1, y: 2}
			p := make(param, fieldCount)
			// 遍历 schemaMap，生成 param 中每个 k，v
			for _, schema := range schemaMap {
				// 生成每个 field 的值
				value := generateRandomValueFloat(schema.minValue, schema.maxValue)
				switch schema.fieldType {
				case fieldTypeFloat:
					p[schema.fieldName] = value
				case fieldTypeInt:
					p[schema.fieldName] = int(value)
				default:
					return nil, errors.Errorf("invalid field[%s] type: %s", schema.fieldName, schema.fieldType)
				}
			}
			ps = append(ps, p)
		}
		serverParamsMap[serverName] = ps
	}

	return serverParamsMap, nil
}

func generateRandomValueFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// genParamsStr 生成对应 server 的 params 的对象数组，返回序列化后的字节切片，用于注入到 cue 模板中
func (c *InjectParamsCache) genServerParamsBytes(serverName string) ([]byte, error) {
	serverParams, ok := c.serverParamsMap[serverName]
	if !ok {
		return nil, errors.Errorf("serverName: %s not found in injectParamsCache", serverName)
	}
	arrBytes, err := json.Marshal(serverParams)
	if err != nil {
		return nil, errors.Wrap(err, "json.Marshal err")
	}
	logger.Debugf("=== arrBytes: %s", string(arrBytes))
	return arrBytes, nil
}

// RefreshServerParams 根据 schema 刷新 serverParamsMap 数据，用于下一次注入
func (c *InjectParamsCache) RefreshServerParams(serverName string) {
	// 根据 schema 刷新 serverParamsMap 数据
	ps := c.serverParamsMap[serverName]
	for _, schema := range c.fieldSchemaMap {
		for i := 0; i < len(ps); i++ {
			old := ps[i][schema.fieldName]
			switch schema.fieldType {
			case fieldTypeFloat:
				oldFloat, ok := old.(float64)
				if !ok {
					logger.Fatalf("Logic error, old value: %+v is not float64", old)
				}
				ps[i][schema.fieldName] = genNewValueFloat(oldFloat, schema)
			case fieldTypeInt:
				oldInt, ok := old.(int)
				if !ok {
					logger.Fatalf("Logic error, old value: %+v is not float64", old)
				}
				ps[i][schema.fieldName] = int(genNewValueFloat(float64(oldInt), schema))
			}
		}
	}
}

// genNewValueFloat 生成新的 float64 类型的值
func genNewValueFloat(oldFloat float64, field fieldSchema) float64 {
	// 如果 maxChange 为默认值，则生成随机值
	if field.maxChange == defaultMaxChange {
		return generateRandomValueFloat(field.minValue, field.maxValue)
	}

	// 如果 maxChange 不为默认值，则生成在变化范围在 maxChange 内的随机值
	max := oldFloat + field.maxChange
	if max > field.maxValue {
		max = field.maxValue
	}
	min := oldFloat - field.maxChange
	if min < field.minValue {
		min = field.minValue
	}
	return generateRandomValueFloat(min, max)
}
