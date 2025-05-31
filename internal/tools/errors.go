package tools

import "errors"

var (
	ErrNoLogsToFlush                 = errors.New("no logs to flush")
	ErrLogFileAlreadyExists          = errors.New("log file already exists")
	ErrFailedToCreateLogFile         = errors.New("failed to create log file")
	ErrFailedToReconstructLogLine    = errors.New("failed to reconstruct log line")
	ErrFailedToOpenLogFile           = errors.New("failed to open log file")
	ErrFailedToParseLogLine          = errors.New("failed to parse log line")
	ErrFailedToRenameLogFile         = errors.New("failed to rename log file")
	ErrFailedToCombineLogs           = errors.New("failed to combine logs")
	ErrFailedToCopyPreviousLogFile   = errors.New("failed to copy previous log file")
	ErrFailedToRemovePreviousLogFile = errors.New("failed to remove previous log file")
	ErrFailedToOpenPreviousLogFile   = errors.New("failed to open previous log file")
	ErrFailedToOpenNewLogFile        = errors.New("failed to open new log file")
	ErrFailedToGlobLogFiles          = errors.New("failed to glob log files")
	ErrFailedToRemoveLogFile         = errors.New("failed to remove log file")
	ErrFailedToCreateOutputFile      = errors.New("failed to create output file")
	ErrFailedToReadFile              = errors.New("failed to read file")
	ErrFailedToCopyTMPFile           = errors.New("failed to copy tmp log file")
)
