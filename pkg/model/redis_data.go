package model

import "math/rand"

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
