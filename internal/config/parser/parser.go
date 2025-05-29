package parser

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize/english"
	"github.com/google/uuid"
)

// ParseFlags parses the flags and returns a map of the flags and their values
func ParseFlags(cfg any, t ...reflect.Type) map[string]*string {
	flags := make(map[string]*string)

	if len(t) == 0 {
		v := reflect.ValueOf(cfg).Elem()
		t = []reflect.Type{v.Type()}
	}

	// Register flags
	for i := range t[0].NumField() {
		field := t[0].Field(i)
		if flagName := field.Tag.Get("flag"); flagName != "" {
			flags[field.Name] = flag.String(flagName, "", field.Tag.Get("usage"))
		}
	}

	flag.Parse()

	return flags
}

// ParseConfig parses the config and returns all errors that occurred while parsing the config
func ParseConfig(cfg any) []error {
	var errors []error

	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	flags := ParseFlags(cfg, t)

	// Set values and collect errors
	for i := range t.NumField() {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip if field has no tags at all
		if field.Tag == "" {
			continue
		}

		// Check flag value first
		if flagVal, ok := flags[field.Name]; ok && *flagVal != "" {
			fieldValue.SetString(*flagVal)
			continue // Skip env vars if flag is set
		}

		// If no flag value, check environment variables
		for _, env := range strings.Split(field.Tag.Get("env"), ",") {
			if val := os.Getenv(env); val != "" {
				fieldValue.SetString(val)
				break
			}
		}

		// If no value set yet, use default value if present
		if fieldValue.String() == "" {
			if _, exists := field.Tag.Lookup("default"); exists {
				fieldValue.SetString(field.Tag.Get("default"))
				continue
			}
		}

		// collect errors only if field is required and no value is set and no default exists
		requiredTag := field.Tag.Get("required")
		requiredOneOfTag := field.Tag.Get("required_one_of")
		requiredAllTag := field.Tag.Get("required_all")
		isRequired, _ := strconv.ParseBool(requiredTag)

		// Skip individual required validation if this field belongs to a required_one_of group
		if requiredOneOfTag == "" && requiredAllTag == "" && fieldValue.String() == "" && isRequired {
			flagName := field.Tag.Get("flag")
			envVars := strings.Split(field.Tag.Get("env"), ",")

			if flagName != "" {
				errors = append(errors, fmt.Errorf("%s is required, set: %s in the environment or use the --%s flag", field.Name, english.WordSeries(envVars, "or"), flagName))
			} else {
				errors = append(errors, fmt.Errorf("%s is required, set: %s in the environment", field.Name, english.WordSeries(envVars, "or")))
			}
		}
	}

	// Validate required groups
	groups := make(map[string][]fieldInfo)
	requiredAllGroups := make(map[string][]fieldInfo)

	for i := range t.NumField() {
		field := t.Field(i)
		fieldValue := v.Field(i)
		requiredOneOfTag := field.Tag.Get("required_one_of")
		requiredAllTag := field.Tag.Get("required_all")

		if requiredOneOfTag != "" {
			groupName := requiredOneOfTag
			groups[groupName] = append(groups[groupName], fieldInfo{
				Name:     field.Name,
				HasValue: fieldValue.String() != "",
				FlagName: field.Tag.Get("flag"),
				EnvVars:  strings.Split(field.Tag.Get("env"), ","),
			})
		}

		if requiredAllTag != "" {
			groupName := requiredAllTag
			requiredAllGroups[groupName] = append(requiredAllGroups[groupName], fieldInfo{
				Name:     field.Name,
				HasValue: fieldValue.String() != "",
				FlagName: field.Tag.Get("flag"),
				EnvVars:  strings.Split(field.Tag.Get("env"), ","),
			})
		}
	}

	// Validate each required group
	for _, fields := range groups {
		fieldsWithValues := 0
		var fieldNames []string

		for _, field := range fields {
			fieldNames = append(fieldNames, field.Name)

			if field.HasValue {
				fieldsWithValues++
			}
		}

		if fieldsWithValues == 0 {
			// No fields in the group have values
			var options []string
			for _, field := range fields {
				if field.FlagName != "" {
					options = append(options, fmt.Sprintf("--%s flag", field.FlagName))
				}
				for _, env := range field.EnvVars {
					if env != "" {
						options = append(options, fmt.Sprintf("%s environment variable", env))
					}
				}
			}
			errors = append(errors, fmt.Errorf("Exactly one of %s is required, provide one of: %s", english.WordSeries(fieldNames, "or"), english.WordSeries(options, "or")))
		} else if fieldsWithValues > 1 {
			// Multiple fields in the group have values
			errors = append(errors, fmt.Errorf("Only one of %s can be provided, but multiple were set", english.WordSeries(fieldNames, "or")))
		}
	}

	// Validate required_all groups
	for _, fields := range requiredAllGroups {
		fieldsWithValues := 0
		var fieldNames []string
		var fieldsWithoutValues []string

		for _, field := range fields {
			fieldNames = append(fieldNames, field.Name)

			if field.HasValue {
				fieldsWithValues++
			} else {
				fieldsWithoutValues = append(fieldsWithoutValues, field.Name)
			}
		}

		// If some fields have values but not all, it's an error
		if fieldsWithValues > 0 && fieldsWithValues < len(fields) {
			var missingOptions []string
			for _, field := range fields {
				if !field.HasValue {
					if field.FlagName != "" {
						missingOptions = append(missingOptions, fmt.Sprintf("--%s flag", field.FlagName))
					}
					for _, env := range field.EnvVars {
						if env != "" {
							missingOptions = append(missingOptions, fmt.Sprintf("%s environment variable", env))
						}
					}
				}
			}
			errors = append(errors, fmt.Errorf("When any of %s is provided, all other mutually exclusive options must be provided. Missing: %s", english.WordSeries(fieldNames, "or"), english.WordSeries(missingOptions, "or")))
		}
	}

	// Final type validation pass
	for i := range t.NumField() {
		field := t.Field(i)
		fieldValue := v.Field(i)
		fieldValueStr := fieldValue.String()

		if validate := field.Tag.Get("validate"); validate != "" && fieldValueStr != "" {
			switch validate {
			case "uuid":
				if _, err := uuid.Parse(fieldValueStr); err != nil {
					errors = append(errors, fmt.Errorf("%s: %s is not a valid %s", field.Name, fieldValueStr, validate))
					continue
				}
			case "boolean":
				if _, err := strconv.ParseBool(fieldValueStr); err != nil {
					errors = append(errors, fmt.Errorf("%s: %s is not a valid boolean", field.Name, fieldValueStr))
					continue
				}
			default:
				errors = append(errors, fmt.Errorf("%s: validate for type %s not implemented", field.Name, validate))
				continue
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// GetRequiredGroupValue returns the value, and flag name for the field that is set in the specified required group
func GetRequiredGroupValue(cfg any, groupName string) (flagName, value string) {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	for i := range t.NumField() {
		field := t.Field(i)
		fieldValue := v.Field(i)
		requiredOneOfTag := field.Tag.Get("required_one_of")

		if requiredOneOfTag == groupName && fieldValue.String() != "" {
			return field.Tag.Get("flag"), fieldValue.String()
		}
	}

	return "", ""
}

// fieldInfo holds information about a field for group validation
type fieldInfo struct {
	Name     string
	HasValue bool
	FlagName string
	EnvVars  []string
}
