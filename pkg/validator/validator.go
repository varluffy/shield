// Package validator provides multilingual validation error translation.
// It supports Chinese and English validation error messages with custom field labels.
package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

// Validator 多语言验证器
type Validator struct {
	validator  *validator.Validate
	translator ut.Translator
	language   string
}

// SupportedLanguages 支持的语言
var SupportedLanguages = []string{"zh", "en"}

// DefaultLanguage 默认语言
const DefaultLanguage = "zh"

// NewValidator 创建多语言验证器
func NewValidator(language string) (*Validator, error) {
	// 验证语言是否支持
	if !isLanguageSupported(language) {
		language = DefaultLanguage
	}

	// 创建验证器
	validate := validator.New()

	// 注册自定义标签名函数
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		// 优先使用label标签，然后是json标签，最后是字段名
		if label := fld.Tag.Get("label"); label != "" {
			return label
		}

		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		if name == "" {
			name = fld.Name
		}
		return name
	})

	// 注册binding标签解析器
	validate.SetTagName("binding")

	// 创建通用翻译器
	zhLocale := zh.New()
	enLocale := en.New()
	uni := ut.New(zhLocale, zhLocale, enLocale)

	// 根据语言获取翻译器
	translator, found := uni.GetTranslator(language)
	if !found {
		return nil, fmt.Errorf("translator for language '%s' not found", language)
	}

	// 注册翻译
	var err error
	switch language {
	case "zh":
		err = zhTranslations.RegisterDefaultTranslations(validate, translator)
	case "en":
		err = enTranslations.RegisterDefaultTranslations(validate, translator)
	default:
		err = zhTranslations.RegisterDefaultTranslations(validate, translator)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to register translations: %w", err)
	}

	// 注册自定义翻译
	registerCustomTranslations(validate, translator, language)

	v := &Validator{
		validator:  validate,
		translator: translator,
		language:   language,
	}

	// 替换gin的默认验证器
	binding.Validator = &ginValidatorWrapper{validator: v}

	return v, nil
}

// Validate 验证结构体
func (v *Validator) Validate(obj interface{}) error {
	return v.validator.Struct(obj)
}

// ValidateAndTranslate 验证并翻译错误信息
func (v *Validator) ValidateAndTranslate(obj interface{}) ([]string, error) {
	err := v.validator.Struct(obj)
	if err == nil {
		return nil, nil
	}

	var messages []string

	// 类型断言为ValidationErrors
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return []string{err.Error()}, err
	}

	// 翻译每个验证错误
	for _, fieldError := range validationErrors {
		message := fieldError.Translate(v.translator)
		messages = append(messages, message)
	}

	return messages, err
}

// TranslateError 翻译验证错误
func (v *Validator) TranslateError(err error) []string {
	if err == nil {
		return nil
	}

	var messages []string

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return []string{err.Error()}
	}

	for _, fieldError := range validationErrors {
		message := fieldError.Translate(v.translator)
		messages = append(messages, message)
	}

	return messages
}

// GetLanguage 获取当前语言
func (v *Validator) GetLanguage() string {
	return v.language
}

// SetLanguage 设置语言
func (v *Validator) SetLanguage(language string) error {
	if !isLanguageSupported(language) {
		return fmt.Errorf("unsupported language: %s", language)
	}

	newValidator, err := NewValidator(language)
	if err != nil {
		return err
	}

	v.validator = newValidator.validator
	v.translator = newValidator.translator
	v.language = language

	return nil
}

// isLanguageSupported 检查语言是否支持
func isLanguageSupported(language string) bool {
	for _, lang := range SupportedLanguages {
		if lang == language {
			return true
		}
	}
	return false
}

// registerCustomTranslations 注册自定义翻译
func registerCustomTranslations(validate *validator.Validate, translator ut.Translator, language string) {
	switch language {
	case "zh":
		registerChineseTranslations(validate, translator)
	case "en":
		registerEnglishTranslations(validate, translator)
	}
}

// registerChineseTranslations 注册中文翻译
func registerChineseTranslations(validate *validator.Validate, translator ut.Translator) {
	translations := []struct {
		tag     string
		message string
	}{
		{"required", "{0}为必填字段"},
		{"email", "{0}必须是有效的邮箱地址"},
		{"min", "{0}长度必须至少为{1}个字符"},
		{"max", "{0}长度不能超过{1}个字符"},
		{"len", "{0}长度必须等于{1}个字符"},
		{"alphanum", "{0}只能包含字母和数字"},
		{"alpha", "{0}只能包含字母"},
		{"numeric", "{0}只能包含数字"},
		{"gt", "{0}必须大于{1}"},
		{"gte", "{0}必须大于或等于{1}"},
		{"lt", "{0}必须小于{1}"},
		{"lte", "{0}必须小于或等于{1}"},
		{"oneof", "{0}必须是以下值之一: {1}"},
		{"unique", "{0}中的值必须是唯一的"},
		{"url", "{0}必须是有效的URL"},
		{"uri", "{0}必须是有效的URI"},
		{"phone", "{0}必须是有效的手机号码"},
		{"datetime", "{0}必须是有效的日期时间格式"},
		{"password", "{0}必须包含大小写字母、数字和特殊字符"},
		{"strong_password", "{0}强度不够，必须包含大小写字母、数字和特殊字符"},
	}

	for _, translation := range translations {
		registerTranslation(validate, translator, translation.tag, translation.message)
	}
}

// registerEnglishTranslations 注册英文翻译
func registerEnglishTranslations(validate *validator.Validate, translator ut.Translator) {
	translations := []struct {
		tag     string
		message string
	}{
		{"required", "{0} is required"},
		{"email", "{0} must be a valid email address"},
		{"min", "{0} must be at least {1} characters long"},
		{"max", "{0} cannot be longer than {1} characters"},
		{"len", "{0} must be exactly {1} characters long"},
		{"alphanum", "{0} can only contain letters and numbers"},
		{"alpha", "{0} can only contain letters"},
		{"numeric", "{0} can only contain numbers"},
		{"gt", "{0} must be greater than {1}"},
		{"gte", "{0} must be greater than or equal to {1}"},
		{"lt", "{0} must be less than {1}"},
		{"lte", "{0} must be less than or equal to {1}"},
		{"oneof", "{0} must be one of: {1}"},
		{"unique", "{0} values must be unique"},
		{"url", "{0} must be a valid URL"},
		{"uri", "{0} must be a valid URI"},
		{"phone", "{0} must be a valid phone number"},
		{"datetime", "{0} must be a valid datetime format"},
		{"password", "{0} must contain uppercase, lowercase, numbers and special characters"},
		{"strong_password", "{0} is not strong enough, must contain uppercase, lowercase, numbers and special characters"},
	}

	for _, translation := range translations {
		registerTranslation(validate, translator, translation.tag, translation.message)
	}
}

// registerTranslation 注册单个翻译
func registerTranslation(validate *validator.Validate, translator ut.Translator, tag, message string) {
	validate.RegisterTranslation(tag, translator, func(ut ut.Translator) error {
		return ut.Add(tag, message, true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(tag, fe.Field(), fe.Param())
		return t
	})
}

// ginValidatorWrapper gin验证器包装器
type ginValidatorWrapper struct {
	validator *Validator
}

// ValidateStruct 实现gin的验证器接口
func (w *ginValidatorWrapper) ValidateStruct(obj interface{}) error {
	return w.validator.Validate(obj)
}

// Engine 返回验证器引擎
func (w *ginValidatorWrapper) Engine() interface{} {
	return w.validator.validator
}

// Global validator instance
var globalValidator *Validator

// InitGlobalValidator 初始化全局验证器
func InitGlobalValidator(language string) error {
	var err error
	globalValidator, err = NewValidator(language)
	return err
}

// GetGlobalValidator 获取全局验证器
func GetGlobalValidator() *Validator {
	if globalValidator == nil {
		// 如果没有初始化，使用默认语言
		globalValidator, _ = NewValidator(DefaultLanguage)
	}
	return globalValidator
}

// ValidateStruct 使用全局验证器验证结构体
func ValidateStruct(obj interface{}) ([]string, error) {
	return GetGlobalValidator().ValidateAndTranslate(obj)
}

// TranslateValidationError 翻译验证错误
func TranslateValidationError(err error) []string {
	return GetGlobalValidator().TranslateError(err)
}
