package logline

import "main/internal/railway"

// searches for the given key and returns the corresponding value (and true) if found, or an empty string (and false)
func attributesHasKeys(attributes []*railway.EnvironmentLogsEnvironmentLogsLogAttributesLogAttribute, keys []string) (string, string, bool) {
	for i := range attributes {
		for j := range keys {
			if keys[j] == attributes[i].Key {
				return attributes[i].Key, attributes[i].Value, true
			}
		}
	}

	return "", "", false
}
