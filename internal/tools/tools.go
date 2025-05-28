package tools

import (
	"fmt"
	"main/internal/logline"
	"main/internal/railway"
	"os"
	"strconv"
)

func FlushLogsToFile(logs []*railway.EnvironmentLogsEnvironmentLogsLog, filename string) error {
	if len(logs) == 0 {
		return ErrNoLogsToFlush
	}

	logFile, err := os.Create(filename)
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

func MustParseBool(value string) bool {
	boolValue, _ := strconv.ParseBool(value)

	return boolValue
}
