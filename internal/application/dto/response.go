package dto

type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func SuccessResponse(data interface{}) StandardResponse {
	return StandardResponse{
		Success: true,
		Data:    data,
	}
}

func ErrorResponse(err string) StandardResponse {
	return StandardResponse{
		Success: false,
		Error:   err,
	}
}
