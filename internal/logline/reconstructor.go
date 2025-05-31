package logline

import (
	"fmt"
	"strconv"
	"strings"

	"main/internal/railway"

	"github.com/buger/jsonparser"
)

// reconstruct a single log into a raw json object
func ReconstructLogLine(log *railway.EnvironmentLogsEnvironmentLogsLog) (jsonObject []byte, err error) {
	jsonObject = []byte("{}")

	jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(log.Timestamp)), "timestamp")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToAppendToJSON, err)
	}

	// append the level attribute to the object
	jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(log.Severity)), "level")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToAppendToJSON, err)
	}

	// remove ANSI escape codes from the message (colour codes, fonts, etc)
	cleanMessage := log.Message

	cleanMessage = AnsiEscapeRe.ReplaceAllString(cleanMessage, "")
	cleanMessage = strings.TrimSpace(cleanMessage)

	jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(cleanMessage)), "message")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToAppendToJSON, err)
	}

	for i := range log.Attributes {
		// skip the level attribute since it was already added to the object above
		if log.Attributes[i].Key == "level" {
			continue
		}

		// append the attribute to the object
		jsonObject, err = jsonparser.Set(jsonObject, []byte(log.Attributes[i].Value), log.Attributes[i].Key)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedToAppendToJSON, err)
		}
	}

	return jsonObject, nil
}
