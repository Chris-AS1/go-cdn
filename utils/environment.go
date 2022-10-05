package utils

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type settings struct {
	DeliveringPort          string
	DeliveringSubPath       string
	DeliveringSubPathEnable string
	EnableDeletion          string
	EnableInsertion         string
	DatabaseUsername        string
	DatabasePassword        string
	DatabasePort            string
	DatabaseURL             string
	DatabaseName            string
	DatabaseTableName       string
	DatabaseIDColumn        string
	DatabaseFilenameColumn  string
	DatabaseSSL             string
	RedisURL                string
}

var (
	EnvSettings = settings{
		DeliveringPort:          "3333",
		DeliveringSubPath:       "/image/",
		DeliveringSubPathEnable: "true",
		EnableDeletion:          "false",
		EnableInsertion:         "false",
		DatabaseUsername:        "",
		DatabasePassword:        "",
		DatabasePort:            "",
		DatabaseURL:             "",
		DatabaseName:            "",
		DatabaseTableName:       "",
		DatabaseIDColumn:        "",
		DatabaseFilenameColumn:  "",
		DatabaseSSL:             "disable",
		RedisURL:                "redis:6379",
	}
)

func parsePath(path string) string {
	path = strings.TrimLeft(path, "/")
	path = strings.TrimRight(path, "/")

	return path
}

func genDeliveringSubPath(env_var string) string {
	if env_var != "" {
		return parsePath(env_var)
	}

	return parsePath(EnvSettings.DeliveringSubPath)
}

func loadVar(env string, dest *string) {
	if env != "" {
		*dest = env
	}
}

func loadDotEnv(key string) string {
	env, err := godotenv.Read(".env")

	if err != nil {
		return ""
	}

	return env[key]
}

func LoadEnv() {
	// Loads .env
	loadVar(loadDotEnv("CDN_PORT"), &EnvSettings.DeliveringPort)
	loadVar(genDeliveringSubPath(loadDotEnv("CDN_SUBPATH")), &EnvSettings.DeliveringSubPath)
	loadVar(loadDotEnv("CDN_SUBPATH_ENABLE"), &EnvSettings.DeliveringSubPathEnable)
	loadVar(loadDotEnv("CDN_ENABLE_DELETE"), &EnvSettings.EnableDeletion)
	loadVar(loadDotEnv("CDN_ENABLE_INSERTION"), &EnvSettings.EnableInsertion)
	loadVar(loadDotEnv("DB_USERNAME"), &EnvSettings.DatabaseUsername)
	loadVar(loadDotEnv("DB_PASSWORD"), &EnvSettings.DatabasePassword)
	loadVar(loadDotEnv("DB_PORT"), &EnvSettings.DatabasePort)
	loadVar(loadDotEnv("DB_URL"), &EnvSettings.DatabaseURL)
	loadVar(loadDotEnv("DB_NAME"), &EnvSettings.DatabaseName)
	loadVar(loadDotEnv("DB_TABLE_NAME"), &EnvSettings.DatabaseTableName)
	loadVar(loadDotEnv("DB_COL_ID"), &EnvSettings.DatabaseIDColumn)
	loadVar(loadDotEnv("DB_COL_FN"), &EnvSettings.DatabaseFilenameColumn)
	loadVar(loadDotEnv("DB_SSL"), &EnvSettings.DatabaseSSL)
	loadVar(loadDotEnv("REDIS_URL"), &EnvSettings.RedisURL)

	// Loads environment
	loadVar(os.Getenv("CDN_PORT"), &EnvSettings.DeliveringPort)
	loadVar(genDeliveringSubPath(os.Getenv("CDN_SUBPATH")), &EnvSettings.DeliveringSubPath)
	loadVar(os.Getenv("CDN_SUBPATH_ENABLE"), &EnvSettings.DeliveringSubPathEnable)
	loadVar(os.Getenv("CDN_ENABLE_DELETE"), &EnvSettings.EnableDeletion)
	loadVar(os.Getenv("CDN_ENABLE_INSERTION"), &EnvSettings.EnableInsertion)
	loadVar(os.Getenv("DB_USERNAME"), &EnvSettings.DatabaseUsername)
	loadVar(os.Getenv("DB_PASSWORD"), &EnvSettings.DatabasePassword)
	loadVar(os.Getenv("DB_PORT"), &EnvSettings.DatabasePort)
	loadVar(os.Getenv("DB_URL"), &EnvSettings.DatabaseURL)
	loadVar(os.Getenv("DB_NAME"), &EnvSettings.DatabaseName)
	loadVar(os.Getenv("DB_TABLE_NAME"), &EnvSettings.DatabaseTableName)
	loadVar(os.Getenv("DB_COL_ID"), &EnvSettings.DatabaseIDColumn)
	loadVar(os.Getenv("DB_COL_FN"), &EnvSettings.DatabaseFilenameColumn)
	loadVar(os.Getenv("DB_SSL"), &EnvSettings.DatabaseSSL)
	loadVar(os.Getenv("REDIS_URL"), &EnvSettings.RedisURL)
}
