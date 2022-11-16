/**
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/Alvearie/hri-mgmt-api/common/test"
	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	for _, tc := range []struct {
		name           string
		config         Config
		expectedErrMsg string
	}{
		{
			name: "valid config",
			config: Config{
				ConfigPath:         "validPath",
				AuthDisabled:       false,
				LogLevel:           "info",
				NewRelicEnabled:    true,
				NewRelicAppName:    "nrAppName",
				NewRelicLicenseKey: "nrLicenseKey",
				TlsEnabled:         true,
				TlsCertPath:        "./server-cert.pem",
				TlsKeyPath:         "./server-key.pem",
				MongoDBUri:         "mongoDbUri",
				MongoDBName:        "HRI-DEV",
				MongoColName:       "HRI-Mgmt",
				AzOidcIssuer:       "https://sts.windows.net/ceaa63aa-5d5c-4c7d-94b0-02f9a3ab6a8c/",
				AzJwtAudienceId:    "c33ac4da-21c6-426b-abcc-27e24ff1ccf9",
			},
		},
		{
			name: "no config file specified",
			config: Config{
				ConfigPath: "",
			},
			expectedErrMsg: "no config file supplied",
		},
		{
			name:   "Invalid Config Missing Required config Params",
			config: Config{ConfigPath: "validPath", TlsEnabled: true},
			expectedErrMsg: "Configuration errors:\n\tOIDC Issuer is an invalid URL:  " +
				"\n\tAn Elasticsearch base URL was not specified\n\tAn Elasticsearch username was not specified" +
				"\n\tAn Elasticsearch password was not specified\n\tAn Elasticsearch certificate was not specified" +
				"\n\tAn Elasticsearch service CRN was not specified" +
				"\n\tThe Kafka administration url was not specified" + "\n\tNo Kafka brokers were defined" +
				"\n\tTLS is enabled but a path to a TLS certificate for the server was not specified" +
				"\n\tTLS is enabled but a path to a TLS key for the server was not specified" +
				"\n\tMongoDB uri was not specified" +
				"\n\tMongoDB name was not specified" +
				"\n\tMongoDB collection name was not specified" +
				"\n\tAz AD OIDC Issuer is an invalid URL:  ",
		},
		{
			name: "invalid oidc issuer url",
			config: Config{
				ConfigPath:         "validPath",
				AuthDisabled:       false,
				LogLevel:           "info",
				NewRelicEnabled:    true,
				NewRelicAppName:    "nrAppName",
				NewRelicLicenseKey: "nrLicenseKey",
				MongoDBUri:         "mongoDbUri",
				MongoDBName:        "HRI-DEV",
				MongoColName:       "HRI-Mgmt",
				AzOidcIssuer:       "https://sts.windows.net/ceaa63aa-5d5c-4c7d-94b0-02f9a3ab6a8c/",
				AzJwtAudienceId:    "c33ac4da-21c6-426b-abcc-27e24ff1ccf9",
			},
			expectedErrMsg: "Configuration errors:\n\tOIDC Issuer is an invalid URL:  invalidUrl.gov",
		},
		{
			name: "nr enabled but no app name or license",
			config: Config{
				ConfigPath:      "validPath",
				AuthDisabled:    false,
				LogLevel:        "info",
				NewRelicEnabled: true,
				MongoDBUri:      "mongoDbUri",
				MongoDBName:     "HRI-DEV",
				MongoColName:    "HRI-Mgmt",
				AzOidcIssuer:    "https://sts.windows.net/ceaa63aa-5d5c-4c7d-94b0-02f9a3ab6a8c/",
				AzJwtAudienceId: "c33ac4da-21c6-426b-abcc-27e24ff1ccf9",
			},
			expectedErrMsg: "Configuration errors:\n\tNew Relic monitoring enabled, but the New Relic app name was not specified\n\tNew Relic monitoring enabled, but the New Relic license key was not specified",
		},
		{
			name: "AuthDisabled - no OidcIssuer is required",
			config: Config{
				ConfigPath:         "validPath",
				AuthDisabled:       true,
				LogLevel:           "info",
				NewRelicEnabled:    true,
				NewRelicAppName:    "nrAppName",
				NewRelicLicenseKey: "nrLicenseKey",
				TlsEnabled:         true,
				TlsCertPath:        "./server-cert.pem",
				TlsKeyPath:         "./server-key.pem",
				MongoDBUri:         "mongoDbUri",
				MongoDBName:        "HRI-DEV",
				MongoColName:       "HRI-Mgmt",
				AzOidcIssuer:       "https://sts.windows.net/ceaa63aa-5d5c-4c7d-94b0-02f9a3ab6a8c/",
				AzJwtAudienceId:    "c33ac4da-21c6-426b-abcc-27e24ff1ccf9",
			},
		},
		{
			name: "Bad elasticsearch certificate",
			config: Config{
				ConfigPath:         "validPath",
				AuthDisabled:       false,
				LogLevel:           "info",
				NewRelicEnabled:    true,
				NewRelicAppName:    "nrAppName",
				NewRelicLicenseKey: "nrLicenseKey",
				MongoDBUri:         "mongoDbUri",
				MongoDBName:        "HRI-DEV",
				MongoColName:       "HRI-Mgmt",
				AzOidcIssuer:       "https://sts.windows.net/ceaa63aa-5d5c-4c7d-94b0-02f9a3ab6a8c/",
				AzJwtAudienceId:    "c33ac4da-21c6-426b-abcc-27e24ff1ccf9",
			},
			expectedErrMsg: "Configuration errors:\n\tThe Elasticsearch certificate is invalid",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateConfig(tc.config)
			if err != nil {
				assert.Equal(t, tc.expectedErrMsg, err.Error())
			} else {
				assert.Equal(t, tc.expectedErrMsg, "")
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	configPath := test.FindConfigPath(t)
	for _, tc := range []struct {
		name                    string
		commandLineFlags        []string
		envVars                 [][2]string // key-value pairs
		passAlternateConfigPath bool
		configPath              string
		expectedConfig          Config
		expectedErrMsg          string
	}{
		{
			name: "empty call",
		},
		{
			name:                    "no config path passed in anywhere",
			passAlternateConfigPath: true,
			expectedErrMsg:          "no config file supplied",
		},
		{
			name:                    "invalid config path passed in directly",
			passAlternateConfigPath: true,
			configPath:              "invalid config path",
			//expectedErrMsg:          "open invalid config path: no such file or directory",
			expectedErrMsg: "open invalid config path: The system cannot find the file specified.",
		},
		{
			name:    "invalid config path passed in through env var",
			envVars: [][2]string{{"CONFIG_PATH", "invalid config path"}},
			//expectedErrMsg: "open invalid config path: no such file or directory",
			expectedErrMsg: "open invalid config path: The system cannot find the file specified.",
		},
		{
			name:             "invalid config path passed in through flag",
			commandLineFlags: []string{"-config-path=invalid config path"},
			//expectedErrMsg:   "open invalid config path: no such file or directory",
			expectedErrMsg: "open invalid config path: The system cannot find the file specified.",
		},
		{
			name:           "send incorrect type var from env (string instead of bool)",
			envVars:        [][2]string{{"VALIDATION", "incorrect value"}},
			expectedErrMsg: "error parsing env vars: error setting flag \"validation\" from env var \"VALIDATION\": parse error",
		},
		{
			name:             "send incorrect type var from flag (string instead of bool)",
			commandLineFlags: []string{"-validation=incorrect"},
			expectedErrMsg:   "error parsing commandline args: invalid boolean value \"incorrect\" for -validation: parse error",
		},
		{
			name:             "variables override correctly (cl flags > env vars > config.yml)",
			commandLineFlags: []string{"-jwt-audience-id=ValFromFlag", "-validation=true", fmt.Sprintf("-kafka-brokers=%s,%s", "broker1", "broker2")},
			envVars:          [][2]string{{"OIDC_ISSUER", "http://ValFromEnv.gov"}, {"JWT_AUDIENCE_ID", "ValFromEnv"}},
			expectedConfig: Config{
				ConfigPath:         configPath,
				AuthDisabled:       false,
				LogLevel:           "info",
				NewRelicEnabled:    true,
				NewRelicAppName:    "nrAppName",
				NewRelicLicenseKey: "nrLicenseKey0000000000000000000000000000",
				TlsEnabled:         true,
				TlsCertPath:        "./server-cert.pem",
				TlsKeyPath:         "./server-key.pem",
				MongoDBUri:         "mongodb://hi",
				MongoDBName:        "HRI-DEV",
				MongoColName:       "HRI-Mgmt",
				AzOidcIssuer:       "https://sts.windows.net/ceaa63aa-5d5c-4c7d-94b0-02f9a3ab6a8c/",
				AzJwtAudienceId:    "c33ac4da-21c6-426b-abcc-27e24ff1ccf9",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Select the proper configPath to use
			testConfigPath := configPath
			if tc.passAlternateConfigPath {
				testConfigPath = tc.configPath
			}
			// Reset environment variables for a clean slate
			os.Clearenv()
			// Set the environment variables
			for _, envVar := range tc.envVars {
				os.Setenv(envVar[0], envVar[1])
			}
			config, err := GetConfig(testConfigPath, tc.commandLineFlags)
			// If GetConfig resulted in an error, there's no reason to compare configs
			if err == nil && expectedConfigExists(tc.expectedConfig) && !reflect.DeepEqual(tc.expectedConfig, config) {
				err = errors.New("generated config does not match the expected config")
			}

			if err != nil {
				assert.Equal(t, tc.expectedErrMsg, err.Error())
			} else {
				assert.Equal(t, tc.expectedErrMsg, "")
			}
		})
	}
}

// Whether specified or not, each test case will have a config struct in the expectConfig param.
// We define a "nonexistent config" to be a config whose string attributes are empty.
func expectedConfigExists(c Config) bool {
	if c.ConfigPath != "" {
		return true
	}
	// if c.OidcIssuer != "" {
	// 	return true
	// }
	// if c.JwtAudienceId != "" {
	// 	return true
	// }
	return false
}
