package railway

import (
	"context"
	"fmt"
	"main/internal/config"
	"time"
)

func GetAllDeploymentLogsBlocking(ctx context.Context, railwayClient *RailwayClient, logs *[]*EnvironmentLogsEnvironmentLogsLog, options GetLogsOptions) error {
	if options.DeploymentId == "" {
		return ErrDeploymentIdRequired
	}

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)

	deployment, err := Deployment(ctx, railwayClient, options.DeploymentId)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToGetDeployment, err)
	}

	if options.ProgressChannel != nil {
		options.ProgressChannel <- ProgressInfo{
			DownloadedLogs: 0,
			CurrentDate:    time.Now().UTC(),
		}
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		logsResponse, err := EnvironmentLogs(ctx, railwayClient,
			0,         // after limit
			timestamp, // anchor date
			time.Unix(0, 0).UTC().Format(time.RFC3339Nano), // before limit (Unix epoch)
			// 5000 is the maximum number of logs that the API will allow us to fetch
			// Why 4999? Becuase the API will return 5001 logs even if we set the limit to 5000, so this is compensating for that to get the even count up in the progress indicator
			4999,
			deployment.Deployment.EnvironmentId, // environment id
			buildFilter(config.Railway.DeploymentID, config.Railway.Filter), // deployment id
		)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToGetLogs, err)
		}

		if len(logsResponse.EnvironmentLogs) == 0 {
			return ErrNoLogsFound
		}

		if logsResponse.EnvironmentLogs[0].Timestamp == timestamp {
			break
		}

		// Prepend new logs to the beginning since we're going backwards in time
		*logs = append(logsResponse.EnvironmentLogs, *logs...)

		timestamp = logsResponse.EnvironmentLogs[0].Timestamp

		if options.ProgressChannel != nil {
			currentDate, _ := time.Parse(time.RFC3339Nano, timestamp)
			options.ProgressChannel <- ProgressInfo{
				DownloadedLogs: int64(len(*logs)),
				CurrentDate:    currentDate,
			}
		}
	}

	return nil
}

func GetAllDeploymentLogsAsync(ctx context.Context, railwayClient *RailwayClient, logs *[]*EnvironmentLogsEnvironmentLogsLog, options GetLogsOptions) {
	go func() {
		if err := GetAllDeploymentLogsBlocking(ctx, railwayClient, logs, options); err != nil {
			options.ErrorChannel <- err
			return
		}

		options.DoneChannel <- true
	}()
}

func buildFilter(deploymentId string, filter string) string {
	filterString := "@deployment:" + deploymentId

	if filter != "" {
		filterString = fmt.Sprintf("%s %s", filterString, filter)
	}

	return filterString
}
