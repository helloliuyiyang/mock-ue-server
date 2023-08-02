output: {
	"repCharacterDatas": [
			for i, param in inject.params {
				{
					"entityId": "\(inject.serverName)MyVirtualSpaceCharacterBase_C_\(i+1)",
					"position":
					{
						"x": param.x,
						"y": param.y,
						"z": 2330.822509765625
					},
					"rotation":
					{
						"pitch": 0,
						"yaw": param.yaw,
						"roll": 0
					},
					"characterProperties": {},
					"playerStateProperties": {
						"16": "AAEAAA==",
						"25": "FQAAAGwwMDY0ODUxMi11ZS1kZXZlbG9wAA=="
					},
					"walkSpeed": 150,
					"netGuid": param.guid,
				}
			}
	]
}

inject: {
    serverName: $serverName,
    params: [$userCount]arr[...{x|float|-7000~-5000|100,y|float|13000~16000|100,yaw|float|-130~130|2,guid|int|8377614~8388888|0}]
}
