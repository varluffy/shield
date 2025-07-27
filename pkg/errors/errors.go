// Package errors provides custom error types and error handling utilities.
// It defines business error codes and error mapping functionality.
package errors

import (
	"fmt"
	"net/http"
)

// 业务错误码定义
const (
	// 通用错误码 (1000-1999)
	CodeSuccess         = 0    // 成功
	CodeInternalError   = 1000 // 内部错误
	CodeInvalidRequest  = 1001 // 无效请求
	CodeValidationError = 1002 // 验证错误
	CodeUnauthorized    = 1003 // 未授权
	CodeForbidden       = 1004 // 禁止访问
	CodeNotFound        = 1005 // 未找到
	CodeConflict        = 1006 // 冲突
	CodeRateLimitError  = 1007 // 请求频率限制
	CodeTimeout         = 1008 // 请求超时

	// 用户相关错误码 (2000-2999)
	CodeUserNotFound        = 2001 // 用户不存在
	CodeUserAlreadyExists   = 2002 // 用户已存在
	CodeUserInactive        = 2003 // 用户未激活
	CodeUserLocked          = 2004 // 用户被锁定
	CodeInvalidCredentials  = 2005 // 凭据无效
	CodePasswordTooWeak     = 2006 // 密码过弱
	CodeUserEmailExists     = 2007 // 邮箱已存在
	CodeUserUsernameExists  = 2008 // 用户名已存在
	CodeUserPermissionError = 2009 // 用户权限不足
	CodeCaptchaRequired     = 2010 // 需要验证码
	CodeCaptchaInvalid      = 2011 // 验证码无效
	CodeCaptchaExpired      = 2012 // 验证码已过期
	CodeCaptchaGenerate     = 2013 // 验证码生成失败

	// 数据库相关错误码 (3000-3999)
	CodeDatabaseError       = 3001 // 数据库错误
	CodeRecordNotFound      = 3002 // 记录不存在
	CodeDuplicateEntry      = 3003 // 重复记录
	CodeConstraintViolation = 3004 // 约束违反
	CodeTransactionError    = 3005 // 事务错误

	// 外部服务错误码 (4000-4999)
	CodeExternalServiceError     = 4001 // 外部服务错误
	CodeThirdPartyServiceTimeout = 4002 // 第三方服务超时
	CodeAPIRateLimitExceeded     = 4003 // API调用频率超限

	// 文件处理错误码 (5000-5999)
	CodeFileNotFound        = 5001 // 文件不存在
	CodeFileUploadError     = 5002 // 文件上传错误
	CodeFileFormatError     = 5003 // 文件格式错误
	CodeFileSizeExceeded    = 5004 // 文件大小超限
	CodeFilePermissionError = 5005 // 文件权限错误
)

// 错误码到消息的映射
var errorMessages = map[int]string{
	CodeSuccess:         "success",
	CodeInternalError:   "内部服务器错误",
	CodeInvalidRequest:  "无效的请求",
	CodeValidationError: "参数验证失败",
	CodeUnauthorized:    "未授权访问",
	CodeForbidden:       "禁止访问",
	CodeNotFound:        "资源不存在",
	CodeConflict:        "资源冲突",
	CodeRateLimitError:  "请求频率超限",
	CodeTimeout:         "请求超时",

	CodeUserNotFound:        "用户不存在",
	CodeUserAlreadyExists:   "用户已存在",
	CodeUserInactive:        "用户未激活",
	CodeUserLocked:          "用户被锁定",
	CodeInvalidCredentials:  "用户名或密码错误",
	CodePasswordTooWeak:     "密码强度不足",
	CodeUserEmailExists:     "邮箱已被使用",
	CodeUserUsernameExists:  "用户名已被使用",
	CodeUserPermissionError: "用户权限不足",
	CodeCaptchaRequired:     "需要验证码",
	CodeCaptchaInvalid:      "验证码错误",
	CodeCaptchaExpired:      "验证码已过期",
	CodeCaptchaGenerate:     "验证码生成失败",

	CodeDatabaseError:       "数据库操作失败",
	CodeRecordNotFound:      "记录不存在",
	CodeDuplicateEntry:      "记录重复",
	CodeConstraintViolation: "数据约束违反",
	CodeTransactionError:    "事务处理失败",

	CodeExternalServiceError:     "外部服务调用失败",
	CodeThirdPartyServiceTimeout: "第三方服务超时",
	CodeAPIRateLimitExceeded:     "API调用频率超限",

	CodeFileNotFound:        "文件不存在",
	CodeFileUploadError:     "文件上传失败",
	CodeFileFormatError:     "文件格式不支持",
	CodeFileSizeExceeded:    "文件大小超出限制",
	CodeFilePermissionError: "文件权限不足",
}

// 错误码到HTTP状态码的映射
var codeToHTTPStatus = map[int]int{
	CodeSuccess:         http.StatusOK,
	CodeInternalError:   http.StatusInternalServerError,
	CodeInvalidRequest:  http.StatusBadRequest,
	CodeValidationError: http.StatusBadRequest,
	CodeUnauthorized:    http.StatusUnauthorized,
	CodeForbidden:       http.StatusForbidden,
	CodeNotFound:        http.StatusNotFound,
	CodeConflict:        http.StatusConflict,
	CodeRateLimitError:  http.StatusTooManyRequests,
	CodeTimeout:         http.StatusRequestTimeout,

	CodeUserNotFound:        http.StatusNotFound,
	CodeUserAlreadyExists:   http.StatusConflict,
	CodeUserInactive:        http.StatusForbidden,
	CodeUserLocked:          http.StatusForbidden,
	CodeInvalidCredentials:  http.StatusUnauthorized,
	CodePasswordTooWeak:     http.StatusBadRequest,
	CodeUserEmailExists:     http.StatusConflict,
	CodeUserUsernameExists:  http.StatusConflict,
	CodeUserPermissionError: http.StatusForbidden,
	CodeCaptchaRequired:     http.StatusBadRequest,
	CodeCaptchaInvalid:      http.StatusBadRequest,
	CodeCaptchaExpired:      http.StatusBadRequest,
	CodeCaptchaGenerate:     http.StatusInternalServerError,

	CodeDatabaseError:       http.StatusInternalServerError,
	CodeRecordNotFound:      http.StatusNotFound,
	CodeDuplicateEntry:      http.StatusConflict,
	CodeConstraintViolation: http.StatusBadRequest,
	CodeTransactionError:    http.StatusInternalServerError,

	CodeExternalServiceError:     http.StatusBadGateway,
	CodeThirdPartyServiceTimeout: http.StatusGatewayTimeout,
	CodeAPIRateLimitExceeded:     http.StatusTooManyRequests,

	CodeFileNotFound:        http.StatusNotFound,
	CodeFileUploadError:     http.StatusBadRequest,
	CodeFileFormatError:     http.StatusBadRequest,
	CodeFileSizeExceeded:    http.StatusRequestEntityTooLarge,
	CodeFilePermissionError: http.StatusForbidden,
}

// BusinessError 业务错误
type BusinessError struct {
	Code       int    `json:"code"`              // 业务错误码
	Message    string `json:"message"`           // 错误消息
	HTTPStatus int    `json:"-"`                 // HTTP状态码
	Details    string `json:"details,omitempty"` // 详细信息
}

// Error 实现error接口
func (e *BusinessError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("code: %d, message: %s, details: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// NewBusinessError 创建业务错误
func NewBusinessError(code int, details ...string) *BusinessError {
	message := errorMessages[code]
	if message == "" {
		message = "未知错误"
	}

	httpStatus := codeToHTTPStatus[code]
	if httpStatus == 0 {
		httpStatus = http.StatusInternalServerError
	}

	err := &BusinessError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}

	if len(details) > 0 {
		err.Details = details[0]
	}

	return err
}

// NewBusinessErrorWithMessage 创建带自定义消息的业务错误
func NewBusinessErrorWithMessage(code int, message string, details ...string) *BusinessError {
	httpStatus := codeToHTTPStatus[code]
	if httpStatus == 0 {
		httpStatus = http.StatusInternalServerError
	}

	err := &BusinessError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}

	if len(details) > 0 {
		err.Details = details[0]
	}

	return err
}

// 便捷错误创建函数

// ErrUserNotFound 用户不存在错误
func ErrUserNotFound() *BusinessError {
	return NewBusinessError(CodeUserNotFound)
}

// ErrUserAlreadyExists 用户已存在错误
func ErrUserAlreadyExists() *BusinessError {
	return NewBusinessError(CodeUserAlreadyExists)
}

// ErrUserEmailExists 邮箱已存在错误
func ErrUserEmailExists() *BusinessError {
	return NewBusinessError(CodeUserEmailExists)
}

// ErrUserUsernameExists 用户名已存在错误
func ErrUserUsernameExists() *BusinessError {
	return NewBusinessError(CodeUserUsernameExists)
}

// ErrInvalidCredentials 凭据无效错误
func ErrInvalidCredentials() *BusinessError {
	return NewBusinessError(CodeInvalidCredentials)
}

// ErrValidationFailed 验证失败错误
func ErrValidationFailed(details ...string) *BusinessError {
	if len(details) > 0 {
		return NewBusinessError(CodeValidationError, details[0])
	}
	return NewBusinessError(CodeValidationError)
}

// ErrDuplicateEntry 重复记录错误
func ErrDuplicateEntry(details string) *BusinessError {
	return NewBusinessError(CodeDuplicateEntry, details)
}

// ErrInternalError 内部错误
func ErrInternalError(details string) *BusinessError {
	return NewBusinessError(CodeInternalError, details)
}

// ErrUnauthorized 未授权错误
func ErrUnauthorized() *BusinessError {
	return NewBusinessError(CodeUnauthorized)
}

// ErrForbidden 禁止访问错误
func ErrForbidden() *BusinessError {
	return NewBusinessError(CodeForbidden)
}

// ErrRateLimit 频率限制错误
func ErrRateLimit() *BusinessError {
	return NewBusinessError(CodeRateLimitError)
}

// ErrRateLimitExceeded 请求频率超限错误
func ErrRateLimitExceeded() *BusinessError {
	return NewBusinessError(CodeAPIRateLimitExceeded)
}

// ErrCaptchaGenerate 验证码生成失败错误
func ErrCaptchaGenerate() *BusinessError {
	return NewBusinessError(CodeCaptchaGenerate)
}

// ErrCaptchaInvalid 验证码无效错误
func ErrCaptchaInvalid() *BusinessError {
	return NewBusinessError(CodeCaptchaInvalid)
}

// ErrCaptchaExpired 验证码已过期错误
func ErrCaptchaExpired() *BusinessError {
	return NewBusinessError(CodeCaptchaExpired)
}

// ErrCaptchaRequired 需要验证码错误
func ErrCaptchaRequired() *BusinessError {
	return NewBusinessError(CodeCaptchaRequired)
}

// ErrUserInactive 用户未激活错误
func ErrUserInactive() *BusinessError {
	return NewBusinessError(CodeUserInactive)
}

// ErrInvalidToken 无效token错误
func ErrInvalidToken() *BusinessError {
	return NewBusinessError(CodeUnauthorized, "invalid token")
}

// ErrUserPermissionError 用户权限不足错误
func ErrUserPermissionError() *BusinessError {
	return NewBusinessError(CodeUserPermissionError)
}

// ErrInvalidRequest 无效请求错误
func ErrInvalidRequest() *BusinessError {
	return NewBusinessError(CodeInvalidRequest)
}
