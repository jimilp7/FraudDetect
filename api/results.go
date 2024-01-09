package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// AnalysisResult represents the result of an analysis
type AnalysisResult struct {
	Result string `json:"result"`
}

// getAnalysisResults godoc
// @Summary Get analysis results
// @Description Retrieves the results of the completed analysis.
// @Tags transactions
// @Produce json
// @Param analysisId path string true "Analysis ID"
// @Success 200 {object} AnalysisResult
// @Router /results/{analysisId} [get]
func getAnalysisResults(c *gin.Context) {
	analysisID := c.Param("analysisId")
	// Retrieve the engine instance
	EnginesMutex.Lock()
	engine, exists := Engines[analysisID]
	EnginesMutex.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Analysis not found"})
		return
	}

	// Return the final analysis result
	c.JSON(200, gin.H{
		"result": engine.AnalysisResult,
	})
}
