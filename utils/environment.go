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
	DatabaseUsername        string
	DatabasePassword        string
	DatabasePort            string
	DatabaseURL             string
	DatabaseName            string
	DatabaseSSL             string
}

var (
	EnvSettings = settings{
		DeliveringPort:          "3333",
		DeliveringSubPath:       "/image/",
		DeliveringSubPathEnable: "true",
		DatabaseUsername:        "",
		DatabasePassword:        "",
		DatabasePort:            "",
		DatabaseURL:             "",
		DatabaseName:            "",
		DatabaseSSL:             "disabled",
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
	loadVar(loadDotEnv("CDN_PORT"), &EnvSettings.DeliveringPort)
	loadVar(genDeliveringSubPath(loadDotEnv("CDN_SUBPATH")), &EnvSettings.DeliveringSubPath)
	loadVar(loadDotEnv("CDN_SUBPATH_ENABLE"), &EnvSettings.DeliveringSubPathEnable)
	loadVar(loadDotEnv("DB_USERNAME"), &EnvSettings.DatabaseUsername)
	loadVar(loadDotEnv("DB_PASSWORD"), &EnvSettings.DatabasePassword)
	loadVar(loadDotEnv("DB_PORT"), &EnvSettings.DatabasePort)
	loadVar(loadDotEnv("DB_URL"), &EnvSettings.DatabaseURL)
	loadVar(loadDotEnv("DB_NAME"), &EnvSettings.DatabaseName)
	loadVar(loadDotEnv("DB_SSL"), &EnvSettings.DatabaseSSL)

	loadVar(os.Getenv("CDN_PORT"), &EnvSettings.DeliveringPort)
	loadVar(genDeliveringSubPath(os.Getenv("CDN_SUBPATH")), &EnvSettings.DeliveringSubPath)
	loadVar(os.Getenv("CDN_SUBPATH_ENABLE"), &EnvSettings.DeliveringSubPathEnable)
	loadVar(os.Getenv("DB_USERNAME"), &EnvSettings.DatabaseUsername)
	loadVar(os.Getenv("DB_PASSWORD"), &EnvSettings.DatabasePassword)
	loadVar(os.Getenv("DB_PORT"), &EnvSettings.DatabasePort)
	loadVar(os.Getenv("DB_URL"), &EnvSettings.DatabaseURL)
	loadVar(os.Getenv("DB_NAME"), &EnvSettings.DatabaseName)
	loadVar(os.Getenv("DB_SSL"), &EnvSettings.DatabaseSSL)
}
