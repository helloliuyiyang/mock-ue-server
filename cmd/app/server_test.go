package app

import (
	"reflect"
	"testing"
)

func Test_parseServerNames(t *testing.T) {
	type args struct {
		serverNamesStr string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{"normal1", args{"server1, server2, server3"},
			[]string{"server1", "server2", "server3"}, false},
		{"normal2", args{" server1 , server2 , server3 "},
			[]string{"server1", "server2", "server3"}, false},
		{"normal3", args{"01,02,03"}, []string{"01", "02", "03"}, false},
		{"empty1", args{"server1, , server3"}, nil, true},
		{"empty2", args{""}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseServerNames(tt.args.serverNamesStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseServerNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseServerNames() got = %v, want %v", got, tt.want)
			}
		})
	}
}
