:: build mock-ue-server
go build ../cmd/mock-ue-server.go

:: run command example
:: ./mock-ue-server -s 801,802,803,804,805,806,807,808,809,810,811,812,813,814,815,816,817,818,819,820,821,822,823,824,825 --user-count 40 -t ./redis_data.cue
:: ./mock-ue-server -s 801,802,803,804,805,806,807,808,809,810,811,812,813,814,815,816,817,818,819,820 --user-count 50 -t ./redis_data.cue
:: ./mock-ue-server -d 10.12.3.74:6379 -s 02,03,04,05 --user-count 50 -t ./redis_data.cue