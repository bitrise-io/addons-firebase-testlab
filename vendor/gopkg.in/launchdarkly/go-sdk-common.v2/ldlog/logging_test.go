package ldlog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type logSink struct {
	output []string
}

func (l *logSink) Println(values ...interface{}) {
	l.output = append(l.output, strings.TrimSpace(fmt.Sprintln(values...)))
}

func (l *logSink) Printf(format string, values ...interface{}) {
	l.output = append(l.output, fmt.Sprintf(format, values...))
}

func TestCanWriteToUnconfiguredLogger(t *testing.T) {
	l := Loggers{}
	l.Warn("test message, please ignore") // just testing that we don't get a nil pointer
}

func TestLevelIsInfoByDefault(t *testing.T) {
	ls := logSink{}
	l := Loggers{}
	assert.Equal(t, Info, l.GetMinLevel())
	assert.False(t, l.IsDebugEnabled())

	l.SetBaseLogger(&ls)
	l.Debug("0")
	l.Debugf("%s!", "1")
	l.Info("2")
	l.Infof("%s!", "3")
	l.Warn("4")
	l.Warnf("%s!", "5")
	l.Error("6")
	l.Errorf("%s!", "7")

	assert.Equal(t, []string{"INFO: 2", "INFO: 3!", "WARN: 4", "WARN: 5!", "ERROR: 6", "ERROR: 7!"}, ls.output)
}

func TestCanSetLevel(t *testing.T) {
	ls := logSink{}
	l := Loggers{}

	l.SetBaseLogger(&ls)
	l.SetMinLevel(Error)
	assert.Equal(t, Error, l.GetMinLevel())
	assert.False(t, l.IsDebugEnabled())

	l.Debug("0")
	l.Debugf("%s!", "1")
	l.Info("2")
	l.Infof("%s!", "3")
	l.Warn("4")
	l.Warnf("%s!", "5")
	l.Error("6")
	l.Errorf("%s!", "7")
	assert.Equal(t, []string{"ERROR: 6", "ERROR: 7!"}, ls.output)

	l.SetMinLevel(Debug)
	assert.Equal(t, Debug, l.GetMinLevel())
	assert.True(t, l.IsDebugEnabled())

	l.Debug("8")
	l.Debugf("%s!", "9")
	assert.Equal(t, []string{"ERROR: 6", "ERROR: 7!", "DEBUG: 8", "DEBUG: 9!"}, ls.output)
}

func TestCanSetLoggerForSpecificLevel(t *testing.T) {
	lsMain := logSink{}
	lsWarn := logSink{}
	l := Loggers{}
	l.SetBaseLoggerForLevel(Warn, &lsWarn)
	l.SetBaseLogger(&lsMain)
	l.Info("a")
	l.Warn("b")
	assert.Equal(t, []string{"INFO: a"}, lsMain.output)
	assert.Equal(t, []string{"WARN: b"}, lsWarn.output)
}

func TestCanGetLoggerForSpecificLevel(t *testing.T) {
	ls := logSink{}
	l := Loggers{}
	l.SetBaseLogger(&ls)
	l.ForLevel(Info).Println("a")
	l.ForLevel(Warn).Println("b")
	l.ForLevel(LogLevel(99)).Println("ignore")
	assert.Equal(t, []string{"INFO: a", "WARN: b"}, ls.output)
}

func TestSetBaseLoggerForLevelWithNilReferenceRestoresMainBaseLogger(t *testing.T) {
	lsMain := logSink{}
	lsWarn := logSink{}
	l := Loggers{}
	l.SetBaseLoggerForLevel(Warn, &lsWarn)
	l.SetBaseLogger(&lsMain)
	l.SetBaseLoggerForLevel(Warn, nil)
	l.Warn("x")
	assert.Equal(t, []string{"WARN: x"}, lsMain.output)
}

func TestSetBaseLoggerWithNilReferenceDoesNothing(t *testing.T) {
	ls := logSink{}
	l := Loggers{}
	l.SetBaseLogger(&ls)
	l.SetBaseLogger(nil)
	l.Info("x")
	assert.Equal(t, []string{"INFO: x"}, ls.output)
}

func TestInit(t *testing.T) {
	l := Loggers{}
	assert.False(t, l.inited)
	l.Init()
	assert.True(t, l.inited)
}

func TestCallingInitAgainDoesNotOverrideMinLevel(t *testing.T) {
	l := Loggers{}
	l.SetMinLevel(Error)
	l.Init()
	assert.Equal(t, Error, l.minLevel)
}

func TestNewDefaultLoggers(t *testing.T) {
	l := NewDefaultLoggers()
	assert.Equal(t, Info, l.minLevel)
	assert.True(t, l.inited)
}

func TestNewDisabledLoggers(t *testing.T) {
	l := NewDisabledLoggers()
	assert.Equal(t, None, l.minLevel)
}

func TestSetPrefix(t *testing.T) {
	ls := logSink{}
	l := Loggers{}
	l.SetBaseLogger(&ls)
	l.SetPrefix("my-prefix")
	l.Info("here's a message")
	assert.Equal(t, []string{"INFO: my-prefix here's a message"}, ls.output)
}

func TestPrintlnWithMultipleValues(t *testing.T) {
	ls := logSink{}
	l := Loggers{}
	l.SetBaseLogger(&ls)
	l.Info("a", "b", "c")
	assert.Equal(t, []string{"INFO: a b c"}, ls.output)
}

func TestLevelName(t *testing.T) {
	for level, s := range map[LogLevel]string{
		Debug:         "Debug",
		Info:          "Info",
		Warn:          "Warn",
		Error:         "Error",
		None:          "None",
		LogLevel(999): "?",
	} {
		assert.Equal(t, s, level.Name())
		assert.Equal(t, s, level.String())
	}
}
