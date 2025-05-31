package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"main/internal/config"
	"main/internal/railway"
	"main/internal/tools"

	"github.com/briandowns/spinner"
	"github.com/dustin/go-humanize"
)

func init() {
	if err := tools.ClearTempLogFiles(); err != nil {
		fmt.Printf("Error clearing temp log files: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	// Create the railway client
	railwayClient := railway.NewAuthedClient(config.Railway.AccountToken.String())

	flagName, value := config.Railway.GetRequiredGroupValue("service_or_deployment")

	// Create the log file name
	logFileName := fmt.Sprintf("%s-%s.jsonl", flagName, value)

	// If the log file does not exist and the resume flag is provided, exit
	if _, err := os.Stat(logFileName); err != nil && config.Railway.Resume.Bool() {
		fmt.Println("Could not find a log file to resume from but the --resume flag was provided")
		os.Exit(1)
	}

	// If the log file does not exist and the overwrite flag is provided, exit
	if _, err := os.Stat(logFileName); err != nil && config.Railway.OverwriteFile.Bool() {
		fmt.Println("Could not find a log file to resume from but the --overwrite flag was provided")
		os.Exit(1)
	}

	// Check if the log file already exists to avoid overwriting
	if _, err := os.Stat(logFileName); err == nil && config.Railway.OverwriteFile.Bool() && config.Railway.Resume.Bool() {
		fmt.Printf("Log file %s already exists, delete or remove it to continue\n", logFileName)
		fmt.Println("If you want to resume downloading logs from the oldest downloaded log, use the --resume flag")
		fmt.Println("If you want to overwrite the existing log file, use the --overwrite flag")
		os.Exit(1)
	}

	// Set up signal handling for Ctrl / Cmd + C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create the channels to communicate with the goroutines
	doneChannel := make(chan bool)
	errorChannel := make(chan error)
	logLinesChannel := make(chan railway.LogLinesResponse)

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Create the resume from timestamp
	resumeFromTimestamp := time.Time{}

	// If the resume flag is set, read the last downloaded log timestamp
	if config.Railway.Resume.Bool() {
		lastDownloadedLogTimestamp, err := tools.ReadFirstLineTimestamp(logFileName)
		if err != nil {
			fmt.Printf("Error reading first line timestamp: %s\n", err)
			os.Exit(1)
		}

		resumeFromTimestamp = lastDownloadedLogTimestamp

		fmt.Printf("Resuming from %s\n", resumeFromTimestamp.UTC().Format("January 2, 2006 15:04:05 MST"))
	}

	// Create the spinner
	logDownloadSpinner := spinner.New(spinner.CharSets[11], (100 * time.Millisecond))
	logDownloadSpinner.Suffix = " 0 Logs"
	logDownloadSpinner.Reverse()
	logDownloadSpinner.Start()

	// Initialize the variable to track the number of logs downloaded
	downloadedLogs := int64(0)

	go func() {
		for logLines := range logLinesChannel {
			if err := tools.FlushLogsToFile(logLines.Logs, fmt.Sprintf("./tmp/%d.jsonl", logLines.OldestLogTimestamp.UTC().UnixMilli())); err != nil {
				errorChannel <- err
				return
			}

			downloadedLogs += int64(len(logLines.Logs))

			logDownloadSpinner.Suffix = fmt.Sprintf(" %s Logs - Position: %s",
				humanize.Comma(downloadedLogs),
				logLines.OldestLogTimestamp.UTC().Format("January 2, 2006 15:04:05 MST"),
			)
		}
	}()

	// Start the log collection goroutine
	railway.GetAllDeploymentLogsAsync(ctx, railwayClient, logLinesChannel, railway.GetLogsOptions{
		ResumeFromTimestamp: resumeFromTimestamp,
		DeploymentId:        config.Railway.DeploymentID.String(),
		EnvironmentId:       config.Railway.EnvironmentID.String(),
		ServiceId:           config.Railway.ServiceID.String(),
		Filter:              config.Railway.Filter.String(),
		ErrorChannel:        errorChannel,
		DoneChannel:         doneChannel,
	})

	// Print the start message
	fmt.Println("Collecting logs in the background... Press Ctrl / Cmd + C to stop and save logs")

	// Wait for either Ctrl+C or background goroutine to finish
	select {
	case <-sigChan:
		logDownloadSpinner.Stop()

		fmt.Println("Received interrupt signal, stopping...")

		cancel() // Cancel the context to stop the goroutine
	case <-doneChannel:
		logDownloadSpinner.Stop()

		fmt.Println("Log collection completed")
	case err := <-errorChannel:
		logDownloadSpinner.Stop()

		fmt.Printf("Error: %s\n", strings.TrimSpace(err.Error()))
	}

	// If no logs were collected, exit
	if downloadedLogs == 0 {
		fmt.Println("No logs collected, exiting...")
		os.Exit(0)
	}

	// Create the flush logs spinner
	flushLogsSpinner := spinner.New(spinner.CharSets[11], (100 * time.Millisecond))
	flushLogsSpinner.Suffix = " Flushing logs"
	flushLogsSpinner.Reverse()

	// Start the flush logs spinner if there are more than 3,000,000 logs
	// Why? Because any log line amount over 3,000,000 will create a noticeable delay in the flushing process
	if downloadedLogs >= 3_000_000 {
		flushLogsSpinner.Start()
	}

	// Flush logs to file before exiting
	// This handles the reconstruction of the multiple *.jsonl files into a single log file
	// if `useResume` is true, it will prepend the newly downloaded logs to the existing log file
	if err := tools.FinalLogWrite(logFileName, config.Railway.Resume.Bool()); err != nil {
		fmt.Printf("Error saving logs: %s\n", err)
		os.Exit(1)
	}

	// Stop the flush logs spinner
	// no-op if the spinner was not started
	flushLogsSpinner.Stop()

	// Print the completion message
	if config.Railway.Resume.Bool() {
		fmt.Printf("Flushed an additional %s logs to file: %s\n", humanize.Comma(downloadedLogs), logFileName)
	} else {
		fmt.Printf("Flushed %s logs to file: %s\n", humanize.Comma(downloadedLogs), logFileName)
	}
}
