package tools

import "errors"

var (
	ErrNoLogsToFlush              = errors.New("no logs to flush")
	ErrLogFileAlreadyExists       = errors.New("log file already exists")
	ErrFailedToCreateLogFile      = errors.New("failed to create log file")
	ErrFailedToReconstructLogLine = errors.New("failed to reconstruct log line")
	ErrFailedToOpenLogFile        = errors.New("failed to open log file")
	ErrFailedToParseLogLine       = errors.New("failed to parse log line")
)
