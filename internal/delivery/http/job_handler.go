package http

import (
	"context"
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
	userID := c.MustGet("user_id").(string)

	// PENYUNTIKAN SKENARIO: Menerima sinyal simulasi dari Frontend via Query Param
	chaosInject := c.Query("force_error") == "true"
	heavyCompute := c.Query("heavy") == "true"

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")

	// Suntikkan kedua flag ke dalam context
	ctx := context.WithValue(c.Request.Context(), "chaos_inject", chaosInject)
	ctx = context.WithValue(ctx, "heavy_compute", heavyCompute)

	job, err := h.jobUsecase.ProcessKYCUpload(ctx, userID, fileHeader.Filename, file, contentType)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"message": "Upload accepted. Job is currently being processed.",
		"job_id":  job.ID,
		"status":  job.Status,
	}

	if chaosInject {
		response["warning"] = "Chaos Engineering Triggered. Node failure expected."
	}
	if heavyCompute {
		response["info"] = "Heavy Computation Triggered. Processing will take longer."
	}

	c.JSON(http.StatusAccepted, response)
}
