package log

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

var logger *Logger

func init() {
	logger = NewLogger("logs")
	exitFunctions = []func() error{}
	go logger.listenOsSignal()
}

func DefaultLogger() *Logger {
	return logger
}

func NewLogger(outputPath string) *Logger {
	logger := Logger{}
	logger.init(outputPath)
	return &logger
}

type Logger struct {
	buffer     chan LoggerField
	exit       chan struct{}
	rwMtx      sync.RWMutex
	outputDir  string
	stdLogger  *logrus.Logger
	errLogger  *logrus.Logger
	stdoutFile *os.File
	stderrFile *os.File
	timestamp  time.Time
}

func (l *Logger) init(outputPath string) {
	if l.buffer == nil {
		l.buffer = make(chan LoggerField, 100)
		l.exit = make(chan struct{})

		go l.serve()
	}

	l.rwMtx = sync.RWMutex{}
	l.rwMtx.Lock()
	l.outputDir = outputPath
	if l.outputDir == "" {
		l.outputDir = "logs"
	}

	// 检查日志输出目录是否存在，不存在则创建
	if _, checkDirExistErr := os.Stat(l.outputDir); os.IsNotExist(checkDirExistErr) {
		if mkdirErr := os.Mkdir(l.outputDir, 0o755); mkdirErr != nil {
			panic(mkdirErr)
		}
	}

	// 创建/打开/追加日志文件
	l.timestamp = time.Now()
	stdoutFilePath := filepath.Join(l.outputDir, fmt.Sprintf("stdout_%s.log", l.timestamp.Format("2006-01-02")))
	stderrFilePath := filepath.Join(l.outputDir, fmt.Sprintf("stderr_%s.log", l.timestamp.Format("2006-01-02")))
	if stdoutFile, openStdoutFileErr := os.OpenFile(stdoutFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o755); openStdoutFileErr != nil {
		panic(openStdoutFileErr)
	} else {
		l.stdoutFile = stdoutFile
	}
	if stderrFile, openStderrFileErr := os.OpenFile(stderrFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o755); openStderrFileErr != nil {
		panic(openStderrFileErr)
	} else {
		l.stderrFile = stderrFile
	}

	// 初始化日志对象
	l.stdLogger = logrus.New()
	l.errLogger = logrus.New()
	l.stdLogger.SetOutput(l.stdoutFile)
	l.errLogger.SetOutput(l.stderrFile)
	l.stdLogger.SetLevel(logrus.DebugLevel)
	l.errLogger.SetLevel(logrus.ErrorLevel)
	l.stdLogger.SetFormatter(&logrus.JSONFormatter{})
	l.errLogger.SetFormatter(&logrus.JSONFormatter{})
	l.rwMtx.Unlock()
}

func (l *Logger) checkDateChange() {
	if time.Now().Day() != l.timestamp.Day() {
		l.rwMtx.Lock()
		_ = l.stdoutFile.Close()
		_ = l.stderrFile.Close()
		l.rwMtx.Unlock()
		l.init(l.outputDir)
	}
}

func (l *Logger) serve() {
	defer os.Exit(0)

	logFunction := func(fields LoggerField) {
		l.checkDateChange()
		l.rwMtx.RLock()
		switch fields.Level() {
		case Debug:
			l.stdLogger.WithFields(fields.EncodePayload()).Debugln(fields.Message())
		case Info:
			l.stdLogger.WithFields(fields.EncodePayload()).Infoln(fields.Message())
		case Warn:
			l.stdLogger.WithFields(fields.EncodePayload()).Warnln(fields.Message())
		case Error:
			l.stdLogger.WithFields(fields.EncodePayload()).Errorln(fields.Message())
			l.errLogger.WithFields(fields.EncodePayload()).Errorln(fields.Message())
		case Panic:
			l.stdLogger.WithFields(fields.EncodePayload()).Errorln(fields.Message())
			l.errLogger.WithFields(fields.EncodePayload()).Panicln(fields.Message())
		default:
			l.stdLogger.WithFields(fields.EncodePayload()).Infoln(fields.Message())
		}
		l.rwMtx.RUnlock()
	}

	for {
		select {
		case fields := <-l.buffer:
			logFunction(fields)
		case <-l.exit:
			for {
				select {
				case fields := <-l.buffer:
					logFunction(fields)
				default:
					// 执行额外的退出函数
					for _, function := range exitFunctions {
						if e := function(); e != nil {
							logFunction(DefaultField().WithLevel(Error).WithCaller(Internal).
								WithMessage("exit function error").WithExtra(e))
						}
					}
					return
				}
			}
		}
	}
}

func (l *Logger) listenOsSignal() {
	sgChan := make(chan os.Signal, 1)
	signal.Notify(sgChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-sgChan
	signalString := ""
	switch sig {
	case syscall.SIGTERM:
		signalString = "SIGTERM"
	case syscall.SIGINT:
		signalString = "SIGINT"
	case syscall.SIGQUIT:
		signalString = "SIGQUIT"
	}
	l.Log(DefaultField().WithLevel(Info).WithCaller(Internal).WithMessage("os signal received").WithExtra(signalString))
	l.exit <- struct{}{}
}

func (l *Logger) Log(fields LoggerField) {
	l.buffer <- fields
}
