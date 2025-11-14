package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tangyuweng/ecom/internal/application/dto"
	"github.com/tangyuweng/ecom/internal/application/usecase"
	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/presentation/http/middleware"
)

type UserHandler struct {
	userUseCase *usecase.UserUseCase
}

func NewUserHandler(userUseCase *usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUseCase: userUseCase}
}

// GetUsers godoc
// @Summary      獲取所有使用者
// @Description  獲取所有使用者列表(僅管理員)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} dto.StandardResponse{data=dto.UserListResponse} "獲取成功"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      403 {object} dto.StandardResponse "權限不足"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /admin/users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	response, err := h.userUseCase.GetUsers(c.Request.Context(), userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == entity.ErrUserUnauthorized {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}

// UpdateUserRole godoc
// @Summary      更新使用者權限
// @Description  更新使用者權限 (需要管理員權限)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "使用者 ID"
// @Param        request body dto.UpdateUserRoleRequest true "權限"
// @Success      200 {object} dto.StandardResponse{data=dto.UserResponse} "更新成功"
// @Failure      400 {object} dto.StandardResponse "請求參數錯誤"
// @Failure      401 {object} dto.StandardResponse "未授權"
// @Failure      403 {object} dto.StandardResponse "無權限"
// @Failure      404 {object} dto.StandardResponse "商品不存在"
// @Failure      500 {object} dto.StandardResponse "伺服器錯誤"
// @Router       /admin/users/{id}/role [patch]
func (h *UserHandler) UpdateUserRole(c *gin.Context) {
	adminID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	userID := c.Param("id")

	var req dto.UpdateUserRoleRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	response, err := h.userUseCase.UpdateRole(c.Request.Context(), adminID, userID, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err {
		case entity.ErrUserNotFound:
			statusCode = http.StatusNotFound
		case entity.ErrUserUnauthorized:
			statusCode = http.StatusForbidden
		case entity.ErrInvalidUserRole:
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}
