package kim_go_log_helper

import (
	"fmt"
	rotateLogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

type Log struct {
	Env         string
	Level       string
	FilePath    string
	FileName    string
	PrettyPrint bool
}

func (l *Log) Get() (*logrus.Logger, error) {
	logger := logrus.New()
	logger.SetReportCaller(true) // 开启调用者信息

	// 创建日志目录
	if err := os.MkdirAll(l.FilePath, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	fileName := path.Join(l.FilePath, l.FileName)

	// 设置输出目标
	if l.Env != "release" {
		logger.Out = os.Stdout
		// 开发环境使用文本格式更易读
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
		})
	} else {
		logger.Out = io.Discard // 避免重复写入
	}

	// 设置日志级别
	logLevel, err := logrus.ParseLevel(strings.ToLower(l.Level))
	if err != nil {
		return nil, fmt.Errorf("无效的日志级别: %w", err)
	}
	logger.SetLevel(logLevel)

	// 配置日志切割
	logWriter, err := rotateLogs.New(
		// 分割后的文件名称
		fileName+".%Y%m%d.log",
		// 生成软链，指向最新日志文件
		rotateLogs.WithLinkName(fileName),
		// 设置最大保存时间(7天)
		rotateLogs.WithMaxAge(7*24*time.Hour),
		// 设置日志切割时间间隔(1天)
		rotateLogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		return nil, fmt.Errorf("创建日志切割器失败: %w", err)
	}

	// 简化级别映射
	writeMap := make(lfshook.WriterMap)
	for _, level := range logrus.AllLevels {
		writeMap[level] = logWriter
	}

	// 生产环境使用JSON格式
	lfHook := lfshook.NewHook(writeMap, &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     l.PrettyPrint,
	})

	logger.AddHook(lfHook)
	return logger, nil
}
