// Package httpclient provides a unified HTTP client with logging and tracing support.
package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/varluffy/shield/internal/config"
	"github.com/varluffy/shield/pkg/logger"
	"github.com/varluffy/shield/pkg/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

//go:generate mockgen -source=client.go -destination=mocks/client_mock.go

// HTTPClient 统一HTTP客户端接口
type HTTPClient interface {
	// Get 发送GET请求
	Get(ctx context.Context, url string) (*Response, error)

	// Post 发送POST请求
	Post(ctx context.Context, url string, body interface{}) (*Response, error)

	// Put 发送PUT请求
	Put(ctx context.Context, url string, body interface{}) (*Response, error)

	// Delete 发送DELETE请求
	Delete(ctx context.Context, url string) (*Response, error)

	// Patch 发送PATCH请求
	Patch(ctx context.Context, url string, body interface{}) (*Response, error)

	// Request 发送自定义请求
	Request(ctx context.Context, method, url string, body interface{}, headers map[string]string) (*Response, error)

	// SetBaseURL 设置基础URL
	SetBaseURL(baseURL string) HTTPClient

	// SetHeader 设置默认请求头
	SetHeader(key, value string) HTTPClient

	// SetTimeout 设置请求超时时间
	SetTimeout(timeout time.Duration) HTTPClient
}

// Response 统一响应结构
type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string
	IsSuccess  bool
	Error      error
	TraceID    string
}

// HTTPClientImpl HTTP客户端实现
type HTTPClientImpl struct {
	client *resty.Client
	config *config.HTTPClientConfig
	logger *logger.Logger
	tracer trace.Tracer
}

// NewHTTPClient 创建HTTP客户端
func NewHTTPClient(cfg *config.HTTPClientConfig, logger *logger.Logger) HTTPClient {
	client := resty.New()

	// 基础配置
	client.SetTimeout(time.Duration(cfg.Timeout) * time.Second)
	client.SetRetryCount(cfg.RetryCount)
	client.SetRetryWaitTime(time.Duration(cfg.RetryWaitTime) * time.Second)
	client.SetRetryMaxWaitTime(time.Duration(cfg.RetryMaxWaitTime) * time.Second)
	client.SetHeader("User-Agent", cfg.UserAgent)

	// 创建客户端实例
	httpClient := &HTTPClientImpl{
		client: client,
		config: cfg,
		logger: logger,
		tracer: otel.Tracer("ultrafit-http-client"),
	}

	// 设置中间件
	if cfg.EnableLog {
		client.OnBeforeRequest(httpClient.beforeRequestMiddleware)
		client.OnAfterResponse(httpClient.afterResponseMiddleware)
	}

	if cfg.EnableTrace {
		client.OnBeforeRequest(httpClient.tracingMiddleware)
	}

	// 设置重试条件
	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		return r.StatusCode() >= 500 || err != nil
	})

	return httpClient
}

// Get 发送GET请求
func (h *HTTPClientImpl) Get(ctx context.Context, url string) (*Response, error) {
	return h.Request(ctx, "GET", url, nil, nil)
}

// Post 发送POST请求
func (h *HTTPClientImpl) Post(ctx context.Context, url string, body interface{}) (*Response, error) {
	return h.Request(ctx, "POST", url, body, nil)
}

// Put 发送PUT请求
func (h *HTTPClientImpl) Put(ctx context.Context, url string, body interface{}) (*Response, error) {
	return h.Request(ctx, "PUT", url, body, nil)
}

// Delete 发送DELETE请求
func (h *HTTPClientImpl) Delete(ctx context.Context, url string) (*Response, error) {
	return h.Request(ctx, "DELETE", url, nil, nil)
}

// Patch 发送PATCH请求
func (h *HTTPClientImpl) Patch(ctx context.Context, url string, body interface{}) (*Response, error) {
	return h.Request(ctx, "PATCH", url, body, nil)
}

// Request 发送自定义请求
func (h *HTTPClientImpl) Request(ctx context.Context, method, url string, body interface{}, headers map[string]string) (*Response, error) {
	var span trace.Span
	if h.config.EnableTrace {
		ctx, span = h.tracer.Start(ctx, fmt.Sprintf("HTTP %s %s", method, url))
		defer span.End()
	}

	req := h.client.R().SetContext(ctx)

	// 设置请求头
	if headers != nil {
		req.SetHeaders(headers)
	}

	// 自动注入TraceID
	if traceID := tracing.GetTraceIDFromContext(ctx); traceID != "" {
		req.SetHeader("X-Trace-ID", traceID)
	}

	// 设置请求体
	if body != nil {
		req.SetBody(body)
		req.SetHeader("Content-Type", "application/json")
	}

	// 发送请求
	var resp *resty.Response
	var err error

	switch method {
	case "GET":
		resp, err = req.Get(url)
	case "POST":
		resp, err = req.Post(url)
	case "PUT":
		resp, err = req.Put(url)
	case "DELETE":
		resp, err = req.Delete(url)
	case "PATCH":
		resp, err = req.Patch(url)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	// 构造响应
	response := &Response{
		Headers: resp.Header(),
		TraceID: tracing.GetTraceIDFromContext(ctx),
	}

	if err != nil {
		response.Error = err
		response.IsSuccess = false
		if span != nil {
			span.RecordError(err)
		}
		return response, err
	}

	response.StatusCode = resp.StatusCode()
	response.Body = resp.Body()
	response.IsSuccess = resp.IsSuccess()

	// 记录span属性
	if span != nil {
		span.SetAttributes(
			attribute.String("http.method", method),
			attribute.String("http.url", url),
			attribute.Int("http.status_code", resp.StatusCode()),
			attribute.Int("http.response_size", len(resp.Body())),
		)
	}

	return response, nil
}

// SetBaseURL 设置基础URL
func (h *HTTPClientImpl) SetBaseURL(baseURL string) HTTPClient {
	h.client.SetHostURL(baseURL)
	return h
}

// SetHeader 设置默认请求头
func (h *HTTPClientImpl) SetHeader(key, value string) HTTPClient {
	h.client.SetHeader(key, value)
	return h
}

// SetTimeout 设置请求超时时间
func (h *HTTPClientImpl) SetTimeout(timeout time.Duration) HTTPClient {
	h.client.SetTimeout(timeout)
	return h
}

// UnmarshalJSON 便捷方法：将响应体解析为JSON
func (r *Response) UnmarshalJSON(v interface{}) error {
	if !r.IsSuccess {
		return fmt.Errorf("response not successful: status %d", r.StatusCode)
	}
	return json.Unmarshal(r.Body, v)
}

// String 便捷方法：获取响应体字符串
func (r *Response) String() string {
	return string(r.Body)
}

// beforeRequestMiddleware 请求前中间件
func (h *HTTPClientImpl) beforeRequestMiddleware(c *resty.Client, req *resty.Request) error {
	ctx := req.Context()

	// 获取TraceID
	traceID := tracing.GetTraceIDFromContext(ctx)

	// 记录请求日志
	fields := []zap.Field{
		zap.String("method", req.Method),
		zap.String("url", req.URL),
		zap.Any("headers", req.Header),
	}

	if traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	// 记录请求体（限制大小）
	if req.Body != nil {
		bodyStr := h.formatBody(req.Body)
		if len(bodyStr) > h.config.MaxLogBodySize {
			bodyStr = bodyStr[:h.config.MaxLogBodySize] + "...[truncated]"
		}
		fields = append(fields, zap.String("request_body", bodyStr))
	}

	h.logger.InfoWithTrace(ctx, "HTTP request started", fields...)
	return nil
}

// afterResponseMiddleware 响应后中间件
func (h *HTTPClientImpl) afterResponseMiddleware(c *resty.Client, resp *resty.Response) error {
	ctx := resp.Request.Context()

	// 获取TraceID
	traceID := tracing.GetTraceIDFromContext(ctx)

	// 记录响应日志
	fields := []zap.Field{
		zap.String("method", resp.Request.Method),
		zap.String("url", resp.Request.URL),
		zap.Int("status_code", resp.StatusCode()),
		zap.Duration("duration", resp.Time()),
		zap.Int("response_size", len(resp.Body())),
	}

	if traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	// 记录响应体（限制大小）
	if len(resp.Body()) > 0 {
		bodyStr := string(resp.Body())
		if len(bodyStr) > h.config.MaxLogBodySize {
			bodyStr = bodyStr[:h.config.MaxLogBodySize] + "...[truncated]"
		}
		fields = append(fields, zap.String("response_body", bodyStr))
	}

	// 根据状态码选择日志级别
	if resp.IsSuccess() {
		h.logger.InfoWithTrace(ctx, "HTTP request completed successfully", fields...)
	} else {
		h.logger.WarnWithTrace(ctx, "HTTP request completed with error", fields...)
	}

	return nil
}

// tracingMiddleware 链路追踪中间件
func (h *HTTPClientImpl) tracingMiddleware(c *resty.Client, req *resty.Request) error {
	ctx := req.Context()

	// 自动注入TraceID到请求头
	if traceID := tracing.GetTraceIDFromContext(ctx); traceID != "" {
		req.SetHeader("X-Trace-ID", traceID)

		// 也可以注入其他OpenTelemetry标准头
		if span := trace.SpanFromContext(ctx); span != nil {
			spanContext := span.SpanContext()
			if spanContext.IsValid() {
				req.SetHeader("traceparent", fmt.Sprintf("00-%s-%s-01",
					spanContext.TraceID().String(),
					spanContext.SpanID().String()))
			}
		}
	}

	return nil
}

// formatBody 格式化请求体用于日志记录
func (h *HTTPClientImpl) formatBody(body interface{}) string {
	switch v := body.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		if jsonData, err := json.Marshal(v); err == nil {
			return string(jsonData)
		}
		return fmt.Sprintf("%+v", v)
	}
}
