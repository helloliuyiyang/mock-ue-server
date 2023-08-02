:: build mock-ue-server
go build ../cmd/mock-ue-server.go

:: ===========================================================================================================
:: run command example

:: 950 人，19 server * 50 人
:: ./mock-ue-server -s "2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20" --user-count 50 -t ./redis_data.cue

:: 指定 redis server
:: ./mock-ue-server -d 10.12.3.74:6379 -s "2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20" --user-count 50 -t ./redis_data.cue