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

func main() {
	// Create the railway client
	railwayClient := railway.NewAuthedClient(config.Railway.AccountToken)

	flagName, value := config.Railway.GetRequiredGroupValue("service_or_deployment")

	// Create the log file name
	logFileName := fmt.Sprintf("%s-%s.jsonl", flagName, value)

	// Check if the log file already exists to avoid overwriting
	if _, err := os.Stat(logFileName); err == nil && !tools.MustParseBool(config.Railway.OverwriteFile) && !tools.MustParseBool(config.Railway.Resume) {
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
	progressChannel := make(chan railway.ProgressInfo)

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Create the log lines slice to store the logs
	logLines := []*railway.EnvironmentLogsEnvironmentLogsLog{}

	// Create the resume from timestamp
	resumeFromTimestamp := time.Time{}

	// If the resume flag is set, read the last downloaded log timestamp
	if tools.MustParseBool(config.Railway.Resume) {
		lastDownloadedLogTimestamp, err := tools.ReadFirstLineTimestamp(logFileName)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
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

	// Start the progress channel goroutine
	go func() {
		for progress := range progressChannel {
			logDownloadSpinner.Suffix = fmt.Sprintf(" %s Logs - Position: %s",
				humanize.Comma(progress.DownloadedLogs),
				progress.CurrentDate.UTC().Format("January 2, 2006 15:04:05 MST"),
			)
		}
	}()

	// Start the log collection goroutine
	railway.GetAllDeploymentLogsAsync(ctx, railwayClient, &logLines, railway.GetLogsOptions{
		ResumeFromTimestamp: resumeFromTimestamp,
		DeploymentId:        config.Railway.DeploymentID,
		ProgressChannel:     progressChannel,
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
	if len(logLines) == 0 {
		fmt.Println("No logs collected, exiting...")
		os.Exit(0)
	}

	// Create the flush logs spinner
	flushLogsSpinner := spinner.New(spinner.CharSets[11], (100 * time.Millisecond))
	flushLogsSpinner.Suffix = " Flushing logs"
	flushLogsSpinner.Reverse()

	// Start the flush logs spinner if there are more than 50,000 logs
	// Why? Because any log line amount over 50,000 will create a noticeable delay in the flushing process
	if len(logLines) >= 50_000 {
		flushLogsSpinner.Start()
	}

	// Flush logs to file before exiting
	if err := tools.FlushLogsToFile(logLines, logFileName); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	// Stop the flush logs spinner
	flushLogsSpinner.Stop()

	// Print the completion message
	fmt.Printf("Flushed %s logs to file: %s\n", humanize.Comma(int64(len(logLines))), logFileName)
}
