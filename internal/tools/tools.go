package tools

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"main/internal/logline"
	"main/internal/railway"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"
)

const (
	TMP_PATH = "./tmp"
)

func FlushLogsToFile(logs []*railway.EnvironmentLogsEnvironmentLogsLog, filename string) error {
	if len(logs) == 0 {
		return ErrNoLogsToFlush
	}

	// Create directory path if it doesn't exist
	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory path: %w", err)
	}

	logFile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToCreateLogFile, err)
	}

	defer logFile.Close()

	for _, logLine := range logs {
		logLineJson, err := logline.ReconstructLogLine(logLine)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToReconstructLogLine, err)
		}

		logFile.Write(logLineJson)
		logFile.Write([]byte("\n"))
	}

	return nil
}

func CombineLogFiles(logFilesLocation string, outputFilename string) error {
	files, err := filepath.Glob(filepath.Join(logFilesLocation, "*.jsonl"))
	if err != nil {
		return fmt.Errorf("failed to glob log files: %w", err)
	}

	slices.SortFunc(files, func(a, b string) int {
		aUnix, _ := strconv.ParseInt(filepath.Base(a), 10, 64)
		bUnix, _ := strconv.ParseInt(filepath.Base(b), 10, 64)
		return int(aUnix - bUnix)
	})

	outputFile, err := os.OpenFile(outputFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToCreateOutputFile, err)
	}

	defer outputFile.Close()

	for _, file := range files {
		f, err := os.OpenFile(file, os.O_RDONLY, 0644)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToReadFile, err)
		}

		defer f.Close()

		if _, err := io.Copy(outputFile, f); err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToCopyTMPFile, err)
		}

		f.Close()

		if err := os.Remove(file); err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToRemoveLogFile, err)
		}
	}

	return nil
}

type LogLine struct {
	Timestamp time.Time `json:"timestamp"`
}

func ReadFirstLineTimestamp(filename string) (time.Time, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
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

func FinalLogWrite(filename string, useResume bool) error {
	if useResume {
		if err := os.Rename(filename, ("previous_" + filename)); err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToRenameLogFile, err)
		}
	}

	if err := CombineLogFiles("./tmp", filename); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToCombineLogs, err)
	}

	if useResume {
		oldLogFile, err := os.OpenFile(("previous_" + filename), os.O_RDONLY, 0644)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToOpenPreviousLogFile, err)
		}

		defer oldLogFile.Close()

		newLogFile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToOpenNewLogFile, err)
		}

		defer newLogFile.Close()

		if _, err := io.Copy(newLogFile, oldLogFile); err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToCopyPreviousLogFile, err)
		}

		oldLogFile.Close()

		if err := os.Remove(("previous_" + filename)); err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToRemovePreviousLogFile, err)
		}
	}

	return nil
}

func ClearTempLogFiles() error {
	files, err := filepath.Glob(filepath.Join(TMP_PATH, "*.jsonl"))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToGlobLogFiles, err)
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToRemoveLogFile, err)
		}
	}

	return nil
}
