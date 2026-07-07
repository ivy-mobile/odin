package v2_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	logv2 "github.com/ivy-mobile/odin/xutil/xlog/v2"
)

var logger logv2.Logger

func init() {
	logger = logv2.New()
}

func TestDebug(_ *testing.T) {
	logger.Debug().Int64("i64", 1).Msg("test1")
	logger.Debug().Int("int", 1).Msg("test1")
	logger.Debug().Uint64("u64", 1).Msg("test1")
	logger.Debug().Float("f64", 1).Msg("test1")
	logger.Debug().Bool("bool", true).Msg("test1")
	logger.Debug().Time("t", time.Now()).Msg("test1")
	logger.Debug().Duration("d", time.Second).Msg("test1")
	logger.Debug().Any("any", struct{}{}).Msg("test1")
	logger.Debug().Err(errors.New("test error")).Msg("test1")

	logger.Debug().Int64("i64", 1).Msgf("test2")
	logger.Debug().Int("int", 1).Msgf("test2")
	logger.Debug().Uint64("u64", 1).Msgf("test2")
	logger.Debug().Float("f64", 1).Msgf("test2")
	logger.Debug().Bool("bool", true).Msgf("test2")
	logger.Debug().Time("t", time.Now()).Msgf("test2")
	logger.Debug().Duration("d", time.Second).Msgf("test2")
	logger.Debug().Any("any", struct{}{}).Msgf("test2")
	logger.Debug().Any("err", errors.New("test error")).Msgf("test2")
}

func TestInfo(_ *testing.T) {
	logger.Info().Int64("i64", 1).Msg("test1")
	logger.Info().Int("int", 1).Msg("test1")
	logger.Info().Uint64("u64", 1).Msg("test1")
	logger.Info().Float("f64", 1).Msg("test1")
	logger.Info().Bool("bool", true).Msg("test1")
	logger.Info().Time("t", time.Now()).Msg("test1")
	logger.Info().Duration("d", time.Second).Msg("test1")
	logger.Info().Any("any", struct{}{}).Msg("test1")
	logger.Info().Err(errors.New("test error")).Msg("test1")

	logger.Info().Int64("i64", 1).Msgf("test2")
	logger.Info().Int("int", 1).Msgf("test2")
	logger.Info().Uint64("u64", 1).Msgf("test2")
	logger.Info().Float("f64", 1).Msgf("test2")
	logger.Info().Bool("bool", true).Msgf("test2")
	logger.Info().Time("t", time.Now()).Msgf("test2")
	logger.Info().Duration("d", time.Second).Msgf("test2")
	logger.Info().Any("any", struct{}{}).Msgf("test2")
	logger.Info().Err(errors.New("test error")).Msgf("test2")
}

func TestWarn(_ *testing.T) {
	logger.Warn().Int64("i64", 1).Msg("test1")
	logger.Warn().Int("int", 1).Msg("test1")
	logger.Warn().Uint64("u64", 1).Msg("test1")
	logger.Warn().Float("f64", 1).Msg("test1")
	logger.Warn().Bool("bool", true).Msg("test1")
	logger.Warn().Time("t", time.Now()).Msg("test1")
	logger.Warn().Duration("d", time.Second).Msg("test1")
	logger.Warn().Any("any", struct{}{}).Msg("test1")
	logger.Warn().Any("err", errors.New("test error")).Msg("test1")

	logger.Warn().Int64("i64", 1).Msgf("test2")
	logger.Warn().Int("int", 1).Msgf("test2")
	logger.Warn().Uint64("u64", 1).Msgf("test2")
	logger.Warn().Float("f64", 1).Msgf("test2")
	logger.Warn().Bool("bool", true).Msgf("test2")
	logger.Warn().Time("t", time.Now()).Msgf("test2")
	logger.Warn().Duration("d", time.Second).Msgf("test2")
	logger.Warn().Any("any", struct{}{}).Msgf("test2")
	logger.Warn().Any("err", errors.New("test error")).Msgf("test2")
}

func TestError(_ *testing.T) {
	logger.Error().Int64("i64", 1).Msg("test1")
	logger.Error().Int("int", 1).Msg("test1")
	logger.Error().Uint64("u64", 1).Msg("test1")
	logger.Error().Float("f64", 1).Msg("test1")
	logger.Error().Bool("bool", true).Msg("test1")
	logger.Error().Time("t", time.Now()).Msg("test1")
	logger.Error().Duration("d", time.Second).Msg("test1")
	logger.Error().Any("any", struct{}{}).Msg("test1")
	logger.Error().Any("err", errors.New("test error")).Msg("test1")

	logger.Error().Int64("i64", 1).Msgf("test2")
	logger.Error().Int("int", 1).Msgf("test2")
	logger.Error().Uint64("u64", 1).Msgf("test2")
	logger.Error().Float("f64", 1).Msgf("test2")
	logger.Error().Bool("bool", true).Msgf("test2")
	logger.Error().Time("t", time.Now()).Msgf("test2")
	logger.Error().Duration("d", time.Second).Msgf("test2")
	logger.Error().Any("any", struct{}{}).Msgf("test2")
	logger.Error().Any("err", errors.New("test error")).Msgf("test2")
}

func BenchmarkLog(b *testing.B) {
	log := logv2.New(
		logv2.WithLevel("info"),
		logv2.WithMode("console"),
	)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		log.Info().Msg("test1")
	}
}

func TestOlog(_ *testing.T) {
	log := logv2.New(
		logv2.WithLevel("info"),
		logv2.WithMode("file"),
	)
	log.Debug().Str("Name", "Art").Int("Age", 18).Msg("debug test")
	log.Info().Str("Name", "Art").Int("Age", 18).Msg("info test")
	log.Warn().Str("Name", "Art").Int("Age", 18).Msg("warn test")
	log.Error().Str("Name", "Art").Int("Age", 18).Msg("error test")

	log.With("Module", "Test1").With("Func", "TestOlog").Info().Msg("model test")
	log.Info().Str("Name", "Art").Int("Age", 18).Msg("info test")
}

func TestObserveModeFormatOutputs(t *testing.T) {
	tmpDir := t.TempDir()

	t.Log("console target + text format")
	logv2.New(
		logv2.WithMode(logv2.ModeConsole),
		logv2.WithFormat(logv2.FormatText),
	).With("case", "console-text").Info().
		Str("name", "Art").
		Int("age", 18).
		Bool("active", true).
		Msg("observe xlog output")

	t.Log("console target + json format")
	logv2.New(
		logv2.WithMode(logv2.ModeConsole),
		logv2.WithFormat(logv2.FormatJSON),
	).With("case", "console-json").Info().
		Str("name", "Art").
		Int("age", 18).
		Bool("active", true).
		Msg("observe xlog output")

	fileJSON := filepath.Join(tmpDir, "xlog-json.log")
	t.Logf("file target + json format: %s", fileJSON)
	logv2.New(
		logv2.WithMode(logv2.ModeFile),
		logv2.WithFormat(logv2.FormatJSON),
		logv2.WithFile(fileJSON, 1, 1, 1, false, true),
	).With("case", "file-json").Info().
		Str("name", "Art").
		Int("age", 18).
		Bool("active", true).
		Msg("observe xlog output")
	logFileForObservation(t, fileJSON)

	fileText := filepath.Join(tmpDir, "xlog-text.log")
	t.Logf("file target + text format: %s", fileText)
	logv2.New(
		logv2.WithMode(logv2.ModeFile),
		logv2.WithFormat(logv2.FormatText),
		logv2.WithFile(fileText, 1, 1, 1, false, true),
	).With("case", "file-text").Info().
		Str("name", "Art").
		Int("age", 18).
		Bool("active", true).
		Msg("observe xlog output")
	logFileForObservation(t, fileText)
}

func logFileForObservation(t *testing.T, filename string) {
	t.Helper()

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Logf("read log file failed: %v", err)
		return
	}
	t.Logf("%s content:\n%s", filename, string(content))
}
