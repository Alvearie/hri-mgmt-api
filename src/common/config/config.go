/**
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffyaml"
	"net/url"
	"strings"
)

// Config Final config struct returned to be passed around
type Config struct {
	ConfigPath         string
	OidcIssuer         string
	JwtAudienceId      string
	Validation         bool
	AuthDisabled       bool
	ElasticUrl         string
	ElasticUsername    string
	ElasticPassword    string
	ElasticCert        string
	ElasticServiceCrn  string
	KafkaAdminUrl      string // required for IBM Event Streams to manage topics
	KafkaBrokers       StringSlice
	KafkaProperties    StringMap // valid properties: https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md
	LogLevel           string
	NewRelicEnabled    bool
	NewRelicAppName    string
	NewRelicLicenseKey string
	TlsEnabled         bool
	TlsCertPath        string
	TlsKeyPath         string
}

// StringSlice is a flag.Value that collects each Set string into a slice, allowing for repeated flags.
type StringSlice []string

// Set implements flag.Value and appends the string to the slice.
func (ss *StringSlice) Set(s string) error {
	*ss = append(*ss, strings.Split(s, ",")...)
	return nil
}

// Return an empty string. This method is required for the flag.Value interface, but is not needed for the hri-mgmt-api.
func (ss *StringSlice) String() string {
	return ""
}

// StringMap is a flag.Value that collects each Set(string) into a map
// using ':' as the key:value separator and allowing for repeated flags.
type StringMap map[string]string

// Set implements flag.Value and appends the string to the slice.
func (sm *StringMap) Set(s string) error {
	if *sm == nil {
		*sm = make(StringMap)
	}

	entries := strings.Split(s, ",")
	for i := range entries {
		tokens := strings.Split(entries[i], ":")
		if len(tokens) != 2 {
			return errors.New("invalid StringMap entry '" + entries[i] + "'; it must contain exactly one ':' to separate the key from the value")
		}
		(*sm)[tokens[0]] = tokens[1]
	}
	return nil
}

// This method is required for the flag.Value interface, but is not needed for the hri-mgmt-api.
func (sm *StringMap) String() string {
	return fmt.Sprint(*sm)
}

// ValidateConfig Perform verification on the finalized config.  Return an error if validation failed.
func ValidateConfig(config Config) error {
	if len(config.ConfigPath) == 0 {
		return errors.New("no config file supplied")
	}

	errorBuilder := strings.Builder{}
	errorHeader := "Configuration errors:"
	errorBuilder.WriteString(errorHeader)

	// Make sure OidcIssuer in the form of a valid URL
	if !config.AuthDisabled && !isValidUrl(config.OidcIssuer) {
		errorBuilder.WriteString("\n\tOIDC Issuer is an invalid URL:  " + config.OidcIssuer)
	}
	if config.ElasticUrl == "" {
		errorBuilder.WriteString("\n\tAn Elasticsearch base URL was not specified")
	}
	if config.ElasticUsername == "" {
		errorBuilder.WriteString("\n\tAn Elasticsearch username was not specified")
	}
	if config.ElasticPassword == "" {
		errorBuilder.WriteString("\n\tAn Elasticsearch password was not specified")
	}
	if config.ElasticCert == "" {
		errorBuilder.WriteString("\n\tAn Elasticsearch certificate was not specified")
	} else {
		// Ensure the base64-encoded cert decodes without errors
		cert, _ := pem.Decode([]byte(config.ElasticCert))
		if cert == nil {
			errorBuilder.WriteString("\n\tThe Elasticsearch certificate is invalid")
		}
	}
	if config.ElasticServiceCrn == "" {
		errorBuilder.WriteString("\n\tAn Elasticsearch service CRN was not specified")
	}
	if config.KafkaAdminUrl == "" {
		errorBuilder.WriteString("\n\tThe Kafka administration url was not specified")
	}
	if len(config.KafkaBrokers) == 0 {
		errorBuilder.WriteString("\n\tNo Kafka brokers were defined")
	}
	if config.NewRelicEnabled && config.NewRelicAppName == "" {
		errorBuilder.WriteString("\n\tNew Relic monitoring enabled, but the New Relic app name was not specified")
	}
	if config.NewRelicEnabled && config.NewRelicLicenseKey == "" {
		errorBuilder.WriteString("\n\tNew Relic monitoring enabled, but the New Relic license key was not specified")
	}
	if config.TlsEnabled {
		if config.TlsCertPath == "" {
			errorBuilder.WriteString("\n\tTLS is enabled but a path to a TLS certificate for the server was not specified")
		}
		if config.TlsKeyPath == "" {
			errorBuilder.WriteString("\n\tTLS is enabled but a path to a TLS key for the server was not specified")
		}
	}

	errorMsg := errorBuilder.String()
	if len(errorMsg) > len(errorHeader) {
		return errors.New(errorMsg)
	}
	return nil
}

// GetConfig generates a Config struct containing config values from command line flags,
// environment variables, and a config file, in that order.
// ./config.yml is used by default, but you can change that by using the --config-path flag.
// For example, in your command line, you would type:
// [executable] --config-path ./alternateConfig.yml
func GetConfig(configPath string, commandLineFlags []string) (Config, error) {
	fs := flag.NewFlagSet("mgmtApiConfig", flag.ContinueOnError)
	// Define all possible config values via flag declarations.
	// In runtime, flag names are capitalized, and separator characters are converted to underscores.
	config := Config{}
	fs.StringVar(&config.ConfigPath, "config-path", configPath, "(Optional) Path of an alternate config file")
	fs.BoolVar(&config.AuthDisabled, "auth-disabled", false, "(Optional) True to disable Authorization using OAuth")
	fs.StringVar(&config.OidcIssuer, "oidc-issuer", "", "(Optional) The base URL of the OIDC issuer to use for OAuth authentication (e.g. https://us-south.appid.cloud.ibm.com/oauth/v4/<tenantId>)")
	fs.StringVar(&config.JwtAudienceId, "jwt-audience-id", "", "(Optional) The ID of the HRI Management API within your authorization service.")
	fs.BoolVar(&config.Validation, "validation", false, "(Optional) True to enable record validation, false otherwise")
	fs.StringVar(&config.ElasticUrl, "elastic-url", "", "(Optional) The base url to the Elasticsearch instance")
	fs.StringVar(&config.ElasticUsername, "elastic-username", "", "(Optional) Elasticsearch user name")
	fs.StringVar(&config.ElasticPassword, "elastic-password", "", "(Optional) Elasticsearch password")
	fs.StringVar(&config.ElasticCert, "elastic-cert", "", "(Optional) Elasticsearch TLS public certificate")
	fs.StringVar(&config.ElasticServiceCrn, "elastic-crn", "", "(Optional) Elasticsearch service CRN")
	fs.StringVar(&config.KafkaAdminUrl, "kafka-admin-url", "", "(Optional) Kafka administration url")
	fs.Var(&config.KafkaBrokers, "kafka-brokers", "(Optional) A list of Kafka brokers, separated by \",\"")
	fs.Var(&config.KafkaProperties, "kafka-properties", "(Optional) A list of Kafka properties, entries separated by \",\", key value pairs separated by \":\"")
	fs.StringVar(&config.LogLevel, "log-level", "info", "(Optional) Minimum Log Level for logging output. Available levels are: Trace, Debug, Info, Warning, Error, Fatal and Panic.")
	fs.BoolVar(&config.NewRelicEnabled, "new-relic-enabled", false, "(Optional) True to enable New Relic monitoring, false otherwise")
	fs.StringVar(&config.NewRelicAppName, "new-relic-app-name", "", "(Optional) Application name to aggregate data under in New Relic")
	fs.StringVar(&config.NewRelicLicenseKey, "new-relic-license-key", "", "(Optional) New Relic license key")
	fs.BoolVar(&config.TlsEnabled, "tls-enabled", false, "(Optional) Toggle enabling an encrypted connection via TLS")
	fs.StringVar(&config.TlsCertPath, "tls-cert-path", "", "(Optional) path of TLS certificate signed by the Kubernetes CA")
	fs.StringVar(&config.TlsKeyPath, "tls-key-path", "", "(Optional) path of key from TLS certificate signed by the Kubernetes CA")

	err := ff.Parse(fs, commandLineFlags,
		ff.WithIgnoreUndefined(true),
		ff.WithConfigFileFlag("config-path"),
		ff.WithConfigFileParser(ffyaml.Parser),
		ff.WithAllowMissingConfigFile(false),
		ff.WithEnvVarNoPrefix(),
	)
	if err != nil {
		// If the issue isn't related to CL args, also print the usage guide
		// (it will get printed automatically otherwise).
		if !strings.Contains(err.Error(), "error parsing commandline args: invalid ") {
			fs.Usage()
		}
		return config, err
	}
	err = ValidateConfig(config)
	if err != nil {
		fs.Usage()
		return Config{}, err
	}
	return config, nil
}

func isValidUrl(str string) bool {
	u, err := url.ParseRequestURI(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
