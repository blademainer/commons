package logger

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	ENV_LOG_LEVEL = "LOG_LEVEL"

	LOG_LEVEL_DEBUG = "debug"
	LOG_LEVEL_INFO  = "info"
	LOG_LEVEL_WARN  = "warn"
	LOG_LEVEL_ERROR = "error"

	DEFAULT_LOG_LEVEL = LOG_LEVEL_DEBUG
)

func InitLogLevelFromEnv() {
	logLevel, exists := DEFAULT_LOG_LEVEL, false
	if logLevel, exists = os.LookupEnv(ENV_LOG_LEVEL); !exists {
		logLevel = DEFAULT_LOG_LEVEL
	}
	Log.SetLevel(logLevel)
}

func SetLogLevel() {
	logLevel, _ := os.LookupEnv(ENV_LOG_LEVEL)
	switch logLevel {
	case LOG_LEVEL_DEBUG:
		logrus.SetLevel(logrus.DebugLevel)
	case LOG_LEVEL_INFO:
		logrus.SetLevel(logrus.InfoLevel)
	case LOG_LEVEL_WARN:
		logrus.SetLevel(logrus.WarnLevel)
	case LOG_LEVEL_ERROR:
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.DebugLevel)
	}
}

// the config information for logging
type LoggerConfig struct {
	Level      string `yaml:"Level"`
	FileName   string `yaml:"FileName"`
	MaxBackups int    `yaml:"MaxBackups"`
	MaxSize    string `yaml:"MaxSize"`
	MaxAge     int    `yaml:"MaxAge"`
	Hooks      []string // the name of plugin hook
}

// The Option for logger
type Option struct {
	Level      logrus.Level
	FileName   string // the file name for logging
	MaxSize    int    // the maximum size of a log file
	MaxBackups int    // the maximum number of backup files
	MaxAge     int    // the maximum days for log files
}

// define the logger format string
type defaultLoggerFormatter struct {
}

// 2016-05-20 09:10:20 DEBUG log.go(2011:23): backend etcd
func (formatter defaultLoggerFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	return []byte(fmt.Sprintf("%s %s %s:%d %s\n", timestamp, strings.ToUpper(entry.Level.String()), entry.Data["file"], os.Getpid(), entry.Message)), nil
}

// set the log level for logging
func (l *Logger) SetLevel(level string) {
	if logLevel, err := logrus.ParseLevel(level); err != nil {
		logrus.Errorf("invalid level: %s", level)
	} else {
		//logrus.SetLevel(logLevel)
		l.logger.Level = logLevel
	}
}

// sets the standard logger formatter.
func (l *Logger) SetFormatter(formatter logrus.Formatter) {
	l.logger.Formatter = formatter
	//logrus.SetFormatter(formatter)
}

// Add a hook to an instance of logger
func (l *Logger) AddHook(hook logrus.Hook) {
	//logrus.AddHook(hook)
	l.logger.Hooks.Add(hook)
}

// set the output for logging
func (l *Logger) SetOutput(out io.Writer) {
	if out != nil {
		l.logger.Out = out
	}
}

func (config *LoggerConfig) GetOption() (*Option, error) {
	unitMap := map[string]int{
		"KB": 1,
		"MB": 1024,
		"GB": 1024 * 1024,
		"TB": 1024 * 1024 * 1024,
	}
	unit := 1024
	mslen := len(config.MaxSize)
	if mslen < 2 {
		return nil, errors.New("rolling log MaxSize format is : 100MB  unit support:KB MB GB TB")
	}
	sunit := config.MaxSize[mslen-2:]
	var ok bool
	if unit, ok = unitMap[sunit]; !ok {
		return nil, errors.New("rolling log MaxSize format is : 100MB  unit support:KB MB GB TB")
	}
	logLevel, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("unknow log level %s", config.Level)
	}
	maxsizeUnit, _ := strconv.ParseUint(config.MaxSize[:mslen-2], 10, 64)
	Log.Debugf("rolling file size:%d%s", maxsizeUnit, sunit)
	maxsize := int(int64(maxsizeUnit) * int64(unit) / int64(1024))
	//convert maxsize to MB
	option := &Option{
		Level:      logLevel,
		FileName:   config.FileName,
		MaxBackups: config.MaxBackups,
		MaxSize:    maxsize,
		MaxAge:     config.MaxAge,
	}
	return option, nil

}

// create a new lumberjack logger
func (l *Logger) Init(config LoggerConfig) error {
	opt, err := config.GetOption()
	if err != nil {
		return err
	}
	l.logger.Level = opt.Level
	if opt.FileName == "" {
		l.SetOutput(os.Stdout)
		l.Error("log to stderr")
		return nil
	}

	l.SetOutput(&lumberjack.Logger{
		Filename:   opt.FileName,
		MaxSize:    opt.MaxSize,
		MaxBackups: opt.MaxBackups,
		MaxAge:     opt.MaxAge,
	})
	return nil
}

// export the *Logger object
type Logger struct {
	logger *logrus.Logger
}

// new a logger for logging
func NewLogger() *Logger {
	// log as custom formatter instead of the default ASCII formatter.
	//if logger == nil {
	//	logger = logrus.StandardLogger()
	//}
	log := &Logger{logger: logrus.New()}
	log.SetFormatter(&defaultLoggerFormatter{})

	return log
}
func (l *Logger) WithCaller(calllevel int) *logrus.Entry {
	_, file, line, _ := runtime.Caller(calllevel)
	d, f := filepath.Split(file)
	path := fmt.Sprintf("%s/%s", filepath.Base(d), f)
	return l.logger.WithField("file", fmt.Sprintf("(%s:%d)", path, line))

}
func (l *Logger) GetLevel() logrus.Level {
	return l.logger.Level
}
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.WithCaller(2).WithError(err)
}

func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.logger.WithFields(fields)
}

// The exported function for logger
func (l *Logger) Debug(v ...interface{}) {
	if l.IsDebugEnabled() {
		l.WithCaller(2).Debug(v...)
	}
}

func (l *Logger) Info(v ...interface{}) {
	if l.IsInfoEnabled() {
		l.WithCaller(2).Info(v...)
	}
}

func (l *Logger) Warn(v ...interface{}) {
	if l.IsWarnEnabled() {
		l.WithCaller(2).Warn(v...)
	}
}

func (l *Logger) Error(v ...interface{}) {
	if l.IsErrorEnabled() {
		l.WithCaller(2).Error(v...)
	}
}

func (l *Logger) Fatal(v ...interface{}) {
	if l.IsFatalEnabled() {
		l.WithCaller(2).Fatal(v...)
	}
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.IsDebugEnabled() {
		l.WithCaller(2).Debugf(format, v...)
	}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	if l.IsInfoEnabled() {
		l.WithCaller(2).Infof(format, v...)
	}
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.IsWarnEnabled() {
		l.WithCaller(2).Warnf(format, v...)
	}
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.IsErrorEnabled() {
		l.WithCaller(2).Errorf(format, v...)
	}
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	if l.IsFatalEnabled() {
		l.WithCaller(2).Fatalf(format, v...)
	}
}

func (l *Logger) IsDebugEnabled() bool {
	return l.GetLevel() >= logrus.DebugLevel
}

func (l *Logger) IsInfoEnabled() bool {
	return l.GetLevel() >= logrus.InfoLevel
}

func (l *Logger) IsWarnEnabled() bool {
	return l.GetLevel() >= logrus.WarnLevel
}

func (l *Logger) IsErrorEnabled() bool {
	return l.GetLevel() >= logrus.ErrorLevel
}

func (l *Logger) IsFatalEnabled() bool {
	return l.GetLevel() >= logrus.FatalLevel
}

var Log = NewLogger()
var AccessLog = NewLogger()

func Access(fields map[string]interface{}, msg string) {
	if AccessLog.GetLevel() >= logrus.InfoLevel {
		AccessLog.WithFields(logrus.Fields(fields)).Infof(msg)
	}
}
