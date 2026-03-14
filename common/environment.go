package common

import (
	"os"
	"strings"
)

type Environment struct {
	HttpPort      int
	BaseUrl       string
	SiteName      string
	MaxPostLength int
}

const (
	HTTP_PORT       string = "HTTP_PORT"
	LOG_LEVEL       string = "LOG_LEVEL"
	BASE_URL        string = "BASE_URL"
	SITE_NAME       string = "SITE_NAME"
	MAX_POST_LENGTH string = "MAX_POST_LENGTH"
)

func NewEnvironment() *Environment {
	var baseUrl = GetEnvString(BASE_URL, "https://pincer.wtf")
	if !strings.HasSuffix(baseUrl, "/") {
		baseUrl = baseUrl + "/"
	}

	return &Environment{
		HttpPort:      GetEnvInt(HTTP_PORT, 8001),
		BaseUrl:       baseUrl,
		SiteName:      GetEnvString(SITE_NAME, "Pincer"),
		MaxPostLength: GetEnvInt(MAX_POST_LENGTH, 500),
	}
}

func GetEnvInt(variable string, def int) int {
	val := os.Getenv(variable)

	if val == "" {
		return def
	}

	return TryParseInt(variable, def)
}

func GetEnvString(variable string, def string) string {
	val := os.Getenv(variable)
	if val == "" {
		return def
	}

	return val
}

// Returns an environment variable as a boolean
// true will be returned if it matches "True" ignoring case
// false will be returned if it matches "False" ignoring case
// otherwise the default value will be returned
func GetEnvBool(variable string, def bool) bool {
	val := os.Getenv(variable)
	val = strings.ToLower(strings.TrimSpace(val))
	switch val {
	case "true":
		return true
	case "false":
		return false
	default:
		return def
	}
}
