package services

import (
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"log"
	"os"
	"sync"
)

func StringPointer(s string) *string {
	return &s
}

// FraudDetectionEngine is responsible for analyzing fraud in transactions.
type FraudDetectionEngine struct {
	fileID         string
	analysisID     string
	AnalysisResult string
	rules          []string
	AnalysisMap    map[string]*AnalysisStatus
	AnalysisLock   *sync.Mutex
	client         *openai.Client
}

// AnalysisStatus represents the status of an analysis.
type AnalysisStatus struct {
	Status string // "processing", "complete", "failed"
}

// NewFraudDetectionEngine creates a new instance of FraudDetectionEngine.
func NewFraudDetectionEngine(fileID, analysisID string, rules []string) *FraudDetectionEngine {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	engine := &FraudDetectionEngine{
		fileID:         fileID,
		analysisID:     analysisID,
		rules:          rules,
		AnalysisMap:    make(map[string]*AnalysisStatus),
		AnalysisLock:   &sync.Mutex{},
		AnalysisResult: "Not Available, please check status of Analysis through /analyze/:analysisId/status",
		client:         openai.NewClient(os.Getenv("OPENAI_API_KEY")),
	}

	// Initialize the analysis status to "processing"
	engine.AnalysisMap[analysisID] = &AnalysisStatus{
		Status: "processing",
	}

	return engine
}

// RunAnalysis runs the fraud analysis.
func (engine *FraudDetectionEngine) RunEngine() {
	// Implement the logic to run the analysis.
	// Use engine.fileID to locate and process the CSV file.
	// Update engine.analysisMap with the results.

	detector := NewGPTFraudDetector(engine.client, engine.fileID, engine.rules)
	result := detector.RunAnalysis()

	// Update the status and result ID
	engine.AnalysisLock.Lock()
	if status, exists := engine.AnalysisMap[engine.analysisID]; exists {
		status.Status = "complete"
		engine.AnalysisResult = result
	}

	engine.AnalysisLock.Unlock()
}
