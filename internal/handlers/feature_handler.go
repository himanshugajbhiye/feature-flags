package handlers

import (
	"feature-flags/internal/models"
	"feature-flags/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FeatureHandler struct {
	featureService *services.FeatureService
}

func NewFeatureHandler(featureService *services.FeatureService) *FeatureHandler {
	return &FeatureHandler{
		featureService: featureService,
	}
}

type CreateFeatureRequest struct {
	Name      string             `json:"name" binding:"required"`
	Type      models.FeatureType `json:"type" binding:"required"`
	IsEnabled bool               `json:"is_enabled"`
}

type AddDependencyRequest struct {
	ParentID string `json:"parent_id" binding:"required"`
	ChildID  string `json:"child_id" binding:"required"`
}

func (h *FeatureHandler) CreateFeature(c *gin.Context) {
	var req CreateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	feature := &models.Feature{
		Name:      req.Name,
		Type:      req.Type,
		IsEnabled: req.IsEnabled,
	}

	if err := h.featureService.CreateFeature(c.Request.Context(), feature); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, feature)
}

func (h *FeatureHandler) AddDependency(c *gin.Context) {
	var req AddDependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parentID, err := primitive.ObjectIDFromHex(req.ParentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent_id"})
		return
	}

	childID, err := primitive.ObjectIDFromHex(req.ChildID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid child_id"})
		return
	}

	if err := h.featureService.AddChild(c.Request.Context(), parentID, childID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "dependency added successfully"})
}

func (h *FeatureHandler) EnableFeature(c *gin.Context) {
	id := c.Param("id")
	featureID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feature id"})
		return
	}

	if err := h.featureService.EnableFeature(c.Request.Context(), featureID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "feature enabled successfully"})
}

func (h *FeatureHandler) DisableFeature(c *gin.Context) {
	id := c.Param("id")
	featureID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feature id"})
		return
	}

	if err := h.featureService.DisableFeature(c.Request.Context(), featureID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "feature disabled successfully"})
}

func (h *FeatureHandler) GetFeatureStatus(c *gin.Context) {
	id := c.Param("id")
	featureID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feature id"})
		return
	}

	feature, err := h.featureService.GetFeatureStatus(c.Request.Context(), featureID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "feature not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, feature)
}
