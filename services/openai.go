package services

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
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

func (detector *GPTFraudDetector) RunAnalysis() string {
	prompt := detector.preparePrompt()

	ctx := context.Background()

	// File upload
	engineFileId := "iQ5UKtTa5FJQ2zMaZLLbB"
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
		Instructions: StringPointer("You are a Professional Fraud Detection Specialist named Onyx working at JP Morgan Chase and have over 20 years experience in flagging fraudulent transactions. You will be provided with a CSV containing credit card transaction data."),
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
		run, _ := detector.client.RetrieveRun(ctx, thread.ID, createRun.ID)

		if run.Status == "completed" {
			order := "desc"
			messages, _ := detector.client.ListMessage(ctx, thread.ID, nil, &order, nil, nil)

			for _, message := range messages.Messages {
				if message.RunID != nil && *message.RunID == createRun.ID && message.Role == "assistant" {
					respContent += message.Content[0].Text.Value
					break
				}
			}
			break

		} else {
			fmt.Println("Waiting for the Assistant to process...")
		}
	}

	//fmt.Println(respContent)
	return respContent
}

func (detector *GPTFraudDetector) preparePrompt() string {
	prompt := "Analyze the given dataset of financial transactions. Apply the following rules to detect potential fraud:\n\n. Return the list of fraudulent transactions in a detailed JSON Format, if none, return an empty JSON. Do not store in file, return directly."

	for _, rule := range detector.rules {
		prompt += "- " + rule + "\n"
	}

	return prompt
}
