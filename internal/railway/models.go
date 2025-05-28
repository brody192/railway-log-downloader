package railway

import "time"

type ProgressInfo struct {
	DownloadedLogs int64
	CurrentDate    time.Time
}

type GetLogsOptions struct {
	DeploymentId string

	ProgressChannel chan ProgressInfo

	ErrorChannel chan error // Not used in blocking mode
	DoneChannel  chan bool  // Not used in blocking mode
}
