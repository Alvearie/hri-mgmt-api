module github.com/Alvearie/hri-mgmt-api

require (
	github.com/IBM/event-streams-go-sdk-generator v1.0.0
	github.com/IBM/resource-controller-go-sdk-generator v1.0.1
	github.com/confluentinc/confluent-kafka-go v1.7.0
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/elastic/go-elasticsearch/v7 v7.11.0
	github.com/go-playground/locales v0.13.0
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-playground/validator/v10 v10.4.1
	github.com/golang/mock v1.5.0
	github.com/kr/pretty v0.3.0 // indirect
	github.com/labstack/echo/v4 v4.6.1
	github.com/mattn/go-colorable v0.1.11 // indirect
	github.com/newrelic/go-agent/v3 v3.11.0
	github.com/newrelic/go-agent/v3/integrations/nrecho-v4 v1.0.1
	github.com/peterbourgon/ff/v3 v3.1.0
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/net v0.0.0-20211005001312-d4b1ae081e3b // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20211004093028-2c5d950f24ef // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/Sirupsen/logrus v1.8.1 => github.com/sirupsen/logrus v1.8.1

go 1.15
