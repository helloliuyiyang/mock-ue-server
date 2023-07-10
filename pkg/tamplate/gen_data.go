package tamplate

import (
	"bytes"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/pkg/errors"

	"mock-ue-server/pkg/logutil"
)

var logger = logutil.GetLogger()

func GenRedisData(tpl *TemplateInfo, serverName string, paramsBytes []byte) ([]byte, error) {
	// 1. 生成正式的 cue 模板
	cueBytes := tpl.TplContent
	// 替换 params 的值
	cueBytes = bytes.ReplaceAll(cueBytes, tpl.ParamsValueTobeReplaced, paramsBytes)
	// 替换 serverName 的值
	cueBytes = bytes.ReplaceAll(cueBytes, tpl.ServerNameTobeReplaced, []byte(serverName))
	logger.Debugf("=== cueBytes: \n%s\n", cueBytes)

	// 2. cue 渲染 json
	cc := cuecontext.New()
	value := cc.CompileBytes(cueBytes)
	b, err := value.LookupPath(cue.ParsePath("output")).MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "marshal json failed")
	}
	return b, nil
}
