package ldlogtest

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldlog"
)

type mockLogTestValues struct {
	level     ldlog.LogLevel
	nextLevel ldlog.LogLevel
	println   func(ldlog.Loggers, ...interface{})
	printf    func(ldlog.Loggers, string, ...interface{})
}

var allTestValues = []mockLogTestValues{
	{ldlog.Debug, ldlog.Info, ldlog.Loggers.Debug, ldlog.Loggers.Debugf},
	{ldlog.Info, ldlog.Warn, ldlog.Loggers.Info, ldlog.Loggers.Infof},
	{ldlog.Warn, ldlog.Error, ldlog.Loggers.Warn, ldlog.Loggers.Warnf},
	{ldlog.Error, ldlog.None, ldlog.Loggers.Error, ldlog.Loggers.Errorf},
}

func TestMockLogCapturesMessagesForLevel(t *testing.T) {
	for _, v := range allTestValues {
		t.Run(v.level.String(), func(t *testing.T) {
			m := NewMockLog()
			m.Loggers.SetMinLevel(v.level)
			v.println(m.Loggers, "hello")
			v.printf(m.Loggers, "yes: %t", true)

			m.Loggers.SetMinLevel(v.nextLevel)
			v.println(m.Loggers, "shouldn't see this")

			o1 := m.GetOutput(v.level)
			assert.Equal(t, []string{"hello", "yes: true"}, o1)

			o2 := m.GetAllOutput()
			assert.Equal(t, []MockLogItem{
				{v.level, "hello"},
				{v.level, "yes: true"},
			}, o2)
		})
	}
}

func TestMessageMatching(t *testing.T) {
	m := NewMockLog()
	m.Loggers.Info("first")
	m.Loggers.Info("second")
	m.Loggers.Warn("third")

	testShouldFail := func(t *testing.T, action func(*testing.T)) {
		var tt testing.T
		action(&tt)
		assert.True(t, tt.Failed(), "test should have failed")
	}

	shouldMatch := func(t *testing.T, level ldlog.LogLevel, pattern string) {
		assert.True(t, m.HasMessageMatch(level, pattern))
		m.AssertMessageMatch(t, true, level, pattern)
		testShouldFail(t, func(tt *testing.T) { m.AssertMessageMatch(tt, false, level, pattern) })
	}

	shouldNotMatch := func(t *testing.T, level ldlog.LogLevel, pattern string) {
		assert.False(t, m.HasMessageMatch(level, pattern))
		m.AssertMessageMatch(t, false, level, pattern)
		testShouldFail(t, func(tt *testing.T) { m.AssertMessageMatch(tt, true, level, pattern) })
	}

	shouldMatch(t, ldlog.Info, ".econd")
	shouldMatch(t, ldlog.Warn, "t.*")
	shouldNotMatch(t, ldlog.Info, "third")
	shouldNotMatch(t, ldlog.Error, ".")
}

func TestDump(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	m := NewMockLog()
	m.Loggers.Info("first")
	m.Loggers.Warn("second")
	m.Dump(buf)
	assert.Equal(t, "Info: first\nWarn: second\n", string(buf.Bytes()))
}
