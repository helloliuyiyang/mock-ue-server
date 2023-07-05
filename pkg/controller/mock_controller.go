package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"

	"mock-ue-server/pkg/database"
	"mock-ue-server/pkg/logutil"
	"mock-ue-server/pkg/model"
)

var logger = logutil.GetLogger()

type MockController struct {
	redisCli    *redis.Client
	serverNames []string
	userCount   int
}

func NewMockController(db string, serverNames []string, userCount int) (*MockController, error) {
	err := database.InitRedisCli(db)
	if err != nil {
		return nil, err
	}
	return &MockController{
		redisCli:    database.GetRedisCli(),
		serverNames: serverNames,
		userCount:   userCount,
	}, nil
}

func (c *MockController) Run(ctx context.Context, cancel context.CancelFunc) {
	// 循环 1s 输出一次日志
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 遍历所有 serverName，分别生成对应的 redis 数据
			for _, serverName := range c.serverNames {
				redisData := model.RedisData{CharacterDatas: make([]*model.CharacterData, 0, c.userCount)}
				for i := 0; i < c.userCount; i++ {
					data := model.NewDefaultCharacterData(fmt.Sprintf("%s-%d", serverName, i))
					redisData.CharacterDatas = append(redisData.CharacterDatas, data)
				}
				bytes, err := json.MarshalIndent(redisData, "", "    ")
				if err != nil {
					logger.Errorf("Json marshal failed: %+v", err)
					continue
				}
				key := fmt.Sprintf("%s---PawnDatas", serverName)
				_, err = c.redisCli.Set(ctx, key, string(bytes), 0).Result()
				if err != nil {
					logger.Errorf("Redis set failed: %+v", err)
					continue
				}
				logger.Infof("Redis set success. key: %s, value: %s", key, string(bytes))
			}

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

		case <-ctx.Done():
			logger.Info("MockController exit")
			return
		}
	}
}

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
