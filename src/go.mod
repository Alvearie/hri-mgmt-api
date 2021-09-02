module github.com/Alvearie/hri-mgmt-api

require (
	github.com/IBM/event-streams-go-sdk-generator v1.0.0
	github.com/IBM/resource-controller-go-sdk-generator v1.0.1
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/elastic/go-elasticsearch/v7 v7.11.0
	github.com/go-playground/locales v0.13.0
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-playground/validator/v10 v10.4.1
	github.com/golang/mock v1.5.0
	github.com/kr/pretty v0.1.0 // indirect
	github.com/labstack/echo/v4 v4.2.0
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/newrelic/go-agent/v3 v3.11.0
	github.com/newrelic/go-agent/v3/integrations/nrecho-v4 v1.0.1
	github.com/peterbourgon/ff/v3 v3.1.0
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/segmentio/kafka-go v0.3.5
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/Sirupsen/logrus v1.8.1 => github.com/sirupsen/logrus v1.8.1

go 1.15
