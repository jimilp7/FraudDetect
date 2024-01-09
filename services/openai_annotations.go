package services

import (
	"fmt"
)

// Utility function to parse Annotations from a given slice of interface{} and return generated FileID by Assistant.
func ParseAnnotation(rawAnnotation interface{}) (string, error) {
	annotationMap, ok := rawAnnotation.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid annotation format")
	}

	// Check the type and unmarshal to the corresponding struct.
	annotationType, ok := annotationMap["type"].(string)
	if !ok {
		return "", fmt.Errorf("annotation type not found")
	}

	switch annotationType {
	case "file_citation": // Not supported yet
		return "", nil
	case "file_path":
		fileID, _ := annotationMap["file_path"].(map[string]interface{})["file_id"].(string)
		return fileID, nil
	default:
		return "", fmt.Errorf("unknown annotation type: %s", annotationType)
	}
}
