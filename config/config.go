package config

import (
	"gorm.io/gorm"
	"html/template"
)

// AppConfig holds the application config.
type AppConfig struct {
	DB            *gorm.DB
	Env           *EnvVariables
	UseCache      bool
	TemplateCache map[string]*template.Template
}

// EnvVariables holds environment variables used in the application.
type EnvVariables struct {
	PostgresHost   string
	PostgresUser   string
	PostgresPass   string
	PostgresDBName string
	JWTSecret      string
}
