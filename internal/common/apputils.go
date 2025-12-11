package common

import (
	"context"
	"core/app"
	"core/internal/common/templates"
	"core/internal/domain_model/model"
	"core/internal/repo"
	"core/pkg/storage/s3"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/mold/v4"
	"github.com/go-playground/mold/v4/modifiers"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	f "github.com/gofiber/fiber/v2"
)

type AppUtils struct {
	Conform    *mold.Transformer
	Validator  *validator.Validate
	Translator ut.Translator
	Templates  *template.Template
	S3Cred     s3.KeyS3
}

//go:embed templates/*
var templateFs embed.FS

func NewAppUtils() *AppUtils {
	tmpl := template.Must(
		template.
			New("templates").
			Funcs(templates.TemplateFuncs()).
			ParseFS(templateFs, "templates/*.html"),
	)

	en := en.New()
	translator, _ := ut.New(en, en).GetTranslator("en")

	conform := modifiers.New()
	validate := validator.New(validator.WithRequiredStructEnabled())

	enTranslations.RegisterDefaultTranslations(validate, translator)

	accessKey := app.Config("S3_ACCESS_KEY")
	secKey := app.Config("S3_SECRET_KEY")
	endpoint := app.Config("S3_ENDPOINT_API")
	endpoint_cdn := app.Config("S3_ENDPOINT_CDN")
	bucketName := app.Config("S3_BUCKET")
	isSSL := app.Config("S3_SSL") == "true"

	s3Credential := s3.KeyS3{
		Bucket:      bucketName,
		AccessKey:   accessKey,
		SecretKey:   secKey,
		EndPointApi: endpoint,
		EndPointCdn: endpoint_cdn,
		IsSSL:       isSSL,
	}

	return &AppUtils{
		Conform:    conform,
		Validator:  validate,
		Translator: translator,
		Templates:  tmpl,
		S3Cred:     s3Credential,
	}
}

// NOTE: validation only works if data is a struct
func (a *AppUtils) Validate(ctx context.Context, data any) error {
	if err := a.Conform.Struct(ctx, data); err != nil {
		return err
	}

	if err := a.Validator.StructCtx(ctx, data); err != nil {
		validationErr, ok := err.(validator.ValidationErrors)

		if !ok || a.Translator == nil {
			return err
		}

		errStrings := []string{}
		for _, fieldErr := range validationErr {
			errStrings = append(errStrings, fieldErr.Translate(a.Translator))
		}

		return fmt.Errorf("validation failed: %s", strings.Join(errStrings, ", "))
	}

	return nil
}

func (a *AppUtils) ParseBody(c *f.Ctx, body any) error {
	if err := c.BodyParser(body); err != nil {
		return err
	}

	if err := a.Conform.Struct(c.Context(), body); err != nil {
		return err
	}

	if err := a.Validator.Struct(body); err != nil {
		validationErr, ok := err.(validator.ValidationErrors)

		if !ok || a.Translator == nil {
			return err
		}

		errStrings := []string{}
		for _, fieldErr := range validationErr {
			errStrings = append(errStrings, fieldErr.Translate(a.Translator))
		}

		return fmt.Errorf("validation failed: %s", strings.Join(errStrings, ", "))
	}

	return nil
}

func NewS3SystemKey() s3.KeyS3 {
	accessKey := app.Config("S3_ACCESS_KEY")
	secKey := app.Config("S3_SECRET_KEY")
	endpoint := app.Config("S3_ENDPOINT_API")
	endpointCdn := app.Config("S3_ENDPOINT_CDN")
	bucketName := app.Config("S3_BUCKET")
	isSSL := app.Config("S3_SSL") == "true"

	key := s3.KeyS3{
		Bucket:      bucketName,
		AccessKey:   accessKey,
		SecretKey:   secKey,
		EndPointApi: endpoint,
		EndPointCdn: endpointCdn,
		IsSSL:       isSSL,
	}

	return key
}

func S3Credential(ctx context.Context) s3.KeyS3 {
	systemSetting, err := repo.GetSystemSetting(ctx, model.STORAGE_CREDENTIAL)
	if err == nil {
		var setting model.StorageCredential
		if err := json.Unmarshal(systemSetting.Metadata, &setting); err == nil {
			return s3.KeyS3(setting)
		}
	}
	return NewS3SystemKey()
}

func ParseFloatPtr(s *string) *float64 {
	if s == nil {
		return nil
	}
	str := strings.TrimSpace(*s)
	if str == "" {
		return nil
	}
	str = strings.ToLower(str)

	str = strings.ReplaceAll(str, "kg", "")
	str = strings.ReplaceAll(str, "g", "")
	str = strings.TrimSpace(str)
	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return nil
	}
	return &v
}
