package validation

import (
	"context"
	"errors"
	"net/http"
	"reflect"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/contrib/resources"
	"fx.prodigy9.co/httpserver/render"
	"github.com/ggicci/httpin"
	"github.com/go-chi/chi/v5"
)

/**
 * httpin package used to read request data into structs
 * govalidator package used for validation
 *
 * Use header "x-api-action" for conditional validation
 */

// Validator adds a validator to the request
func Middleware(cfg *config.Source) func(http.Handler) http.Handler {
	val := &Validator{}
	val.Init()
	val.AddTranslations(Translations{
		"required":             "{0} is required",
		"required_if":          "{0} is required",
		"required_with":        "{0} is required",
		"required_unless":      "{0} is required",
		"required_without_all": "{0} is required",
		"required_without":     "{0} is required",
		"email":                "Email is invalid",
	})

	httpin.UseGochiURLParam("path", chi.URLParam)
	httpin.RegisterDirectiveExecutor("validator", httpin.DirectiveExecutorFunc(executor(val)), nil)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), "validator", val)
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}

// executor will validate a struct field marked with the "validator" tag
func executor(val *Validator) func(ctx *httpin.DirectiveContext) error {
	return func(ctx *httpin.DirectiveContext) error {
		action := ctx.Request.Header.Get("x-api-action")
		if field := ctx.Value.Elem().FieldByName("Action"); field != (reflect.Value{}) {
			field.SetString(action)
		}
		return val.Validate(ctx.Value.Elem().Interface(), ctx.ValueType)
	}
}

// errorHandler will unwrap httpin error to pass the fx validation errors to render
func errorHandler(rw http.ResponseWriter, r *http.Request, err error) {
	if httpinErr, ok := errors.Unwrap(err).(*httpin.InvalidFieldError); ok {
		unwraped := httpinErr.Unwrap()
		if unwraped.Error() == "EOF" {
			unwraped = errors.New("error parsing body")
		}
		render.Error(rw, r, 400, unwraped)
	}
}

// Validation returns the request reader and struct validator middlewares
//
// First field structs of form with tag "validator" will be validated and then the form
func Validation(form interface{}) []func(next http.Handler) http.Handler {
	return []func(next http.Handler) http.Handler{
		requestToStruct(form),
		structValidator(form),
		resources.RouteResourceMapper(),
	}
}

// requestToStruct reads the request data into the struct
func requestToStruct(form interface{}) func(http.Handler) http.Handler {
	return httpin.NewInput(form, httpin.WithErrorHandler(errorHandler))
}

// structValidator validates the root struct with go validator
func structValidator(form interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			val := req.Context().Value("validator").(*Validator)
			input := req.Context().Value(httpin.Input)

			action := req.Header.Get("x-api-action")
			refVal := reflect.ValueOf(input)
			if field := refVal.Elem().FieldByName("Action"); field != (reflect.Value{}) {
				field.SetString(action)
			}

			errs := val.Validate(refVal.Elem().Interface(), reflect.TypeOf(form))

			if errs != nil {
				render.Error(rw, req, 400, errs)
				return
			}
			next.ServeHTTP(rw, req)
		})
	}
}

// ValidationMessage will set the translations for the current app overwriting existing ones
func ValidationMessage(translations Translations) func(cfg *config.Source) func(http.Handler) http.Handler {
	return func(cfg *config.Source) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				val := req.Context().Value("validator").(*Validator)
				val.AddTranslations(translations)
				ctx := context.WithValue(req.Context(), "validator", val)
				next.ServeHTTP(rw, req.WithContext(ctx))
			})
		}
	}
}
