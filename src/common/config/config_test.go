/**
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package config

import (
	"errors"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"testing"
)

const testCert = `-----BEGIN CERTIFICATE-----
MIIDDzCCAfegAwIBAgIJANEH58y2/kzHMA0GCSqGSIb3DQEBCwUAMB4xHDAaBgNV
BAMME0lCTSBDbG91ZCBEYXRhYmFzZXMwHhcNMTgwNjI1MTQyOTAwWhcNMjgwNjIy
MTQyOTAwWjAeMRwwGgYDVQQDDBNJQk0gQ2xvdWQgRGF0YWJhc2VzMIIBIjANBgkq
hkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA8lpaQGzcFdGqeMlmqjffMPpIQhqpd8qJ
Pr3bIkrXJbTcJJ9uIckSUcCjw4Z/rSg8nnT13SCcOl+1to+7kdMiU8qOWKiceYZ5
y+yZYfCkGaiZVfazQBm45zBtFWv+AB/8hfCTdNF7VY4spaA3oBE2aS7OANNSRZSK
pwy24IUgUcILJW+mcvW80Vx+GXRfD9Ytt6PRJgBhYuUBpgzvngmCMGBn+l2KNiSf
weovYDCD6Vngl2+6W9QFAFtWXWgF3iDQD5nl/n4mripMSX6UG/n6657u7TDdgkvA
1eKI2FLzYKpoKBe5rcnrM7nHgNc/nCdEs5JecHb1dHv1QfPm6pzIxwIDAQABo1Aw
TjAdBgNVHQ4EFgQUK3+XZo1wyKs+DEoYXbHruwSpXjgwHwYDVR0jBBgwFoAUK3+X
Zo1wyKs+DEoYXbHruwSpXjgwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOC
AQEAJf5dvlzUpqaix26qJEuqFG0IP57QQI5TCRJ6Xt/supRHo63eDvKw8zR7tlWQ
lV5P0N2xwuSl9ZqAJt7/k/3ZeB+nYwPoyO3KvKvATunRvlPBn4FWVXeaPsG+7fhS
qsejmkyonYw77HRzGOzJH4Zg8UN6mfpbaWSsyaExvqknCp9SoTQP3D67AzWqb1zY
doqqgGIZ2nxCkp5/FXxF/TMb55vteTQwfgBy60jVVkbF7eVOWCv0KaNHPF5hrqbN
i+3XjJ7/peF3xMvTMoy35DcT3E2ZeSVjouZs15O90kI3k2daS2OHJABW0vSj4nLz
+PQzp/B9cQmOO8dCe049Q3oaUA==
-----END CERTIFICATE-----
`

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
				OidcIssuer:         "https://us-south.appid.cloud.ibm.com/oauth/v4/",
				JwtAudienceId:      "",
				Validation:         false,
				ElasticUrl:         "https://ibm.com",
				ElasticUsername:    "elasticUsername",
				ElasticPassword:    "elasticPassword",
				ElasticCert:        testCert,
				ElasticServiceCrn:  "elasticServiceCrn",
				KafkaUsername:      "kafkaUsername",
				KafkaPassword:      "kafkaPassword",
				KafkaAdminUrl:      "https://ibm.kafka.com",
				KafkaBrokers:       StringSlice{"broker 1", "broker 2"},
				LogLevel:           "info",
				NewRelicEnabled:    true,
				NewRelicAppName:    "nrAppName",
				NewRelicLicenseKey: "nrLicenseKey",
				TlsEnabled:         true,
				TlsCertPath:        "./server-cert.pem",
				TlsKeyPath:         "./server-key.pem",
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
				"\n\tA Kafka username was not specified\n\tA Kafka password was not specified" +
				"\n\tThe Kafka administration url was not specified" + "\n\tNo Kafka brokers were defined" +
				"\n\tTLS is enabled but a path to a TLS certificate for the server was not specified" +
				"\n\tTLS is enabled but a path to a TLS key for the server was not specified",
		},
		{
			name: "invalid oidc issuer url",
			config: Config{
				ConfigPath:         "validPath",
				AuthDisabled:       false,
				OidcIssuer:         "invalidUrl.gov",
				JwtAudienceId:      "",
				Validation:         false,
				ElasticUrl:         "https://ibm.com",
				ElasticUsername:    "elasticUsername",
				ElasticPassword:    "elasticPassword",
				ElasticCert:        testCert,
				ElasticServiceCrn:  "elasticServiceCrn",
				KafkaUsername:      "kafkaUsername",
				KafkaPassword:      "kafkaPassword",
				KafkaAdminUrl:      "https://ibm.kafka.com",
				KafkaBrokers:       StringSlice{"broker 1", "broker 2"},
				LogLevel:           "info",
				NewRelicEnabled:    true,
				NewRelicAppName:    "nrAppName",
				NewRelicLicenseKey: "nrLicenseKey",
			},
			expectedErrMsg: "Configuration errors:\n\tOIDC Issuer is an invalid URL:  invalidUrl.gov",
		},
		{
			name: "nr enabled but no app name or license",
			config: Config{
				ConfigPath:        "validPath",
				AuthDisabled:      false,
				OidcIssuer:        "https://us-south.appid.cloud.ibm.com/oauth/v4/",
				JwtAudienceId:     "",
				Validation:        false,
				ElasticUrl:        "https://ibm.com",
				ElasticUsername:   "elasticUsername",
				ElasticPassword:   "elasticPassword",
				ElasticCert:       testCert,
				ElasticServiceCrn: "elasticServiceCrn",
				KafkaUsername:     "kafkaUsername",
				KafkaPassword:     "kafkaPassword",
				KafkaAdminUrl:     "https://ibm.kafka.com",
				KafkaBrokers:      StringSlice{"broker 1", "broker 2"},
				LogLevel:          "info",
				NewRelicEnabled:   true,
			},
			expectedErrMsg: "Configuration errors:\n\tNew Relic monitoring enabled, but the New Relic app name was not specified\n\tNew Relic monitoring enabled, but the New Relic license key was not specified",
		},
		{
			name: "AuthDisabled - no OidcIssuer is required",
			config: Config{
				ConfigPath:         "validPath",
				AuthDisabled:       true,
				JwtAudienceId:      "",
				Validation:         false,
				ElasticUrl:         "https://ibm.com",
				ElasticUsername:    "elasticUsername",
				ElasticPassword:    "elasticPassword",
				ElasticCert:        testCert,
				ElasticServiceCrn:  "elasticServiceCrn",
				KafkaUsername:      "kafkaUsername",
				KafkaPassword:      "kafkaPassword",
				KafkaAdminUrl:      "https://ibm.kafka.com",
				KafkaBrokers:       StringSlice{"broker 1", "broker 2"},
				LogLevel:           "info",
				NewRelicEnabled:    true,
				NewRelicAppName:    "nrAppName",
				NewRelicLicenseKey: "nrLicenseKey",
				TlsEnabled:         true,
				TlsCertPath:        "./server-cert.pem",
				TlsKeyPath:         "./server-key.pem",
			},
		},
		{
			name: "Bad elasticsearch certificate",
			config: Config{
				ConfigPath:         "validPath",
				AuthDisabled:       false,
				OidcIssuer:         "https://us-south.appid.cloud.ibm.com/oauth/v4/",
				JwtAudienceId:      "",
				Validation:         false,
				ElasticUrl:         "https://elastic.com",
				ElasticUsername:    "elasticUsername",
				ElasticPassword:    "elasticPassword",
				ElasticCert:        "Invalid Certificate",
				ElasticServiceCrn:  "elasticServiceCrn",
				KafkaUsername:      "kafkaUsername",
				KafkaPassword:      "kafkaPassword",
				KafkaAdminUrl:      "https://ibm.kafka.com",
				KafkaBrokers:       StringSlice{"broker 1", "broker 2"},
				LogLevel:           "info",
				NewRelicEnabled:    true,
				NewRelicAppName:    "nrAppName",
				NewRelicLicenseKey: "nrLicenseKey",
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
			expectedErrMsg:          "open invalid config path: no such file or directory",
		},
		{
			name:           "invalid config path passed in through env var",
			envVars:        [][2]string{{"CONFIG_PATH", "invalid config path"}},
			expectedErrMsg: "open invalid config path: no such file or directory",
		},
		{
			name:             "invalid config path passed in through flag",
			commandLineFlags: []string{"-config-path=invalid config path"},
			expectedErrMsg:   "open invalid config path: no such file or directory",
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
				OidcIssuer:         "http://ValFromEnv.gov",
				JwtAudienceId:      "ValFromFlag",
				Validation:         true,
				ElasticUrl:         "https://elastic.com",
				ElasticUsername:    "elasticUsername",
				ElasticPassword:    "elasticPassword",
				ElasticCert:        testCert,
				ElasticServiceCrn:  "elasticCrn",
				KafkaUsername:      "kafkaUsername",
				KafkaPassword:      "kafkaPassword",
				KafkaAdminUrl:      "https://ibm.kafka.com",
				KafkaBrokers:       StringSlice{"broker1", "broker2"},
				LogLevel:           "info",
				NewRelicEnabled:    true,
				NewRelicAppName:    "nrAppName",
				NewRelicLicenseKey: "nrLicenseKey0000000000000000000000000000",
				TlsEnabled:         true,
				TlsCertPath:        "./server-cert.pem",
				TlsKeyPath:         "./server-key.pem",
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
	if c.OidcIssuer != "" {
		return true
	}
	if c.JwtAudienceId != "" {
		return true
	}
	return false
}
