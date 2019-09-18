package logger

import "github.com/sirupsen/logrus"

var logger = NewLogger()


func GetLevel() logrus.Level {
	return logger.GetLevel()
}
func WithError(err error) *logrus.Entry {
	return logger.WithError(err)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return logger.WithFields(fields)
}

// The exported function for logger
func Debug(v ...interface{}) {
	logger.Debug(v)
}

func Info(v ...interface{}) {
	logger.Info(v)
}

func Warn(v ...interface{}) {
	logger.Warn(v)
}

func Error(v ...interface{}) {
	logger.Error(v)
}

func Fatal(v ...interface{}) {
	logger.Fatal(v)
}

func Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v)
}

func Infof(format string, v ...interface{}) {
	logger.Infof(format, v)
}

func Warnf(format string, v ...interface{}) {
	logger.Warnf(format, v)
}

func Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v)
}

func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v)
}

func IsDebugEnabled() bool {
	return logger.IsDebugEnabled()
}

func IsInfoEnabled() bool {
	return logger.IsInfoEnabled()
}

func IsWarnEnabled() bool {
	return logger.IsWarnEnabled()
}

func IsErrorEnabled() bool {
	return logger.IsErrorEnabled()
}

func IsFatalEnabled() bool {
	return logger.IsFatalEnabled()
}