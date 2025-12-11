package handler

import (
	"core/internal/common"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

// if B is empty, returns A, else return B
func strNonEmpty(A, B string) string {
	if B != "" {
		return B
	}
	return A
}

var (
	validate = validator.New(validator.WithRequiredStructEnabled())
	errTrans ut.Translator
	appUtils = common.NewAppUtils()
)

func init() {
	// register any custom validations here
	en := en.New()
	uni := ut.New(en, en)
	errTrans, _ = uni.GetTranslator("en")
	enTranslations.RegisterDefaultTranslations(validate, errTrans)
}
