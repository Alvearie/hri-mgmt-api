package model

import (
	"strings"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// Universal Translator Usage Docs:  https://pkg.go.dev/github.com/go-playground/universal-translator
// Universal Translator Example Files:  https://github.com/go-playground/universal-translator/tree/master/_examples

// The way translations have been implemented here is modelled after the way they are implemented in the validator
// library being used.  There is a large number of translations already built in for the built-in validation tags.
// The translations here expand on the existing ones by adding translations for custom validation tags and for built-in
// validation tags that are missing a built in translation.  It is also possible to overwrite existing built-in
// translations.
// Validator Library built-in translations: https://github.com/go-playground/validator/blob/master/translations/en/en.go

// The translator will always replace the struct field name of the variable whose error is being translated
// via the tag name function registered to the validator (In other words, the variable marked by {0} in the translation).
// However, for any additional fields mentioned in that field's error message, this function is not applied.
// To work around this, this structFieldMarker (and closing bracket ]) enclose the struct fields that are not
// automatically replaced in the translation template string, and after the translation is complete the
// structFieldMarker is used to identify those fields and replace them with their names.
const structFieldMarker = "[STRUCT_FIELD"

// The translator uses curly braces in its translation template string to mark where variable names will go (for example, {0}).
// Unfortunately, it does not offer a way to use curly braces in the translation template string outside of that role.
// These curly brace marker strings are used in place of where a curly brace would go in the translation, and
// post-translation they are replaced with actual curly braces.
const openCurlyBraceMarker = "[CURLY_OPEN]"
const closeCurlyBraceMarker = "[CURLY_CLOSE]"

func fixMissingCurlyBraces(s string) string {
	s = strings.Replace(s, openCurlyBraceMarker, "{", -1)
	s = strings.Replace(s, closeCurlyBraceMarker, "}", -1)
	return s
}

func RegisterCustomTranslations(v *validator.Validate, trans ut.Translator) (err error) {
	translations := []struct {
		tag             string
		translation     string
		override        bool
		customTransFunc validator.TranslationFunc
	}{
		{
			tag:             InjectionCheckValidatorTag,
			translation:     "{0} must not contain the following characters: \"=<>[]" + openCurlyBraceMarker + closeCurlyBraceMarker,
			override:        false,
			customTransFunc: translateFunc,
		},
		{
			tag:             TenantIdValidatorTag,
			translation:     "{0} may only contain lower-case alpha-numeric chars and the following 2 special chars: '-', '_'",
			override:        false,
			customTransFunc: translateFunc,
		},
		{
			tag:             StreamIdValidatorTag,
			translation:     "{0} may only contain lower-case alpha-numeric characters, no more than one '.', and the following 2 special chars: '-', '_'",
			override:        false,
			customTransFunc: translateFunc,
		},
		{
			tag:             "required_without",
			translation:     "{0} must be present if " + structFieldMarker + "{1}] is not present",
			override:        false,
			customTransFunc: translateFuncWithValue,
		},
	}

	for _, t := range translations {
		err = v.RegisterTranslation(t.tag, trans, registrationFunc(t.tag, t.translation, t.override), t.customTransFunc)
		if err != nil {
			return
		}
	}
	return
}

// Use this as the translation function if the validation requires comparing the variable to another struct param.
// Another way to think about it is if the translation has both {0} and {1}, use this function.
// Otherwise, use translateFunc.
func translateFuncWithValue(ut ut.Translator, fe validator.FieldError) string {
	t, err := ut.T(fe.Tag(), fe.Field(), fe.Param())
	if err != nil {
		return fe.(error).Error()
	}
	t = fixMissingCurlyBraces(t)
	return t
}

func translateFunc(ut ut.Translator, fe validator.FieldError) string {
	t, err := ut.T(fe.Tag(), fe.Field())
	if err != nil {
		return fe.(error).Error()
	}
	t = fixMissingCurlyBraces(t)
	return t
}

func registrationFunc(tag string, translation string, override bool) validator.RegisterTranslationsFunc {
	return func(ut ut.Translator) error {
		return ut.Add(tag, translation, override)
	}
}
