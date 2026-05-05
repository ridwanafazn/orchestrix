package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ridwanafazn/orchestrix/internal/usecase"
)

type JobHandler struct {
	jobUsecase usecase.JobUsecase
}

func NewJobHandler(ju usecase.JobUsecase) *JobHandler {
	return &JobHandler{jobUsecase: ju}
}

func (h *JobHandler) HandleUpload(c *gin.Context) {
	// Simulasi UserID (Nanti ini diganti dengan ekstraksi dari JWT Token)
	userID := c.DefaultPostForm("user_id", "guest-12345")

	// Ambil file dari request multipart
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// Buka file
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer file.Close()

	// Eksekusi Usecase
	contentType := fileHeader.Header.Get("Content-Type")
	job, err := h.jobUsecase.ProcessKYCUpload(c.Request.Context(), userID, fileHeader.Filename, file, contentType)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 202 Accepted: Diterima untuk diproses (Asynchronous)
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Upload accepted. Job is currently being processed.",
		"job_id":  job.ID,
		"status":  job.Status,
	})
}
