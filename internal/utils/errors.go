package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func JSONError(c *gin.Context, status int, code, msg string, details interface{}) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": APIError{
			Code:    code,
			Message: msg,
			Details: details,
		},
	})
}

func BadRequest(c *gin.Context, msg string, details interface{}) {
	JSONError(c, http.StatusBadRequest, "bad_request", msg, details)
}

func Unauthorized(c *gin.Context, msg string) {
	JSONError(c, http.StatusUnauthorized, "unauthorized", msg, nil)
}

func Forbidden(c *gin.Context, msg string) {
	JSONError(c, http.StatusForbidden, "forbidden", msg, nil)
}

func NotFound(c *gin.Context, msg string) {
	JSONError(c, http.StatusNotFound, "not_found", msg, nil)
}

func Conflict(c *gin.Context, msg string) {
	JSONError(c, http.StatusConflict, "conflict", msg, nil)
}

func Internal(c *gin.Context, msg string) {
	JSONError(c, http.StatusInternalServerError, "internal_error", msg, nil)
}
