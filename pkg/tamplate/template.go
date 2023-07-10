package tamplate

import (
	"os"
	"regexp"

	"github.com/pkg/errors"
)

const (
	serverNameReplaceFlag = "$serverName"
)

type TemplateInfo struct {
	// 模板文件原始内容
	TplContent []byte
	// 模板文件中 inject.params 对象数组中字段信息，如：x|float|-10~10,y|float|-10000~9000
	FieldsStr string
	// 生成对象数组后需要替换掉的内容，如：[{x: 1.1}] 替换 [$userCount]arr[...{x|float|-10~10}]
	ParamsValueTobeReplaced []byte
	// 服务名需要替换的内容，如：101 替换 $serverName
	ServerNameTobeReplaced []byte
}

func newTemplateInfo(tplContent []byte, fieldsStr string, paramsValueTobeReplaced []byte) *TemplateInfo {
	return &TemplateInfo{
		TplContent:              tplContent,
		FieldsStr:               fieldsStr,
		ParamsValueTobeReplaced: paramsValueTobeReplaced,
		ServerNameTobeReplaced:  []byte(serverNameReplaceFlag),
	}
}

func ParseTplInfo(tplFilePath string) (*TemplateInfo, error) {
	// 读取模板文件
	bytes, err := os.ReadFile(tplFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "read file failed")
	}

	// 解析模板文件
	regex := regexp.MustCompile(`\[\$userCount\]arr\[\.\.\.\{(.*?)\}\]`)
	match := regex.FindStringSubmatch(string(bytes))
	if len(match) < 2 {
		return nil, errors.Errorf("error parsing arr: %s", bytes)
	}

	// 生成模板信息
	// match[0] 为需要替换的内容，如：[{x: 1.1}] 替换 [$userCount]arr[...{x|float|-10~10}]
	// match[1] 为对象数组中对象的字段信息，如：x|float|-10~10,y|float|-10000~9000
	info := newTemplateInfo(bytes, match[1], []byte(match[0]))
	return info, nil
}
