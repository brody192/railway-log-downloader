package logline

import (
	"fmt"
	"strconv"

	"main/internal/railway"

	"github.com/buger/jsonparser"
)

var commonTimeStampAttributes = []string{"time", "_time", "timestamp", "ts", "datetime", "dt"}

// reconstruct a single log into a raw json object
func ReconstructLogLine(log *railway.EnvironmentLogsEnvironmentLogsLog) (jsonObject []byte, err error) {
	jsonObject = []byte("{}")

	// check for a timestamp attribute, fallback to the Railway timestamp if none found
	FoundTimeStampKey, FoundTimeStampValue, hasTimeStampAttr := attributesHasKeys(log.Attributes, commonTimeStampAttributes)

	if !hasTimeStampAttr {
		jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(log.Timestamp)), "timestamp")
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedToAppendToJSON, err)
		}
	}

	if hasTimeStampAttr {
		jsonObject, err = jsonparser.Set(jsonObject, []byte(FoundTimeStampValue), "timestamp")
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedToAppendToJSON, err)
		}
	}

	// append the level attribute to the object
	jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(log.Severity)), "level")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToAppendToJSON, err)
	}

	// remove ANSI escape codes from the message (colour codes, fonts, etc)
	cleanMessage := AnsiEscapeRe.ReplaceAllString(log.Message, "")

	jsonObject, err = jsonparser.Set(jsonObject, []byte(strconv.Quote(cleanMessage)), "message")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToAppendToJSON, err)
	}

	for i := range log.Attributes {
		// skip the level attribute since it was already added to the object above
		if log.Attributes[i].Key == "level" {
			continue
		}

		// skip the timestamp attribute since it was already added to the object above
		if hasTimeStampAttr && log.Attributes[i].Key == FoundTimeStampKey {
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
