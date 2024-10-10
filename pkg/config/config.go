package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	ApiKeyField = field.StringField(
		"api-key",
		field.WithDescription("Opsgenie API Key"),
		field.WithRequired(true),
	)

	ConfigurationFields = []field.SchemaField{
		ApiKeyField,
	}

	ConfigurationSchema = field.Configuration{
		Fields: ConfigurationFields,
	}
)
