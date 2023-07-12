package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"

	"mock-ue-server/pkg/database"
	"mock-ue-server/pkg/logutil"
	"mock-ue-server/pkg/tamplate"
)

const redisKeyTpl = "%s---PawnDatas"

var logger = logutil.GetLogger()

type MockController struct {
	redisCli    *redis.Client
	serverNames []string
	userCount   int
	intervalMs  int
	tplInfo     *tamplate.TemplateInfo
	// injectParamsCache 注入模板 params 的数据缓存，包含了 Schema 和值
	injectParamsCache *InjectParamsCache
}

func NewMockController(serverNames []string, userCount int,
	tplFilePath string, intervalMs int) (*MockController, error) {
	var (
		tplInfo *tamplate.TemplateInfo
		err     error
	)

	// 解析模板信息
	if tplFilePath != "" {
		tplInfo, err = tamplate.ParseTplInfo(tplFilePath)
		if err != nil {
			return nil, err
		}
	}

	// 初始化缓存，Cache 中包含需要注入到模板的数据的 Schema 和 值
	if err = initParamsCache(serverNames, userCount, tplInfo.FieldsStr); err != nil {
		return nil, err
	}

	return &MockController{
		redisCli:          database.GetRedisCli(),
		serverNames:       serverNames,
		userCount:         userCount,
		intervalMs:        intervalMs,
		tplInfo:           tplInfo,
		injectParamsCache: getParamsCache(),
	}, nil
}

// todo cancel 用于定时任务发生致命错误，需要退出整个进程，这里暂时不用，预留
func (c *MockController) Run(ctx context.Context, cancel context.CancelFunc) {
	ticker := time.NewTicker(time.Duration(c.intervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.validateCache(); err != nil {
				logger.Errorf("validateCache failed: %+v", err)
				cancel()
				return
			}
			// 遍历所有 serverName，分别生成对应的 redis 数据
			for _, serverName := range c.serverNames {
				c.mockServer(serverName)
			}
			// 根据 schema 刷新 serverParamsMap 数据，用于下一次注入模板
			c.refreshAllCache()
			logger.Infof("=== Run once ok")
		case <-ctx.Done():
			logger.Info("MockController exit")
			return
		}
	}
}

func (c *MockController) mockServer(serverName string) {
	// 生成注入到模板中 params 的数据
	paramsBytes, err := c.injectParamsCache.genServerParamsBytes(serverName)
	if err != nil {
		logger.Fatalf("genParamsBytes failed: %+v", err)
	}

	bytes, err := tamplate.GenRedisData(c.tplInfo, serverName, paramsBytes)
	if err != nil {
		logger.Fatalf("GenRedisData failed: %+v", err)
	}
	key := fmt.Sprintf(redisKeyTpl, serverName)
	_, err = c.redisCli.Set(context.Background(), key, string(bytes), 0).Result()
	if err != nil {
		logger.Fatalf("Redis set failed: %+v", err)
	}
	//logger.Infof("Redis set success. key: %s", key)
	logger.Debugf("=== key: %s, value: %s", key, string(bytes))
}

func (c *MockController) refreshAllCache() {
	for _, server := range c.serverNames {
		c.injectParamsCache.RefreshServerParams(server)
	}
}

// validateCache 仅用在项目初期，确保生成的数据不会出错
func (c *MockController) validateCache() error {
	serverCount := len(c.serverNames)
	if len(c.injectParamsCache.serverParamsMap) != serverCount {
		return errors.Errorf("serverParamsMap len[%d] not match server count[%d]",
			len(c.injectParamsCache.serverParamsMap), serverCount)
	}

	fieldsCount := len(c.injectParamsCache.fieldSchemaMap)
	for _, ps := range c.injectParamsCache.serverParamsMap {
		if len(ps) != c.userCount {
			return errors.Errorf("serverParamsMap len[%d] not match user count[%d]",
				len(ps), c.userCount)
		}
		for _, p := range ps {
			if len(p) != fieldsCount {
				return errors.Errorf("serverParamsMap len[%d] not match fields count[%d]",
					len(p), fieldsCount)
			}
		}
	}

	return nil
}
