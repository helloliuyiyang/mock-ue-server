output: {
	"characterDatas": [
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
						"yaw": -30.105167388916016,
						"roll": 0
					},
					"walkSpeed": 150,
					"dynamicproperties_int": [ 16, 25, 0, 0, 0 ],
					"dynamicproperties_string_redis": [
						"258",
						"x00517450-ue-client-\(inject.serverName)",
						"",
						"",
						""
					]
				}
			}
	]
}


inject: {
    serverName: $serverName,
    params: [$userCount]arr[...{x|float|-8000~7000|10,y|float|12000~17000|10}]
}