package services

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io"
	"os"
	"path/filepath"
	"time"
)

type GPTFraudDetector struct {
	client *openai.Client
	fileID string
	rules  []string
}

func NewGPTFraudDetector(client *openai.Client, fileID string, rules []string) *GPTFraudDetector {
	return &GPTFraudDetector{
		client: client,
		fileID: fileID,
		rules:  rules,
	}
}

func saveFileLocally(fileName string, content io.ReadCloser) error {
	// Ensure to close the content reader when the function exits
	defer content.Close()

	// Create the ResultFiles directory if it doesn't exist
	dirPath := filepath.Join(".", "ResultFiles")
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return err
	}

	// Create the file within the ResultFiles directory
	filePath := filepath.Join(dirPath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy the content to the file
	_, err = io.Copy(file, content)
	return err
}

func (detector *GPTFraudDetector) RunAnalysis(analysisID string) string {
	prompt := detector.preparePrompt()

	ctx := context.Background()

	// File upload
	engineFileId := detector.fileID
	uploadFileRequest := openai.FileRequest{
		FileName: engineFileId,
		FilePath: fmt.Sprintf("./TransactionFiles/%s.csv", engineFileId),
		Purpose:  "assistants",
	}

	uploadedFile, _ := detector.client.CreateFile(ctx, uploadFileRequest)

	openaiFileID := uploadedFile.ID

	// Create Assistant
	assistantCreateRequest := openai.AssistantRequest{
		Model:        "gpt-4-1106-preview",
		Name:         StringPointer("Fraud Detection Specialist"),
		Instructions: StringPointer("You are a Data Scientist with over 20 years experience working at Google. You are working with the Fraud Detection Team to help them in flagging fraudulent transactions. You will be provided with a CSV File containing credit card transaction data and you will have to think hard to help the team deliver the best results. You are required not to ask follow up questions and work extremely smartly through ambiguity in getting results. Given a set of rules, you should output results."),
		Tools:        []openai.AssistantTool{{Type: "code_interpreter"}},
		FileIDs:      []string{openaiFileID},
	}

	assistant, _ := detector.client.CreateAssistant(ctx, assistantCreateRequest)

	// Prepare the ThreadRequest
	threadRequest := openai.ThreadRequest{}

	// Call the CreateThread function
	thread, _ := detector.client.CreateThread(ctx, threadRequest)

	// create a message for the thread
	messageCreateRequest := openai.MessageRequest{
		Role:    "user",
		Content: prompt,
		FileIds: []string{openaiFileID},
	}

	detector.client.CreateMessage(ctx, thread.ID, messageCreateRequest)

	// Trigger the assistant
	runCreateRequest := openai.RunRequest{
		AssistantID: assistant.ID,
	}

	createRun, _ := detector.client.CreateRun(ctx, thread.ID, runCreateRequest)

	var respContent string

	for {
		time.Sleep(5 * time.Second) // Polling interval
		run, _ := detector.client.RetrieveRun(ctx, thread.ID, createRun.ID)

		if run.Status == "completed" {
			order := "desc"
			messages, _ := detector.client.ListMessage(ctx, thread.ID, nil, &order, nil, nil)

			// Fetch the last assistant message only, as this is the most relevant
			for _, message := range messages.Messages {
				if message.RunID != nil && *message.RunID == createRun.ID && message.Role == "assistant" {
					respContent += message.Content[0].Text.Value
					break
				}
			}

			// Fetch all files generated by the Agent, as this will be probably important. To-Do save in bucket
			for _, message := range messages.Messages {
				for _, rawAnnotation := range message.Content[0].Text.Annotations {
					fileID, _ := ParseAnnotation(rawAnnotation)
					annotatedFile, _ := detector.client.GetFile(ctx, fileID)
					fileContent, _ := detector.client.GetFileContent(ctx, annotatedFile.ID)
					var openaiFileName = analysisID + "_" + annotatedFile.ID // can identify the file by analysisID
					_ = saveFileLocally(openaiFileName, fileContent)
				}
			}

			break // Polling complete, exit
		} else {
			fmt.Println("Waiting for the Assistant to process...")
		}
	}

	//fmt.Println(respContent)
	return respContent
}

func (detector *GPTFraudDetector) preparePrompt() string {
	prompt := "Analyze the given CSV File dataset of financial transactions. Apply the following rules to detect potential fraud:\n\n. Save the results in a JSON file and return it to me. If no results, return an empty json file."

	for _, rule := range detector.rules {
		prompt += "- " + rule + "\n"
	}

	return prompt
}
