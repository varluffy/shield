// Package response provides standardized HTTP response utilities.
// It offers unified response formats and helper functions for API responses.
package response

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/varluffy/shield/pkg/errors"
	"github.com/varluffy/shield/pkg/logger"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Response 统一响应结构
type Response struct {
	Code      int         `json:"code"`               // 业务错误码
	Message   string      `json:"message"`            // 响应消息
	Data      interface{} `json:"data,omitempty"`     // 响应数据
	TraceID   string      `json:"trace_id,omitempty"` // 追踪ID
	Timestamp string      `json:"timestamp"`          // 响应时间戳 (RFC3339格式)
}

// PaginationMeta 分页元数据
type PaginationMeta struct {
	Page       int   `json:"page"`        // 当前页
	Limit      int   `json:"limit"`       // 每页数量
	Total      int64 `json:"total"`       // 总数
	TotalPages int   `json:"total_pages"` // 总页数
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	*Response
	Meta *PaginationMeta `json:"meta"`
}

// ResponseWriter 响应写入器
type ResponseWriter struct {
	logger *logger.Logger
}

// NewResponseWriter 创建响应写入器
func NewResponseWriter(logger *logger.Logger) *ResponseWriter {
	return &ResponseWriter{
		logger: logger,
	}
}

// Success 成功响应
func (w *ResponseWriter) Success(c *gin.Context, data interface{}) {
	w.writeResponse(c, http.StatusOK, errors.CodeSuccess, "success", data)
}

// Created 创建成功响应
func (w *ResponseWriter) Created(c *gin.Context, data interface{}) {
	w.writeResponse(c, http.StatusCreated, errors.CodeSuccess, "created", data)
}

// NoContent 无内容响应
func (w *ResponseWriter) NoContent(c *gin.Context) {
	w.writeResponse(c, http.StatusNoContent, errors.CodeSuccess, "no content", nil)
}

// Error 错误响应
func (w *ResponseWriter) Error(c *gin.Context, err error) {
	if bizErr, ok := err.(*errors.BusinessError); ok {
		w.writeResponse(c, bizErr.HTTPStatus, bizErr.Code, bizErr.Message, nil)
	} else {
		w.writeResponse(c, http.StatusInternalServerError, errors.CodeInternalError, "internal server error", nil)
	}
}

// BadRequest 请求错误响应
func (w *ResponseWriter) BadRequest(c *gin.Context, message string) {
	w.writeResponse(c, http.StatusBadRequest, errors.CodeInvalidRequest, message, nil)
}

// Unauthorized 未授权响应
func (w *ResponseWriter) Unauthorized(c *gin.Context, message string) {
	w.writeResponse(c, http.StatusUnauthorized, errors.CodeUnauthorized, message, nil)
}

// Forbidden 禁止访问响应
func (w *ResponseWriter) Forbidden(c *gin.Context, message string) {
	w.writeResponse(c, http.StatusForbidden, errors.CodeForbidden, message, nil)
}

// NotFound 未找到响应
func (w *ResponseWriter) NotFound(c *gin.Context, message string) {
	w.writeResponse(c, http.StatusNotFound, errors.CodeNotFound, message, nil)
}

// Conflict 冲突响应
func (w *ResponseWriter) Conflict(c *gin.Context, message string) {
	w.writeResponse(c, http.StatusConflict, errors.CodeConflict, message, nil)
}

// ValidationError 验证错误响应
// 支持字符串消息或binding错误，会自动进行多语言翻译
func (w *ResponseWriter) ValidationError(c *gin.Context, errOrMessage interface{}) {
	var message string

	switch v := errOrMessage.(type) {
	case string:
		// 直接使用提供的消息
		message = v
	case error:
		// 尝试翻译验证错误
		if globalValidationErrorTranslator != nil {
			if errorMessages := globalValidationErrorTranslator(v); len(errorMessages) > 0 {
				message = strings.Join(errorMessages, "; ")
			} else {
				message = "参数验证失败: " + v.Error()
			}
		} else {
			// 如果没有翻译器，使用原始错误信息
			message = "参数验证失败: " + v.Error()
		}
	default:
		message = "参数验证失败"
	}

	w.writeResponse(c, http.StatusBadRequest, errors.CodeValidationError, message, nil)
}

// Pagination 分页响应
func (w *ResponseWriter) Pagination(c *gin.Context, data interface{}, meta *PaginationMeta) {
	resp := &PaginationResponse{
		Response: &Response{
			Code:      errors.CodeSuccess,
			Message:   "success",
			Data:      data,
			TraceID:   w.getTraceID(c),
			Timestamp: time.Now().Format(time.RFC3339),
		},
		Meta: meta,
	}

	// 记录响应日志
	w.logResponse(c, http.StatusOK, resp)
	c.JSON(http.StatusOK, resp)
}

// writeResponse 写入响应
func (w *ResponseWriter) writeResponse(c *gin.Context, httpStatus, code int, message string, data interface{}) {
	resp := &Response{
		Code:      code,
		Message:   message,
		Data:      data,
		TraceID:   w.getTraceID(c),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// 记录响应日志
	w.logResponse(c, httpStatus, resp)
	c.JSON(httpStatus, resp)
}

// getTraceID 获取追踪ID
func (w *ResponseWriter) getTraceID(c *gin.Context) string {
	// 防止在测试环境中Request为nil的情况
	if c.Request == nil {
		return ""
	}

	span := trace.SpanFromContext(c.Request.Context())
	if span.IsRecording() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// logResponse 记录响应日志
func (w *ResponseWriter) logResponse(c *gin.Context, httpStatus int, resp interface{}) {
	// 防止在测试环境中Request为nil的情况
	if c.Request == nil {
		return
	}

	fields := []zap.Field{
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Int("http_status", httpStatus),
		zap.Any("response", resp),
	}

	// 添加追踪信息
	if traceID := w.getTraceID(c); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	if httpStatus >= 400 {
		w.logger.WarnWithTrace(c.Request.Context(), "API response with error", fields...)
	} else {
		w.logger.InfoWithTrace(c.Request.Context(), "API response", fields...)
	}
}
