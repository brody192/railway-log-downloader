package tools

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"main/internal/config"
	"main/internal/logline"
	"main/internal/railway"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func FlushLogsToFile(logs []*railway.EnvironmentLogsEnvironmentLogsLog, filename string) error {
	if len(logs) == 0 {
		return ErrNoLogsToFlush
	}

	logFile := &os.File{}
	err := error(nil)
	tempFile := &os.File{}

	if MustParseBool(config.Railway.Resume) {
		logFile, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
		if err == nil {
			// Create temp file to store existing content
			tempFile, err = os.CreateTemp(filepath.Dir(filename), "railway-logs-*.tmp")
			if err == nil {
				// Copy existing content to temp file
				logFile.Seek(0, 0)
				io.Copy(tempFile, logFile)
				// Reset to beginning for writing new logs
				logFile.Seek(0, 0)
				logFile.Truncate(0) // Clear the file
			}
		}
	} else {
		logFile, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}

	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToCreateLogFile, err)
	}

	defer logFile.Close()

	if tempFile != nil {
		defer func() {
			tempFile.Close()
			os.Remove(tempFile.Name()) // Clean up temp file
		}()
	}

	for _, logLine := range logs {
		logLineJson, err := logline.ReconstructLogLine(logLine)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToReconstructLogLine, err)
		}

		logFile.Write(logLineJson)
		logFile.Write([]byte("\n"))
	}

	// When resuming, copy the old content back from temp file
	if tempFile != nil {
		tempFile.Seek(0, 0) // Reset temp file to beginning
		io.Copy(logFile, tempFile)
	}

	return nil
}

func MustParseBool(value string) bool {
	boolValue, _ := strconv.ParseBool(value)

	return boolValue
}

type LogLine struct {
	Timestamp time.Time `json:"timestamp"`
}

func ReadFirstLineTimestamp(filename string) (time.Time, error) {
	file, err := os.Open(filename)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %w", ErrFailedToOpenLogFile, err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	if scanner.Scan() {
		line := scanner.Bytes()

		logLine := LogLine{}

		if err := json.Unmarshal(line, &logLine); err != nil {
			return time.Time{}, fmt.Errorf("%w: %w", ErrFailedToParseLogLine, err)
		}

		return logLine.Timestamp, nil
	}

	return time.Time{}, nil
}
