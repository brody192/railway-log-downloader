package railway

import "errors"

var (
	ErrFilterIDRequired       = errors.New("filter id is required")
	ErrDeploymentIdRequired   = errors.New("deployment id is required")
	ErrFailedToGetDeployment  = errors.New("failed to get deployment data")
	ErrFailedToGetLogs        = errors.New("failed to get logs")
	ErrNoLogsFound            = errors.New("no logs found")
	ErrFailedToParseTimestamp = errors.New("failed to parse timestamp")
)
