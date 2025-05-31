package railway

import "time"

type LogLinesResponse struct {
	Logs               []*EnvironmentLogsEnvironmentLogsLog
	OldestLogTimestamp time.Time
}

type GetLogsOptions struct {
	ResumeFromTimestamp time.Time

	DeploymentId  string
	EnvironmentId string
	ServiceId     string

	Filter string

	ErrorChannel chan error // Not used in blocking mode
	DoneChannel  chan bool  // Not used in blocking mode
}
