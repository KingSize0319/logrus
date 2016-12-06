package logrus

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func LogAndAssertJSON(t *testing.T, log func(*Logger), assertions func(fields Fields)) {
	var buffer bytes.Buffer
	var fields Fields

	logger := New()
	logger.Out = &buffer
	logger.Formatter = new(JSONFormatter)

	log(logger)

	err := json.Unmarshal(buffer.Bytes(), &fields)
	assert.Nil(t, err)

	assertions(fields)
}

func LogAndAssertText(t *testing.T, log func(*Logger), assertions func(fields map[string]string)) {
	var buffer bytes.Buffer

	logger := New()
	logger.Out = &buffer
	logger.Formatter = &TextFormatter{
		DisableColors: true,
	}

	log(logger)

	fields := make(map[string]string)
	for _, kv := range strings.Split(buffer.String(), " ") {
		if !strings.Contains(kv, "=") {
			continue
		}
		kvArr := strings.Split(kv, "=")
		key := strings.TrimSpace(kvArr[0])
		val := kvArr[1]
		if kvArr[1][0] == '"' {
			var err error
			val, err = strconv.Unquote(val)
			assert.NoError(t, err)
		}
		fields[key] = val
	}
	assertions(fields)
}

// TestReportCaller verifies that when ReportCaller is set, the 'func' field
// is added, and when it is unset it is not set or modified
func TestReportCaller(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.ReportCaller = false
		log.Print("testNoCaller")
	}, func(fields Fields) {
		assert.Equal(t, "testNoCaller", fields["msg"])
		assert.Equal(t, "info", fields["level"])
		assert.Equal(t, nil, fields["func"])
	})

	LogAndAssertJSON(t, func(log *Logger) {
		log.ReportCaller = true
		log.Print("testWithCaller")
	}, func(fields Fields) {
		assert.Equal(t, "testWithCaller", fields["msg"])
		assert.Equal(t, "info", fields["level"])
		assert.Equal(t, "testing.tRunner", fields["func"])
	})
}

func TestPrint(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Print("test")
	}, func(fields Fields) {
		assert.Equal(t, "test", fields["msg"])
		assert.Equal(t, "info", fields["level"])
	})
}

func TestInfo(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Info("test")
	}, func(fields Fields) {
		assert.Equal(t, "test", fields["msg"])
		assert.Equal(t, "info", fields["level"])
	})
}

func TestWarn(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Warn("test")
	}, func(fields Fields) {
		assert.Equal(t, "test", fields["msg"])
		assert.Equal(t, "warning", fields["level"])
	})
}

func TestInfolnShouldAddSpacesBetweenStrings(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Infoln("test", "test")
	}, func(fields Fields) {
		assert.Equal(t, "test test", fields["msg"])
	})
}

func TestInfolnShouldAddSpacesBetweenStringAndNonstring(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Infoln("test", 10)
	}, func(fields Fields) {
		assert.Equal(t, "test 10", fields["msg"])
	})
}

func TestInfolnShouldAddSpacesBetweenTwoNonStrings(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Infoln(10, 10)
	}, func(fields Fields) {
		assert.Equal(t, "10 10", fields["msg"])
	})
}

func TestInfoShouldAddSpacesBetweenTwoNonStrings(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Infoln(10, 10)
	}, func(fields Fields) {
		assert.Equal(t, "10 10", fields["msg"])
	})
}

func TestInfoShouldNotAddSpacesBetweenStringAndNonstring(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Info("test", 10)
	}, func(fields Fields) {
		assert.Equal(t, "test10", fields["msg"])
	})
}

func TestInfoShouldNotAddSpacesBetweenStrings(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Info("test", "test")
	}, func(fields Fields) {
		assert.Equal(t, "testtest", fields["msg"])
	})
}

func TestWithFieldsShouldAllowAssignments(t *testing.T) {
	var buffer bytes.Buffer
	var fields Fields

	logger := New()
	logger.Out = &buffer
	logger.Formatter = new(JSONFormatter)

	localLog := logger.WithFields(Fields{
		"key1": "value1",
	})

	localLog.WithField("key2", "value2").Info("test")
	err := json.Unmarshal(buffer.Bytes(), &fields)
	assert.Nil(t, err)

	assert.Equal(t, "value2", fields["key2"])
	assert.Equal(t, "value1", fields["key1"])

	buffer = bytes.Buffer{}
	fields = Fields{}
	localLog.Info("test")
	err = json.Unmarshal(buffer.Bytes(), &fields)
	assert.Nil(t, err)

	_, ok := fields["key2"]
	assert.Equal(t, false, ok)
	assert.Equal(t, "value1", fields["key1"])
}

func TestUserSuppliedFieldDoesNotOverwriteDefaults(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.WithField("msg", "hello").Info("test")
	}, func(fields Fields) {
		assert.Equal(t, "test", fields["msg"])
	})
}

func TestUserSuppliedMsgFieldHasPrefix(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.WithField("msg", "hello").Info("test")
	}, func(fields Fields) {
		assert.Equal(t, "test", fields["msg"])
		assert.Equal(t, "hello", fields["fields.msg"])
	})
}

func TestUserSuppliedTimeFieldHasPrefix(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.WithField("time", "hello").Info("test")
	}, func(fields Fields) {
		assert.Equal(t, "hello", fields["fields.time"])
	})
}

func TestUserSuppliedLevelFieldHasPrefix(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.WithField("level", 1).Info("test")
	}, func(fields Fields) {
		assert.Equal(t, "info", fields["level"])
		assert.Equal(t, 1.0, fields["fields.level"]) // JSON has floats only
	})
}

func TestDefaultFieldsAreNotPrefixed(t *testing.T) {
	LogAndAssertText(t, func(log *Logger) {
		ll := log.WithField("herp", "derp")
		ll.Info("hello")
		ll.Info("bye")
	}, func(fields map[string]string) {
		for _, fieldName := range []string{"fields.level", "fields.time", "fields.msg"} {
			if _, ok := fields[fieldName]; ok {
				t.Fatalf("should not have prefixed %q: %v", fieldName, fields)
			}
		}
	})
}

func TestDoubleLoggingDoesntPrefixPreviousFields(t *testing.T) {

	var buffer bytes.Buffer
	var fields Fields

	logger := New()
	logger.Out = &buffer
	logger.Formatter = new(JSONFormatter)

	llog := logger.WithField("context", "eating raw fish")

	llog.Info("looks delicious")

	err := json.Unmarshal(buffer.Bytes(), &fields)
	assert.NoError(t, err, "should have decoded first message")
	assert.Equal(t, len(fields), 4, "should only have msg/time/level/context fields")
	assert.Equal(t, fields["msg"], "looks delicious")
	assert.Equal(t, fields["context"], "eating raw fish")

	buffer.Reset()

	llog.Warn("omg it is!")

	err = json.Unmarshal(buffer.Bytes(), &fields)
	assert.NoError(t, err, "should have decoded second message")
	assert.Equal(t, len(fields), 4, "should only have msg/time/level/context fields")
	assert.Equal(t, "omg it is!", fields["msg"])
	assert.Equal(t, "eating raw fish", fields["context"])
	assert.Nil(t, fields["fields.msg"], "should not have prefixed previous `msg` entry")

}

func TestNestedLoggingReportsCorrectCaller(t *testing.T) {
	var buffer bytes.Buffer
	var fields Fields

	logger := New()
	logger.Out = &buffer
	logger.Formatter = new(JSONFormatter)
	logger.ReportCaller = true

	llog := logger.WithField("context", "eating raw fish")

	llog.Info("looks delicious")

	err := json.Unmarshal(buffer.Bytes(), &fields)
	assert.NoError(t, err, "should have decoded first message")
	assert.Equal(t, len(fields), 5, "should have msg/time/level/func/context fields")
	assert.Equal(t, "looks delicious", fields["msg"])
	assert.Equal(t, "eating raw fish", fields["context"])
	assert.Equal(t, "testing.tRunner", fields["func"])

	buffer.Reset()

	logger.WithFields(Fields{
		"Clyde": "Stubblefield",
	}).WithFields(Fields{
		"Jab'o": "Starks",
	}).WithFields(Fields{
		"uri": "https://www.youtube.com/watch?v=V5DTznu-9v0",
	}).WithFields(Fields{
		"func": "y drummer",
	}).WithFields(Fields{
		"James": "Brown",
	}).Print("The hardest workin' man in show business")

	err = json.Unmarshal(buffer.Bytes(), &fields)
	assert.NoError(t, err, "should have decoded second message")
	assert.Equal(t, 10, len(fields), "should have all builtin fields plus foo,bar,baz,...")
	assert.Equal(t, "Stubblefield", fields["Clyde"])
	assert.Equal(t, "Starks", fields["Jab'o"])
	assert.Equal(t, "https://www.youtube.com/watch?v=V5DTznu-9v0", fields["uri"])
	assert.Equal(t, "y drummer", fields["fields.func"])
	assert.Equal(t, "Brown", fields["James"])
	assert.Equal(t, "The hardest workin' man in show business", fields["msg"])
	assert.Nil(t, fields["fields.msg"], "should not have prefixed previous `msg` entry")
	assert.Equal(t, "testing.tRunner", fields["func"])

	logger.ReportCaller = false // return to default value
}

func logLoop(iterations int, reportCaller bool) {
	var buffer bytes.Buffer

	logger := New()
	logger.Out = &buffer
	logger.Formatter = new(JSONFormatter)
	logger.ReportCaller = reportCaller

	for i := 0; i < iterations; i++ {
		logger.Infof("round %d of %d", i, iterations)
	}
}

// Assertions for upper bounds to reporting overhead
func TestCallerReportingOverhead(t *testing.T) {
	iterations := 5000
	before := time.Now()
	logLoop(iterations, false)
	during := time.Now()
	logLoop(iterations, true)
	after := time.Now()

	elapsedNotReporting := during.Sub(before).Nanoseconds()
	elapsedReporting := after.Sub(during).Nanoseconds()

	maxDelta := 1 * time.Second
	assert.WithinDuration(t, during, before, maxDelta,
		"%d log calls without caller name lookup takes less than %d second(s) (was %d nanoseconds)",
		iterations, maxDelta.Seconds(), elapsedNotReporting)
	assert.WithinDuration(t, after, during, maxDelta,
		"%d log calls without caller name lookup takes less than %d second(s) (was %d nanoseconds)",
		iterations, maxDelta.Seconds(), elapsedReporting)
}

// benchmarks for both with and without caller-function reporting
func BenchmarkWithoutCallerTracing(b *testing.B) {
	for i := 0; i < b.N; i++ {
		logLoop(1000, false)
	}
}

func BenchmarkWithCallerTracing(b *testing.B) {
	for i := 0; i < b.N; i++ {
		logLoop(1000, true)
	}
}

func TestConvertLevelToString(t *testing.T) {
	assert.Equal(t, "debug", DebugLevel.String())
	assert.Equal(t, "info", InfoLevel.String())
	assert.Equal(t, "warning", WarnLevel.String())
	assert.Equal(t, "error", ErrorLevel.String())
	assert.Equal(t, "fatal", FatalLevel.String())
	assert.Equal(t, "panic", PanicLevel.String())
}

func TestParseLevel(t *testing.T) {
	l, err := ParseLevel("panic")
	assert.Nil(t, err)
	assert.Equal(t, PanicLevel, l)

	l, err = ParseLevel("PANIC")
	assert.Nil(t, err)
	assert.Equal(t, PanicLevel, l)

	l, err = ParseLevel("fatal")
	assert.Nil(t, err)
	assert.Equal(t, FatalLevel, l)

	l, err = ParseLevel("FATAL")
	assert.Nil(t, err)
	assert.Equal(t, FatalLevel, l)

	l, err = ParseLevel("error")
	assert.Nil(t, err)
	assert.Equal(t, ErrorLevel, l)

	l, err = ParseLevel("ERROR")
	assert.Nil(t, err)
	assert.Equal(t, ErrorLevel, l)

	l, err = ParseLevel("warn")
	assert.Nil(t, err)
	assert.Equal(t, WarnLevel, l)

	l, err = ParseLevel("WARN")
	assert.Nil(t, err)
	assert.Equal(t, WarnLevel, l)

	l, err = ParseLevel("warning")
	assert.Nil(t, err)
	assert.Equal(t, WarnLevel, l)

	l, err = ParseLevel("WARNING")
	assert.Nil(t, err)
	assert.Equal(t, WarnLevel, l)

	l, err = ParseLevel("info")
	assert.Nil(t, err)
	assert.Equal(t, InfoLevel, l)

	l, err = ParseLevel("INFO")
	assert.Nil(t, err)
	assert.Equal(t, InfoLevel, l)

	l, err = ParseLevel("debug")
	assert.Nil(t, err)
	assert.Equal(t, DebugLevel, l)

	l, err = ParseLevel("DEBUG")
	assert.Nil(t, err)
	assert.Equal(t, DebugLevel, l)

	l, err = ParseLevel("invalid")
	assert.Equal(t, "not a valid logrus Level: \"invalid\"", err.Error())
}

func TestGetSetLevelRace(t *testing.T) {
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				SetLevel(InfoLevel)
			} else {
				GetLevel()
			}
		}(i)

	}
	wg.Wait()
}

func TestLoggingRace(t *testing.T) {
	logger := New()

	var wg sync.WaitGroup
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			logger.Info("info")
			wg.Done()
		}()
	}
	wg.Wait()
}

// Compile test
func TestLogrusInterface(t *testing.T) {
	var buffer bytes.Buffer
	fn := func(l FieldLogger) {
		b := l.WithField("key", "value")
		b.Debug("Test")
	}
	// test logger
	logger := New()
	logger.Out = &buffer
	fn(logger)

	// test Entry
	e := logger.WithField("another", "value")
	fn(e)
}
