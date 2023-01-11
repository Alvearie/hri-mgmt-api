package model

// Custom Validator Tags
const (
	InjectionCheckValidatorTag string = "injection-check-validator"
	TenantIdValidatorTag       string = "tenantid-validator"
	StreamIdValidatorTag       string = "streamid-validator"
)

// Custom Validation RegEx strings
const (
	InjectionCheckContainsString string = `"=<>[]{}`
)
