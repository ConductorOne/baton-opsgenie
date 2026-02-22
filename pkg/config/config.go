package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	ApiKeyField = field.StringField(
		"api-key",
		field.WithDescription("Opsgenie API Key"),
		field.WithIsSecret(true),
		field.WithRequired(true),
	)

	BaseURLField = field.StringField(
		"base-url",
		field.WithDescription("Override the Opsgenie API URL (for testing)"),
		field.WithHidden(true),
		field.WithExportTarget(field.ExportTargetCLIOnly),
	)

	ConfigurationFields = []field.SchemaField{
		ApiKeyField,
		BaseURLField,
	}

	ConfigurationSchema = field.Configuration{
		Fields: ConfigurationFields,
	}
)

//go:generate go run ./gen
var Config = field.NewConfiguration(ConfigurationFields)
