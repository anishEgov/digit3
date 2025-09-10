package util

import (
	"tenant-management-go/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"time"
)

func GetResponseInfo(c *gin.Context, status string) models.ResponseInfo {
	return models.ResponseInfo{
		APIId:  c.FullPath(),
		Ver:    "1.0",
		Ts:     time.Now().Unix(),
		Status: status,
		MsgId:  c.GetHeader("X-Client-Id"),
	}
}

func GenerateUUID() string {
	return uuid.New().String()
} 