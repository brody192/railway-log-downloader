package railway

import (
	"context"
	"fmt"
	"time"
)

var (
	MAX_RETRY_COUNT = 5
	MAX_LOG_FETCH   = 5000 // 5000 is the maximum number of logs that the API will allow us to fetch
)

func GetAllDeploymentLogsBlocking(ctx context.Context, railwayClient *RailwayClient, logs chan<- LogLinesResponse, options GetLogsOptions) error {
	var attribute string
	var value string

	if options.ServiceId != "" {
		attribute = "service"
		value = options.ServiceId
	} else if options.DeploymentId != "" {
		attribute = "deployment"
		value = options.DeploymentId
	} else {
		return ErrFilterIDRequired
	}

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)

	if !options.ResumeFromTimestamp.IsZero() {
		timestamp = options.ResumeFromTimestamp.UTC().Format(time.RFC3339Nano)
	}

	environmentId := options.EnvironmentId

	if environmentId == "" {
		deployment, err := Deployment(ctx, railwayClient, options.DeploymentId)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToGetDeployment, err)
		}

		environmentId = deployment.Deployment.EnvironmentId
	}

	filter := buildFilter(attribute, value, options.Filter)

	logsToFetch := MAX_LOG_FETCH

	loopCount := 0
	errorCount := 0

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if options.ResumeFromTimestamp.IsZero() {
			switch loopCount {
			case 0:
				logsToFetch = logsToFetch - 1
			case 1:
				logsToFetch = logsToFetch + 1
			}
		}

		logsResponse, err := EnvironmentLogs(ctx, railwayClient,
			0,         // after limit
			timestamp, // anchor date
			time.Unix(0, 0).UTC().Format(time.RFC3339Nano), // before limit (Unix epoch)
			logsToFetch,
			environmentId, // environment id
			filter,        // filter
		)
		if err != nil {
			errorCount++

			// dead simple retry logic with static backoff
			if errorCount < MAX_RETRY_COUNT {
				time.Sleep(time.Second)

				continue
			}

			return fmt.Errorf("%w: %w", ErrFailedToGetLogs, err)
		}

		// reset the error count on a successful fetch
		errorCount = 0

		if len(logsResponse.EnvironmentLogs) == 0 {
			return ErrNoLogsFound
		}

		// we've reached the end of the logs
		if logsResponse.EnvironmentLogs[0].Timestamp == timestamp {
			break
		}

		// parse the first log timestamp
		firstLogTimestamp, err := time.Parse(time.RFC3339Nano, logsResponse.EnvironmentLogs[0].Timestamp)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToParseTimestamp, err)
		}

		if logsResponse.EnvironmentLogs[len(logsResponse.EnvironmentLogs)-1].Timestamp == timestamp {
			logs <- LogLinesResponse{
				Logs:               logsResponse.EnvironmentLogs[:len(logsResponse.EnvironmentLogs)-1],
				OldestLogTimestamp: firstLogTimestamp,
			}
		} else {
			logs <- LogLinesResponse{
				Logs:               logsResponse.EnvironmentLogs,
				OldestLogTimestamp: firstLogTimestamp,
			}
		}

		timestamp = logsResponse.EnvironmentLogs[0].Timestamp

		loopCount++
	}

	return nil
}

func GetAllDeploymentLogsAsync(ctx context.Context, railwayClient *RailwayClient, logs chan<- LogLinesResponse, options GetLogsOptions) {
	go func() {
		if err := GetAllDeploymentLogsBlocking(ctx, railwayClient, logs, options); err != nil {
			options.ErrorChannel <- err
			return
		}

		options.DoneChannel <- true
	}()
}

func buildFilter(attribute string, value string, filter string) string {
	filterString := fmt.Sprintf("@%s:%s", attribute, value)

	if filter != "" {
		filterString = fmt.Sprintf("%s %s", filterString, filter)
	}

	return filterString
}
