import (
    "strconv"
)

output: {
	"repCharacterDatas": [
			for i, param in inject.params {
				{
					"entityId": "\(inject.serverName)MyVirtualSpaceCharacterBase_C_\(i+1)",
					"position":
					{
						"x": param.x,
						"y": param.y,
						"z": 2702
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
					// "netGuid" 必须为偶数
					if mod(param.guid, 2) == 0 {
						"netGuid": param.guid
					},
					if mod(param.guid, 2) == 1 {
//						"netGuid": inject.serverNameInt + param.guid + 1
//						"netGuid": strconv.Atoi(intject.serverName) + param.guid + 1 // 无法转换
						"netGuid": param.guid + 1
					},
				}
			}
	]
}

inject: {
    serverName: $serverName,
    params: [$userCount]arr[...{x|float|-6100~-6000|100,y|float|8300~8400|100,yaw|float|-130~130|2,guid|int|8377614~8388888|0}]
//    serverNameInt: strconv.Atoi(serverName)
}
