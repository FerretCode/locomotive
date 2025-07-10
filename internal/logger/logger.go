package logger

import (
	"os"
	"strconv"

	"log/slog"
)

var (
	StdoutLvl = slog.LevelVar{}

	stdoutHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: &StdoutLvl,
	})
	//enable source
	stdoutHandlerWithSource = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})

	stderrHandler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})
	// enable source
	stderrHandlerWithSource = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
	})

	// sends logs to stdout
	Stdout = slog.New(stdoutHandler)
	// sends logs to stdout with source info
	StdoutWithSource = slog.New(stdoutHandlerWithSource)

	// sends logs to stderr
	Stderr = slog.New(stderrHandler)
	// sends logs to stderr with source info
	StderrWithSource = slog.New(stderrHandlerWithSource)
)

func init() {
	StdoutLvl.Set(slog.LevelInfo)

	if b, _ := strconv.ParseBool(os.Getenv("DEBUG")); b {
		StdoutLvl.Set(slog.LevelDebug)
		Stdout.Debug("debug logging enabled")
	}
}
