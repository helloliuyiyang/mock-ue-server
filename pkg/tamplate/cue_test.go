package tamplate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

func Test_cue(t *testing.T) {
	// 读取文件 test.cue
	cueBytes, err := os.ReadFile("./test.cue")
	if err != nil {
		t.Errorf("read file failed: %v", err)
		return
	}

	// 替换 cueBytes 中的 $serverName 为 666
	cueBytes = bytes.ReplaceAll(cueBytes, []byte("$serverName"), []byte("666"))
	regex := regexp.MustCompile(`\[\$userCount\]arr\[\.\.\.\{(.*?)\}\]`)
	match := regex.FindStringSubmatch(string(cueBytes))
	if len(match) < 1 {
		log.Fatalf("error parsing arr: %s", cueBytes)
	}

	fieldsStr := match[1]
	// 生成的对象数组中包含的对象个数
	userCount := 2
	arr, err := genObjArr(fieldsStr, userCount)
	if err != nil {
		fmt.Println("Error generating example:", err)
		return
	}
	arrBytes, err := json.MarshalIndent(arr, "", "    ")
	//bytes, err := json.Marshal(arr)
	//if err != nil {
	//	log.Fatalf("error marshaling data: %v", err)
	//}
	//log.Printf("===arr: \n%s\n", arrBytes)

	// 替换 params 的值
	cueBytes = bytes.ReplaceAll(cueBytes, []byte(match[0]), arrBytes)

	cc := cuecontext.New()
	value := cc.CompileBytes(cueBytes)
	// todo fillPath 为啥不成功咧？
	//value.FillPath(cue.ParsePath("inject"), cc.CompileString(inject))
	//value.FillPath(cue.ParsePath("param"), cc.CompileString(param))
	//fields, err := value.LookupPath(cue.ParsePath("inject")).Fields()
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//for fields.Next() {
	//	fmt.Println(fields.Label(), fields.Value(), fields.Value().Attribute())
	//}
	b, err := value.LookupPath(cue.ParsePath("output")).MarshalJSON()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(b))
}

type Field struct {
	fieldName string
	valueType string
	value     interface{}
}

func generateRandomValueFloat(valueRangeStr string) (float64, error) {
	valueRange := strings.Split(valueRangeStr, "~")
	if len(valueRange) != 2 {
		return 0, fmt.Errorf("invalid value range format: %s", valueRangeStr)
	}
	min, err := strconv.ParseFloat(valueRange[0], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid min value in range: %s", valueRange[0])
	}
	max, err := strconv.ParseFloat(valueRange[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid max value in range: %s", valueRange[1])
	}
	return min + rand.Float64()*(max-min), nil
}

func genObjArr(fieldsStr string, userCount int) ([]interface{}, error) {
	arr := make([]interface{}, 0, userCount)
	for i := 0; i < userCount; i++ {
		// 解析用户定义的 arr, 以逗号分割, eg: allFields = ["x.int.0~100", "y.float.0~100"]
		allFields := strings.Split(fieldsStr, ",")

		// 遍历 allFields
		fields := make([]Field, 0, len(allFields))
		for _, field := range allFields {
			// 解析用户定义的 arr
			split := strings.Split(field, ".")
			if len(split) != 3 {
				return nil, fmt.Errorf("invalid arr format: %s", fieldsStr)
			}

			f := Field{}
			f.fieldName = split[0]
			f.valueType = split[1]
			value, err := generateRandomValueFloat(split[2])
			if err != nil {
				return nil, err
			}
			if f.valueType == "int" {
				f.value = int(value)
			} else if f.valueType == "float" {
				f.value = value
			} else {
				return nil, fmt.Errorf("unknown value type: %s", f.valueType)
			}
			fields = append(fields, f)
		}

		em := make(map[string]interface{})
		for _, field := range fields {
			em[field.fieldName] = field.value
		}
		arr = append(arr, em)
	}
	//default:
	//	return nil, fmt.Errorf("Unknown arr type: %s", fieldName)
	return arr, nil
}

// 下面是不要的代码
//logger.Infof("MockController run. serverNames: %v, userCount: %d", c.serverNames, c.userCount)
//redisData101, err := c.Get101Data()
//if err != nil {
//	logger.Errorf("Get101Data failed: %+v", err)
//	continue
//}
//
//copyCharacter := redisData101.CharacterDatas[0]
//copyCharacter.EntityId = "copy-entity"
//copyCharacter.Position.X += 50
//copyCharacter.Position.Y += 50
//if len(redisData101.CharacterDatas) == 1 {
//	redisData101.CharacterDatas = append(redisData101.CharacterDatas, copyCharacter)
//} else {
//	redisData101.CharacterDatas[1] = copyCharacter
//}
//

//newRedisData := model.RedisData{
//	CharacterDatas: []model.CharacterData{copyCharacter},
//}
//newBytes, err := json.Marshal(newRedisData)
//if err != nil {
//	logger.Errorf("json marshal failed: %+v", err)
//	continue
//}

/*
// Get101Data only for test
func (c *MockController) Get101Data() (*model.RedisData, error) {
	const key101 = "101---PawnDatas"
	var ctx = context.Background()
	result, err := c.redisCli.Get(ctx, key101).Result()
	if err != nil {
		return nil, errors.Wrapf(err, "redis get failed, key: %s", key101)
	}

	ret := model.RedisData{}
	err = json.Unmarshal([]byte(result), &ret)
	if err != nil {
		return nil, errors.Wrapf(err, "json unmarshal failed, result: %s", result)
	}
	//logger.Infof("Get101Data. result: %+v", ret)
	return &ret, nil
}
*/
