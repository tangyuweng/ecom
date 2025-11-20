package dto

type UpdateUserRoleRequest struct {
	UserRole string `json:"user_role" binding:"required,oneof=user admin"`
}

type UpdateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone" binding:"required"`
}

type UpdateUserPasswordRequest struct {
	Password    string `json:"password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Role  string `json:"role"`
}

type UserListResponse struct {
	Total int             `json:"total"`
	Users []*UserResponse `json:"users"`
}
