package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/varluffy/shield/internal/dto"
	"github.com/varluffy/shield/pkg/captcha"
	"github.com/varluffy/shield/pkg/errors"
	"github.com/varluffy/shield/pkg/response"
	"go.uber.org/zap"
)

// CaptchaHandler handles captcha-related HTTP requests
type CaptchaHandler struct {
	captchaService captcha.CaptchaService
	responseWriter *response.ResponseWriter
	logger         *zap.Logger
}

// NewCaptchaHandler creates a new captcha handler
func NewCaptchaHandler(captchaService captcha.CaptchaService, responseWriter *response.ResponseWriter, logger *zap.Logger) *CaptchaHandler {
	return &CaptchaHandler{
		captchaService: captchaService,
		responseWriter: responseWriter,
		logger:         logger,
	}
}

// GenerateCaptcha generates a new captcha
// @Summary Generate captcha
// @Description Generate a new captcha image
// @Tags Captcha
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=dto.CaptchaGenerateResponse} "Success"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /captcha/generate [get]
func (h *CaptchaHandler) GenerateCaptcha(c *gin.Context) {
	ctx := c.Request.Context()

	// 生成验证码
	captchaResp, err := h.captchaService.GenerateCaptcha(ctx)
	if err != nil {
		h.logger.Error("Failed to generate captcha",
			zap.Error(err))

		h.responseWriter.Error(c, errors.ErrCaptchaGenerate())
		return
	}

	// 转换为DTO
	result := &dto.CaptchaGenerateResponse{
		CaptchaID:    captchaResp.CaptchaID,
		CaptchaImage: captchaResp.CaptchaImage,
	}

	h.logger.Info("Captcha generated successfully",
		zap.String("captcha_id", result.CaptchaID))

	h.responseWriter.Success(c, result)
}

// VerifyCaptcha verifies a captcha answer
// @Summary Verify captcha
// @Description Verify captcha answer
// @Tags Captcha
// @Accept json
// @Produce json
// @Param request body dto.CaptchaVerifyRequest true "Captcha verify request"
// @Success 200 {object} response.Response "Success"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /captcha/verify [post]
func (h *CaptchaHandler) VerifyCaptcha(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.CaptchaVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid captcha verify request",
			zap.Error(err))

		h.responseWriter.ValidationError(c, err)
		return
	}

	// 验证验证码
	err := h.captchaService.VerifyCaptcha(ctx, req.CaptchaID, req.Answer)
	if err != nil {
		h.logger.Warn("Captcha verification failed",
			zap.Error(err),
			zap.String("captcha_id", req.CaptchaID))

		h.responseWriter.Error(c, errors.ErrCaptchaInvalid())
		return
	}

	h.logger.Info("Captcha verified successfully",
		zap.String("captcha_id", req.CaptchaID))

	h.responseWriter.Success(c, nil)
}
