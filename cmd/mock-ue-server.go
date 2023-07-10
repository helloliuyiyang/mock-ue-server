package main

import (
	"fmt"
	"os"

	"mock-ue-server/cmd/app"
)

// 用 cobra 方式启动工具
func main() {
	cmd := app.NewMockUeServerCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Printf("Execute failed: %+v\n", err)
		os.Exit(1)
	}
}
