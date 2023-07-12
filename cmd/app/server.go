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

	"mock-ue-server/pkg/config"
	"mock-ue-server/pkg/controller"
	"mock-ue-server/pkg/database"
	"mock-ue-server/pkg/logutil"
	"mock-ue-server/pkg/monitor"
)

const (
	FlagDB           = "db"
	FlagServerNames  = "server-names"
	FlagUserCount    = "user-count"
	FlagIsPrintDebug = "is-print-debug-log"
	FlagTplFile      = "tpl-file"
	FlagInterval     = "interval"
)

var logger = logutil.GetLogger()

func NewMockUeServerCommand() *cobra.Command {
	conf := &config.Config{}
	cmd := &cobra.Command{
		Use:   "mock-ue-server",
		Short: "Short Description",
		//Long:  `mock ue server help you to mock ue server`,
		Long: `Long Description`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// 解析并校验参数
			serverNamesStr := cmd.Flag(FlagServerNames).Value.String()
			serverNames, err := parseServerNames(serverNamesStr)
			if err != nil {
				return err
			}
			userCount, err := cmd.Flags().GetInt(FlagUserCount)
			if err != nil {
				return errors.Wrap(err, "get user count failed")
			}
			intervalMs, err := cmd.Flags().GetInt(FlagInterval)
			if err != nil {
				return errors.Wrap(err, "get interval failed")
			}
			// 判断模板文件是否存在
			tplFile := cmd.Flag(FlagTplFile).Value.String()
			if tplFile == "" {
				return errors.New("param tpl file is empty")
			}
			if _, err := os.Stat(tplFile); os.IsNotExist(err) {
				return errors.Errorf("tpl file not exist: %s", tplFile)
			}

			// todo Db 理论上需要校验
			conf.Db = cmd.Flag(FlagDB).Value.String()
			conf.ServerNames = serverNames
			conf.UserCount = userCount
			conf.TplFilePath = cmd.Flag(FlagTplFile).Value.String()
			conf.IntervalMs = intervalMs
			conf.PrintDebugLog = cmd.Flag(FlagIsPrintDebug).Value.String() == "true"
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(conf)
		},
	}

	// 添加 flag
	// db 为 redis 的 host:port，默认为 127.0.0.1:6379
	cmd.Flags().StringP(FlagDB, "d", "127.0.0.1:6379", "db addr, host:port")
	// server-names 为模拟的服务器名称，该选项为必填项，以逗号分隔，例如 server1,server2,server3
	cmd.Flags().StringP(FlagServerNames, "s", "", "server names, split by comma, required")
	// user-count 为每个 server 连接的用户数，默认为 100
	cmd.Flags().IntP(FlagUserCount, "u", 100, "user count for each server")
	// isPrintDebug 用户输入该参数，打印 debug 日志，否则打印 info 日志
	cmd.Flags().BoolP(FlagIsPrintDebug, "v", false, "is print debug log")
	// todo dev 为开发模式，开发模式下，不会将数据插入数据库，只会打印生成的数据
	// interval 为每次刷新数据的间隔时间，单位 ms，默认为 100ms
	cmd.Flags().IntP(FlagInterval, "i", 100, "interval for refresh data, unit ms")
	// tpl-file 为模板文件路径，该选项为必填项
	cmd.Flags().StringP(FlagTplFile, "t", "", "redis data template file path, required")

	return cmd
}

func run(conf *config.Config) error {
	logutil.Init(conf.PrintDebugLog)

	go monitor.StartPprof()

	// 启动 mock server
	ctx, cancel := context.WithCancel(context.Background())
	err := startMockServer(ctx, conf, cancel)
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

func startMockServer(ctx context.Context, conf *config.Config, cancel context.CancelFunc) error {
	logger.Infof("=== Start server, config[%+v]", conf)

	err := database.InitRedisCli(conf.Db)
	if err != nil {
		return err
	}

	mockController, err := controller.NewMockController(conf.ServerNames, conf.UserCount,
		conf.TplFilePath, conf.IntervalMs)
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
