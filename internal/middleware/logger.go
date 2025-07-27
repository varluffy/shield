// Package middleware contains HTTP middleware for the web server.
// It provides common functionality like logging, CORS, recovery, and request tracing.
package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/varluffy/shield/pkg/logger"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const (
	// MaxRequestBodySize 请求体日志记录的最大大小 (1MB)
	MaxRequestBodySize = 1024 * 1024
	// MaxResponseBodySize 响应体日志记录的最大大小 (1MB)
	MaxResponseBodySize = 1024 * 1024
	// MaxLogBodySize 日志中显示的body最大大小 (10KB)
	MaxLogBodySize = 10 * 1024
)

// responseBodyWriter 响应体写入器
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 写入响应体
func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		ctx := param.Request.Context()

		fields := []zap.Field{
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.String("query", param.Request.URL.RawQuery),
			zap.String("ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("error", param.ErrorMessage),
		}

		// 添加追踪信息
		span := trace.SpanFromContext(ctx)
		if span.IsRecording() {
			spanContext := span.SpanContext()
			fields = append(fields,
				zap.String("trace_id", spanContext.TraceID().String()),
				zap.String("span_id", spanContext.SpanID().String()),
			)
		}

		// 根据状态码选择日志级别
		if param.StatusCode >= 400 {
			if param.StatusCode >= 500 {
				logger.ErrorWithTrace(ctx, "HTTP request completed with server error", fields...)
			} else {
				logger.WarnWithTrace(ctx, "HTTP request completed with client error", fields...)
			}
		} else {
			logger.InfoWithTrace(ctx, "HTTP request completed", fields...)
		}

		return ""
	})
}

// EnhancedLoggerMiddleware 增强型日志中间件，支持记录请求体和响应体
func EnhancedLoggerMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 记录请求开始
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("content_type", c.Request.Header.Get("Content-Type")),
			zap.Int64("content_length", c.Request.ContentLength),
		}

		// 读取请求体
		requestBody := readRequestBody(c, logger)
		if requestBody != "" {
			fields = append(fields, zap.String("request_body", requestBody))
		}

		// 添加追踪信息
		ctx := c.Request.Context()
		span := trace.SpanFromContext(ctx)
		traceID := ""
		if span.IsRecording() {
			spanContext := span.SpanContext()
			traceID = spanContext.TraceID().String()
			fields = append(fields,
				zap.String("trace_id", traceID),
				zap.String("span_id", spanContext.SpanID().String()),
			)
		}

		logger.InfoWithTrace(ctx, "HTTP request started", fields...)

		// 创建响应体写入器
		bodyWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = bodyWriter

		// 处理请求
		c.Next()

		// 记录请求完成日志
		end := time.Now()
		latency := end.Sub(start)
		status := c.Writer.Status()

		// 读取响应体
		responseBody := readResponseBody(bodyWriter.body)

		responseFields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.Int("response_size", bodyWriter.body.Len()),
		}

		// 添加响应体
		if responseBody != "" {
			responseFields = append(responseFields, zap.String("response_body", responseBody))
		}

		// 添加错误信息
		if len(c.Errors) > 0 {
			responseFields = append(responseFields, zap.String("errors", c.Errors.String()))
		}

		// 添加追踪信息
		if traceID != "" {
			responseFields = append(responseFields, zap.String("trace_id", traceID))
		}

		// 根据状态码选择日志级别
		if status >= 500 {
			logger.ErrorWithTrace(ctx, "HTTP request completed with server error", responseFields...)
		} else if status >= 400 {
			logger.WarnWithTrace(ctx, "HTTP request completed with client error", responseFields...)
		} else {
			logger.InfoWithTrace(ctx, "HTTP request completed successfully", responseFields...)
		}
	}
}

// readRequestBody 读取请求体
func readRequestBody(c *gin.Context, logger *logger.Logger) string {
	// 只记录POST, PUT, PATCH请求的body
	if c.Request.Method != http.MethodPost &&
		c.Request.Method != http.MethodPut &&
		c.Request.Method != http.MethodPatch {
		return ""
	}

	// 检查Content-Type，跳过multipart/form-data
	contentType := c.Request.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		return "[multipart/form-data - not logged]"
	}

	// 检查内容长度
	if c.Request.ContentLength > MaxRequestBodySize {
		return "[request body too large - not logged]"
	}

	// 读取body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.WarnWithTrace(c.Request.Context(), "Failed to read request body", zap.Error(err))
		return "[failed to read request body]"
	}

	// 恢复body供后续使用
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// 如果body为空，返回空字符串
	if len(body) == 0 {
		return ""
	}

	// 如果body太大，只记录摘要
	if len(body) > MaxLogBodySize {
		return fmt.Sprintf("[request body size: %d bytes - truncated for logging]", len(body))
	}

	// 尝试格式化JSON
	if strings.Contains(contentType, "application/json") {
		var jsonObj interface{}
		if err := json.Unmarshal(body, &jsonObj); err == nil {
			if formatted, err := json.Marshal(jsonObj); err == nil {
				return string(formatted)
			}
		}
	}

	return string(body)
}

// readResponseBody 读取响应体
func readResponseBody(body *bytes.Buffer) string {
	if body.Len() == 0 {
		return ""
	}

	// 如果响应体太大，只记录摘要
	if body.Len() > MaxLogBodySize {
		return fmt.Sprintf("[response body size: %d bytes - truncated for logging]", body.Len())
	}

	bodyBytes := body.Bytes()

	// 尝试格式化JSON响应
	var jsonObj interface{}
	if err := json.Unmarshal(bodyBytes, &jsonObj); err == nil {
		if formatted, err := json.Marshal(jsonObj); err == nil {
			return string(formatted)
		}
	}

	return string(bodyBytes)
}

// StructuredLoggerMiddleware 结构化日志中间件
func StructuredLoggerMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 记录请求完成日志
		end := time.Now()
		latency := end.Sub(start)

		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.Int("body_size", c.Writer.Size()),
		}

		// 添加错误信息
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		// 添加追踪信息
		ctx := c.Request.Context()
		span := trace.SpanFromContext(ctx)
		if span.IsRecording() {
			spanContext := span.SpanContext()
			fields = append(fields,
				zap.String("trace_id", spanContext.TraceID().String()),
				zap.String("span_id", spanContext.SpanID().String()),
			)
		}

		// 根据状态码选择日志级别
		if c.Writer.Status() >= 500 {
			logger.ErrorWithTrace(ctx, "HTTP server error", fields...)
		} else if c.Writer.Status() >= 400 {
			logger.WarnWithTrace(ctx, "HTTP client error", fields...)
		} else {
			logger.InfoWithTrace(ctx, "HTTP request completed", fields...)
		}
	}
}
