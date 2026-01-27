package main

import (
	cfg "github.com/conductorone/baton-opsgenie/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("opsgenie", cfg.Config)
}
