package kim_go_log_helper

import (
	"fmt"
	rotateLogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"log"
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

func (l *Log) Get() *logrus.Logger {
	logger := logrus.New()
	logger.SetReportCaller(true) // 开启调用者信息
	// 日志文件
	fileName := path.Join(l.FilePath, l.FileName)

	// 写入文件
	src, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalln(fmt.Errorf("写入日志文件失败: %w", err))
	}

	// 设置输出目标
	if l.Env != "release" {
		logger.Out = os.Stdout
		// 开发环境使用文本格式更易读
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
		})
	} else {
		logger.Out = src // 避免重复写入
	}

	// 设置日志级别
	logLevel, err := logrus.ParseLevel(strings.ToLower(l.Level))
	if err != nil {
		log.Fatalln(fmt.Errorf("无效的日志级别: %w", err))
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
		log.Fatalln(fmt.Errorf("创建日志切割器失败: %w", err))
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
	return logger
}
