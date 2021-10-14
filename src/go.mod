module github.com/Alvearie/hri-mgmt-api

require (
	github.com/IBM/event-streams-go-sdk-generator v1.0.0
	github.com/IBM/resource-controller-go-sdk-generator v1.0.1
	github.com/confluentinc/confluent-kafka-go v1.7.0
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/elastic/go-elasticsearch/v7 v7.11.0
	github.com/go-playground/locales v0.14.0
	github.com/go-playground/universal-translator v0.18.0
	github.com/go-playground/validator/v10 v10.9.0
	github.com/golang/mock v1.6.0
	github.com/labstack/echo/v4 v4.6.1
	github.com/mattn/go-colorable v0.1.11 // indirect
	github.com/newrelic/go-agent/v3 v3.15.0
	github.com/newrelic/go-agent/v3/integrations/nrecho-v4 v1.0.1
	github.com/peterbourgon/ff/v3 v3.1.2
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/net v0.0.0-20211005001312-d4b1ae081e3b // indirect
	golang.org/x/oauth2 v0.0.0-20211005180243-6b3c2da341f1 // indirect
	golang.org/x/sys v0.0.0-20211004093028-2c5d950f24ef // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20211005153810-c76a74d43a8e // indirect
	google.golang.org/grpc v1.41.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

// The jwt-go substitution is neccessary for nrecho-v4 to work, as it uses an old version
// of Echo that is dependent on a vulnerable dependency.  The golang-jwt vers. 4 library
// was designed to be able to be substitutable for jwt-go in this way.
replace (
	github.com/Sirupsen/logrus v1.8.1 => github.com/sirupsen/logrus v1.8.1
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v4 v4.0.0
)

go 1.15
