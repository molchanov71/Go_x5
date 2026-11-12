package utils

type ErrorResponse struct {
	Error string `json:"error" example:"customer is required" binding:"required"`
	Code  string `json:"code" example:"VALIDATION_ERROR" binding:"required"`
}

type StatusResponse struct {
	Status string `json:"status" example:"ok" binding:"required"`
}
