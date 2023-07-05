package app

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"mock-ue-server/pkg/controller"
	"mock-ue-server/pkg/logutil"
)

const (
	FlagDB           = "db"
	FlagServerNames  = "server-names"
	FlagUserCount    = "user-count"
	FlagIsPrintDebug = "is-print-debug-log"
)

var logger = logutil.GetLogger()

func NewMockUeServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mock-ue-server",
		Short: "mock ue server help you to mock ue server",
		Long:  `mock ue server help you to mock ue server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 解析并校验参数
			db := cmd.Flag(FlagDB).Value.String()
			serverNamesStr := cmd.Flag(FlagServerNames).Value.String()
			serverNames, err := parseServerNames(serverNamesStr)
			if err != nil {
				return err
			}
			userCount, err := cmd.Flags().GetInt(FlagUserCount)
			if err != nil {
				return errors.Wrap(err, "get user count failed")
			}
			isPrintDebug, err := cmd.Flags().GetBool(FlagIsPrintDebug)
			if err != nil {
				return errors.Wrap(err, "get is print debug log failed")
			}
			logger.Info(isPrintDebug)
			return run(db, serverNames, userCount)
		},
	}

	// 添加 flag
	// db 为 redis 的 host:port，默认为 127.0.0.1:6379
	cmd.Flags().StringP(FlagDB, "d", "127.0.0.1:6379", "db addr, host:port")
	// server-names 为模拟的服务器名称，以逗号分隔，例如 server1,server2,server3
	cmd.Flags().StringP(FlagServerNames, "s", "", "server names, split by comma")
	// user-count 为每个 server 连接的用户数，默认为 100
	cmd.Flags().IntP(FlagUserCount, "u", 100, "user count for each server")
	// isPrintDebug 为 true 时，打印 debug 日志，否则打印 info 日志
	cmd.Flags().BoolP(FlagIsPrintDebug, "v", false, "is print debug log")
	return cmd
}

func run(db string, serverNames []string, userCount int) error {
	// todo isPrintDebug 为 true 时，打印 debug 日志，否则打印 info 日志
	logutil.Init(true)

	// 启动 mock server
	ctx, cancel := context.WithCancel(context.Background())
	err := startMockServer(ctx, db, serverNames, userCount, cancel)
	if err != nil {
		logger.Errorf("Start mock server failed, err: %+v", err)
		return err
	}

	// 监听退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigCh:
		logger.Info("Receive exit signal")
		cancel()
	case <-ctx.Done():
		logger.Info("Mock server exit")
	}
	time.Sleep(1 * time.Second)
	return nil
}

func startMockServer(ctx context.Context, db string, serverNames []string, userCount int,
	cancel context.CancelFunc) error {
	logger.Infof("Start server. db: %s, serverNames: %v, userCount: %d",
		db, serverNames, userCount)

	mockController, err := controller.NewMockController(db, serverNames, userCount)
	if err != nil {
		return err
	}
	go mockController.Run(ctx, cancel)
	return nil
}

func parseServerNames(serverNamesStr string) ([]string, error) {
	// 以逗号分隔 serverNamesStr，清除空格并检查是否为空字符串
	serverNames := make([]string, 0)
	for _, name := range strings.Split(serverNamesStr, ",") {
		trimmedName := strings.TrimSpace(name)
		if trimmedName == "" {
			return nil, errors.New("serverName can not be empty")
		}
		serverNames = append(serverNames, trimmedName)
	}

	// 检查是否解析出 serverNames
	if len(serverNames) == 0 {
		return nil, errors.Errorf("Parse serverNames failed, serverNamesStr: %s", serverNamesStr)
	}

	return serverNames, nil
}
