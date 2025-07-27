package response

// ValidationErrorTranslator 验证错误翻译函数类型
type ValidationErrorTranslator func(error) []string

// 全局验证错误翻译器
var globalValidationErrorTranslator ValidationErrorTranslator

// SetValidationErrorTranslator 设置全局验证错误翻译器
func SetValidationErrorTranslator(translator ValidationErrorTranslator) {
	globalValidationErrorTranslator = translator
}
