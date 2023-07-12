package monitor

import (
	"net/http"
	_ "net/http/pprof"

	log "github.com/sirupsen/logrus"
)

// todo 这里只是 dev 阶段开启 pprof，故没有做 goroutine 管理
func StartPprof() {
	// 开启pprof，监听请求
	ip := "0.0.0.0:6060"
	if err := http.ListenAndServe(ip, nil); err != nil {
		log.Fatal("start pprof failed on %s\n", ip)
	}
}
