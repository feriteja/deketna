package helper

import (
	"github.com/gin-gonic/gin"
)

// SuccessResponse defines the standardized structure for successful responses
type SuccessResponse struct {
	Message string      `json:"message"`        // Description of the operation
	Data    interface{} `json:"data,omitempty"` // Response data (optional)
}

// ErrorResponse defines the standardized structure for error responses
type ErrorResponse struct {
	Error ErrorDetail `json:"error"` // Error details
}

// ErrorDetail defines the structure for detailed error information
type ErrorDetail struct {
	Code    int      `json:"code" ` // HTTP status code
	Message []string `json:"message"`
}

// SendSuccess sends a standardized success response
func SendSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, SuccessResponse{
		Message: message,
		Data:    data,
	})
}

// SendError sends a standardized error response
func SendError(c *gin.Context, statusCode int, messages []string) {
	c.JSON(statusCode, ErrorResponse{
		Error: ErrorDetail{
			Code:    statusCode,
			Message: messages,
		},
	})
}
