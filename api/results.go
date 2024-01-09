package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

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
