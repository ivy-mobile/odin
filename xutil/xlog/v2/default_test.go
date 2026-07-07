package v2

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"gopkg.in/natefinch/lumberjack.v2"
)

func TestOutputFormatDefaultsByMode(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want string
	}{
		{
			name: "console target defaults to json format",
			mode: ModeConsole,
			want: FormatJSON,
		},
		{
			name: "file target defaults to json format",
			mode: ModeFile,
			want: FormatJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := defaultOptions()
			WithMode(tt.mode)(ops)

			require.Equal(t, tt.want, outputFormat(ops))
		})
	}
}

func TestWithFormatOverridesModeDefault(t *testing.T) {
	tests := []struct {
		name   string
		mode   string
		format string
		want   string
	}{
		{
			name:   "console can output json format",
			mode:   ModeConsole,
			format: FormatJSON,
			want:   FormatJSON,
		},
		{
			name:   "file can output text format",
			mode:   ModeFile,
			format: FormatText,
			want:   FormatText,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := defaultOptions()
			WithMode(tt.mode)(ops)
			WithFormat(tt.format)(ops)

			require.Equal(t, tt.want, outputFormat(ops))
		})
	}
}

func TestWithModeAcceptsOnlyOutputTargets(t *testing.T) {
	ops := defaultOptions()
	WithMode(ModeFile)(ops)

	WithMode(FormatJSON)(ops)

	require.Equal(t, ModeFile, ops.mode)
}

func TestWithFormatAcceptsOnlyOutputFormats(t *testing.T) {
	ops := defaultOptions()
	WithFormat(FormatJSON)(ops)

	WithFormat(ModeFile)(ops)

	require.Equal(t, FormatJSON, ops.format)
}

func TestNewOutputAppliesFormatToOutputTarget(t *testing.T) {
	tests := []struct {
		name           string
		mode           string
		format         string
		wantConsole    bool
		wantFileTarget bool
	}{
		{
			name: "console target with default json format",
			mode: ModeConsole,
		},
		{
			name:   "console target with json format",
			mode:   ModeConsole,
			format: FormatJSON,
		},
		{
			name:           "file target with default json format",
			mode:           ModeFile,
			wantFileTarget: true,
		},
		{
			name:           "file target with text format",
			mode:           ModeFile,
			format:         FormatText,
			wantConsole:    true,
			wantFileTarget: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := defaultOptions()
			WithMode(tt.mode)(ops)
			WithFormat(tt.format)(ops)

			output := newOutput(ops)
			consoleOutput, isConsole := output.(zerolog.ConsoleWriter)
			require.Equal(t, tt.wantConsole, isConsole)

			if tt.wantFileTarget {
				if isConsole {
					require.IsType(t, &lumberjack.Logger{}, consoleOutput.Out)
					require.True(t, consoleOutput.NoColor)
					return
				}
				require.IsType(t, &lumberjack.Logger{}, output)
			} else if isConsole {
				require.False(t, consoleOutput.NoColor)
			}
		})
	}
}

func TestTextOutputForFileDisablesANSIColor(t *testing.T) {
	ops := defaultOptions()
	WithMode(ModeFile)(ops)
	WithFormat(FormatText)(ops)

	var output bytes.Buffer
	writer := newTextOutput(&output, ops)
	_, err := writer.Write([]byte(`{"level":"info","message":"hello"}` + "\n"))

	require.NoError(t, err)
	require.NotContains(t, output.String(), "\x1b[")
}
