package response

import "github.com/gin-gonic/gin"

// Envelope is the standard JSON shape every endpoint returns.
type Envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func OK(c *gin.Context, status int, message string, data any) {
	c.JSON(status, Envelope{Success: true, Message: message, Data: data})
}

func Fail(c *gin.Context, status int, message string) {
	c.JSON(status, Envelope{Success: false, Message: message})
}
