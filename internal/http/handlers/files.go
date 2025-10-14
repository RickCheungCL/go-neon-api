package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rick/go-neon-api/internal/db"
	"github.com/rick/go-neon-api/internal/models"
)

type AddFileReq struct {
	URL         string  `json:"url" binding:"required"`
	FileName    *string `json:"fileName"` // for Document
	CustomName  *string `json:"customName"`
	Comment     *string `json:"comment"` // for Photo
	UploadedVia *bool   `json:"uploadedViaLink"`
}

func (h *Handlers) AddPhoto(c *gin.Context) {
	caseID := c.Param("id")
	var req AddFileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item := models.Photo{
		URL:             req.URL,
		CaseID:          caseID,
		UploadedViaLink: req.UploadedVia != nil && *req.UploadedVia,
		CustomName:      req.CustomName,
	}
	if req.Comment != nil {
		item.Comment = req.Comment
	}
	if err := db.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handlers) AddDocument(c *gin.Context) {
	caseID := c.Param("id")
	var req AddFileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.FileName == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fileName is required"})
		return
	}
	item := models.Document{
		URL:             req.URL,
		FileName:        *req.FileName,
		CaseID:          caseID,
		UploadedViaLink: req.UploadedVia != nil && *req.UploadedVia,
		CustomName:      req.CustomName,
	}
	if err := db.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}
	c.JSON(http.StatusCreated, item)
}
