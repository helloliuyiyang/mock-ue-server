/**
该 package 仅用于未指定数据模板的情况，不推荐使用

在这里存一些不用的代码
*/

package model

import "math/rand"

// deprecated
// RedisData 默认数据格式，仅用于未指定数据模板的情况
type RedisData struct {
	CharacterDatas []*CharacterData `json:"characterDatas"`
}

type CharacterData struct {
	EntityId          string            `json:"entityId"`
	Position          Position          `json:"position"`
	Rotation          Rotation          `json:"rotation"`
	DynamicProperties map[string]string `json:"dynamicProperties"`
	WalkSpeed         int               `json:"walkSpeed"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type Rotation struct {
	Pitch int     `json:"pitch"`
	Yaw   float64 `json:"yaw"`
	Roll  int     `json:"roll"`
}

// NewDefaultCharacterData 生成默认的 CharacterData 测试数据
func NewDefaultCharacterData(entityId string) *CharacterData {
	return &CharacterData{
		EntityId: entityId,
		Position: Position{
			X: float64(rand.Intn(10000) - 10000), // -10000 ~ 0
			Y: float64(rand.Intn(5000) + 10000),  // 10000 ~ 15000
			Z: 2330.822509765625,
		},
		Rotation: Rotation{
			Pitch: 0,
			Yaw:   5.9675374031066895,
			Roll:  0,
		},
		DynamicProperties: map[string]string{
			"16": "256",
			"25": "l00648512-ue-develop",
		},
		WalkSpeed: 150,
	}
}

/*
	if c.tplInfo == nil { // 没有模板文件，生成默认数据，废弃，后面临时修改工具来不及，使用这种方式
		redisData := model.RedisData{CharacterDatas: make([]*model.CharacterData, 0, c.userCount)}
		for i := 0; i < c.userCount; i++ {
			data := model.NewDefaultCharacterData(fmt.Sprintf("%s-%d", serverName, i))
			redisData.CharacterDatas = append(redisData.CharacterDatas, data)
		}
		bytes, err := json.Marshal(redisData)
		if err != nil {
			logger.Fatalf("Json marshal failed: %+v", err)
		}
		key := fmt.Sprintf(redisKeyTpl, serverName)
		_, err = c.redisCli.Set(ctx, key, string(bytes), 0).Result()
		if err != nil {
			logger.Errorf("Redis set failed: %+v", err)
			continue
		}
		logger.Infof("Redis set success. key: %s", key)
		logger.Debugf("=== Redis set success. key: %s, value: %s", key, string(bytes))
	} else { // 有模板文件，生成模板数据
*/
