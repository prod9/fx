package validation

import (
	"reflect"

	"fx.prodigy9.co/validate"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entranslations "github.com/go-playground/validator/v10/translations/en"
)

type Translations map[string]string

type Validator struct {
	Validator  *validator.Validate
	Translator ut.Translator
	Translations
}

func (v *Validator) Init() {
	v.Validator = validator.New()

	translator := en.New()
	uni := ut.New(translator, translator)

	v.Translator, _ = uni.GetTranslator("en")
	if err := entranslations.RegisterDefaultTranslations(v.Validator, v.Translator); err != nil {
		panic(err)
	}
}

// AddTranslations adds custom error messages for govalidator tags
func (v *Validator) AddTranslations(translations Translations) {
	for tag, translation := range translations {
		v.addTranslation(tag, translation)
	}
}

func (v *Validator) addTranslation(name, message string) {
	_ = v.Validator.RegisterTranslation(name, v.Translator,
		func(ut ut.Translator) error {
			return ut.Add(name, message, true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T(name, fe.Field())
			return t
		},
	)
}

// Validate validates a struct with go validator and formats the errors
func (v *Validator) Validate(value interface{}, valueType reflect.Type) error {
	errs := &validate.Error{}
	err := v.Validator.Struct(value)

	if err == nil {
		return nil
	}

	for _, err := range err.(validator.ValidationErrors) {
		if field, ok := valueType.FieldByName(err.Field()); ok {
			errs = errs.AddField(
				field.Tag.Get("json"),
				err.Translate(v.Translator),
				err.Value(),
			)
		}
	}

	if errs.Len() == 0 {
		return nil
	} else {
		return errs
	}
}
