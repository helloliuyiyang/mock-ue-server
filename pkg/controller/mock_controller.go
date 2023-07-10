package controller

import (
	"context"
	"fmt"
	"time"

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
			// 遍历所有 serverName，分别生成对应的 redis 数据
			for _, serverName := range c.serverNames {
				// 生成注入到模板中 params 的数据
				paramsBytes, err := c.injectParamsCache.genServerParamsBytes(serverName)
				if err != nil {
					logger.Errorf("genParamsBytes failed: %+v", err)
					continue
				}

				bytes, err := tamplate.GenRedisData(c.tplInfo, serverName, paramsBytes)
				if err != nil {
					logger.Errorf("GenRedisData failed: %+v", err)
					continue
				}
				key := fmt.Sprintf(redisKeyTpl, serverName)
				_, err = c.redisCli.Set(context.Background(), key, string(bytes), 0).Result()
				if err != nil {
					logger.Errorf("Redis set failed: %+v", err)
					continue
				}
				logger.Infof("Redis set success. key: %s", key)
				logger.Debugf("=== key: %s, value: %s", key, string(bytes))

				// 根据 schema 刷新 serverParamsMap 数据，用于下一次注入
				c.injectParamsCache.RefreshServerParams(serverName)
			}
		case <-ctx.Done():
			logger.Info("MockController exit")
			return
		}
	}
}
