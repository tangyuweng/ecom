package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tangyuweng/ecom/conf"
	"github.com/tangyuweng/ecom/internal/application/dto"
	"github.com/tangyuweng/ecom/internal/application/usecase"
)

type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
	config      *conf.Config
}

func NewAuthHandler(authUseCase *usecase.AuthUseCase, cfg *conf.Config) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		config:      cfg,
	}
}

// Register godoc
// @Summary      註冊新用戶
// @Description  創建新的用戶帳號
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "註冊資訊"
// @Success      200 {object} dto.StandardResponse{data=dto.RegisterResponse} "註冊成功"
// @Failure      400 {object} dto.StandardResponse "請求參數錯誤"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	response, err := h.authUseCase.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}

// Login godoc
// @Summary      用戶登入
// @Description  使用 email 和密碼登入
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "登入資訊"
// @Success      200 {object} dto.StandardResponse{data=dto.LoginResponse} "登入成功"
// @Failure      400 {object} dto.StandardResponse "請求參數錯誤"
// @Failure      401 {object} dto.StandardResponse "認證失敗"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	response, err := h.authUseCase.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse(err.Error()))
		return
	}

	c.SetCookie(
		"refresh_token",
		response.RefreshToken,
		h.config.JWT.RefreshTokenExpiry*60*60,
		"/",
		h.config.Cookie.Domain,
		h.config.Cookie.Secure,
		true,
	)

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}

// RefreshToken godoc
// @Summary      刷新訪問令牌
// @Description  使用 refresh token 獲取新的 access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200 {object} dto.StandardResponse{data=dto.RefreshTokenResponse} "刷新成功"
// @Failure      401 {object} dto.StandardResponse "無效的 refresh token"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("refresh token not found"))
		return
	}

	response, err := h.authUseCase.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}

// Logout godoc
// @Summary      用戶登出
// @Description  清除 refresh token cookie
// @Tags         auth
// @Produce      json
// @Success      200 {object} dto.StandardResponse "登出成功"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		h.config.Cookie.Domain,
		h.config.Cookie.Secure,
		true,
	)

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"message": "logged out successfully"}))
}
