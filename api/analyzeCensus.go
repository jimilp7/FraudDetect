package api

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/sashabaranov/go-openai"
	"github.com/xuri/excelize/v2"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

func analyzeCensus(c *gin.Context) {
	file, err := c.FormFile("file")
	// Generate a NanoID
	fileId, err := gonanoid.New()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate an ID"})
		return
	}

	if err != nil {
		c.JSON(400, gin.H{"error": "No file is received"})
		return
	}

	// MIME type validation
	// MIME type validation
	mimeType := file.Header.Get("Content-Type")
	switch mimeType {
	case "text/csv":
		// Save the CSV file to the TransactionFiles directory
		dst, err := os.Create("./TransactionFiles/" + fileId + ".csv")
		if err != nil {
			c.JSON(500, gin.H{"error": "Error in saving CSV file"})
			return
		}
		defer dst.Close()

		// Copy the file content to the new file
		src, err := file.Open()
		if err != nil {
			c.JSON(500, gin.H{"error": "Error in opening uploaded CSV file"})
			return
		}
		defer src.Close()

		if _, err := io.Copy(dst, src); err != nil {
			c.JSON(500, gin.H{"error": "Error in copying CSV file"})
			return
		}

		result, err := processCensus(fileId)

		c.JSON(200, gin.H{"message": result})
	case "application/vnd.ms-excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		// Handle XLS and XLSX files (convert and process)
		if err := convertXLSXToCSV(file, fileId); err != nil {
			fmt.Print(err)
			c.JSON(500, gin.H{"error": "Error in converting XLS/XLSX to CSV"})
			return
		}
		result, err := processCensus(fileId)
		if err != nil {
			return
		}

		c.JSON(200, gin.H{"message": result})
	default:
		c.JSON(400, gin.H{"error": err})
	}
}

func convertXLSXToCSV(file *multipart.FileHeader, fileId string) error {
	// Open the XLSX file
	srcFile, err := file.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()

	xlsx, err := excelize.OpenReader(srcFile)
	if err != nil {
		return err
	}

	// Get all sheet names
	sheets := xlsx.GetSheetList()

	// Check if there's at least one sheet
	if len(sheets) == 0 {
		return errors.New("no sheets found in XLSX file")
	}

	// Use the first sheet
	firstSheetName := sheets[0]

	// Create a new CSV file
	csvFile, err := os.Create("./TransactionFiles/" + fileId + ".csv")
	if err != nil {
		return err
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	// Determine the max number of columns
	maxCols := 0
	rows, err := xlsx.GetRows(firstSheetName)
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	// Process each row, ensuring each has maxCols fields
	for _, row := range rows {
		if len(row) < maxCols {
			row = append(row, make([]string, maxCols-len(row))...)
		}
		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func processCensus(fileName string) (map[string]interface{}, error) {
	// Open the CSV file
	file, err := os.Open("./TransactionFiles/" + fileName + ".csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a CSV reader
	reader := csv.NewReader(file)

	//Initialize the messages with a system message
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are an Data Engineer working at Google with over 20 years experience who is a perfectionist at their job.",
		},
	}

	var allResults []map[string]interface{}

	// Process 10 rows at a time
	const batchSize = 10
	for {
		batch, err := readBatch(reader, batchSize)
		if err != nil {
			if err == io.EOF {
				if len(batch) > 0 {
					fmt.Println("Calling openai")
					resultJSON, _ := callOpenai(batch, &messages)
					fmt.Println("Calling openai done ")
					allResults = append(allResults, resultJSON)
				}
				break // End of file reached
			}
			return nil, err // Other error occurred
		}

		//Call OpenAI with the batch and messages
		resultJSON, err := callOpenai(batch, &messages)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, resultJSON)
	}

	//Flatten
	flatResult := flattenResults(allResults)
	return flatResult, nil
}

func flattenResults(allResults []map[string]interface{}) map[string]interface{} {
	flatResult := make(map[string]interface{})
	for _, resultMap := range allResults {
		for key, value := range resultMap {
			flatResult[key] = value
		}
	}
	return flatResult
}

func callOpenai(csvData [][]string, messages *[]openai.ChatCompletionMessage) (map[string]interface{}, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	csvString, err := convertCSVDataToString(csvData)
	if err != nil {
		fmt.Println("Error converting CSV data to string:", err)
		return nil, err
	}

	userInstruction := fmt.Sprintf("Use this JSON Schema as a reference to parse the following csv by hand and return to me the parsed JSON object.\n<JSON Schema>\n{ \"type\": \"array\", \"items\": { \"type\": \"object\", \"properties\": { \"email\": { \"type\": \"string\" }, \"gender\": { \"type\": \"string\", \"enum\": [\"0\", \"1\", \"m\", \"f\", \"male\", \"female\"] }, \"name\": { \"type\": \"string\" }, \"zip_code\": { \"type\": \"string\", \"pattern\": \"^[0-9]{5}(?:-[0-9]{4})?$\" }, \"last_tobacco_use_date\": { \"type\": [\"string\", \"null\"], \"format\": \"date\" }, \"salary\": { \"type\": [\"integer\", \"null\"] }, \"date_of_birth\": { \"type\": \"string\", \"format\": \"date\" }, \"address1\": { \"type\": [\"string\", \"null\"] }, \"address2\": { \"type\": [\"string\", \"null\"] }, \"city\": { \"type\": [\"string\", \"null\"] }, \"phone_number\": { \"type\": [\"string\", \"null\"] }, \"hireDate\": { \"type\": [\"string\", \"null\"], \"format\": \"date\" }, \"jobTitle\": { \"type\": [\"string\", \"null\"] }, \"ssn\": { \"type\": [\"string\", \"null\"] }, \"employment_type\": { \"type\": [\"string\", \"null\"], \"enum\": [\"fullTime\", \"partTime\"] }, \"dependents\": { \"type\": [\"array\", \"null\"], \"items\": { \"$ref\": \"#/definitions/dependent\" } } }, \"required\": [\"email\", \"gender\", \"name\", \"zip_code\", \"date_of_birth\"] }, \"definitions\": { \"dependent\": { \"type\": \"object\", \"properties\": { \"firstName\": { \"type\": \"string\" }, \"lastName\": { \"type\": \"string\" }, \"zipCode\": { \"type\": [\"string\", \"null\"], \"pattern\": \"^[0-9]{5}(?:-[0-9]{4})?$\" }, \"countyID\": { \"type\": [\"string\", \"null\"] }, \"dateOfBirth\": { \"type\": \"string\", \"format\": \"date\" }, \"gender\": { \"type\": \"string\", \"enum\": [\"0\", \"1\", \"m\", \"f\", \"male\", \"female\"] }, \"lastUsedTobacco\": { \"type\": [\"string\", \"null\"], \"format\": \"date\" }, \"relationship\": { \"type\": \"string\", \"enum\": [\"spouse\", \"child\"] }, \"ssn\": { \"type\": [\"string\", \"null\"] } }, \"required\": [\"firstName\", \"lastName\", \"dateOfBirth\", \"gender\", \"relationship\"] } } }\n</JSON Schema>\n<csv>%s</csv>\n\nFor example:\n<JSON Schema>\n{ \"type\": \"array\", \"items\": { \"type\": \"object\", \"properties\": { \"email\": { \"type\": \"string\" }, \"gender\": { \"type\": \"string\", \"enum\": [\"0\", \"1\", \"m\", \"f\", \"male\", \"female\"] }, \"name\": { \"type\": \"string\" }, \"zip_code\": { \"type\": \"string\", \"pattern\": \"^[0-9]{5}(?:-[0-9]{4})?$\" }, \"last_tobacco_use_date\": { \"type\": [\"string\", \"null\"], \"format\": \"date\" }, \"salary\": { \"type\": [\"integer\", \"null\"] }, \"date_of_birth\": { \"type\": \"string\", \"format\": \"date\" }, \"address1\": { \"type\": [\"string\", \"null\"] }, \"address2\": { \"type\": [\"string\", \"null\"] }, \"city\": { \"type\": [\"string\", \"null\"] }, \"phone_number\": { \"type\": [\"string\", \"null\"] }, \"hireDate\": { \"type\": [\"string\", \"null\"], \"format\": \"date\" }, \"jobTitle\": { \"type\": [\"string\", \"null\"] }, \"ssn\": { \"type\": [\"string\", \"null\"] }, \"employment_type\": { \"type\": [\"string\", \"null\"], \"enum\": [\"fullTime\", \"partTime\"] }, \"dependents\": { \"type\": [\"array\", \"null\"], \"items\": { \"$ref\": \"#/definitions/dependent\" } } }, \"required\": [\"email\", \"gender\", \"name\", \"zip_code\", \"date_of_birth\"] }, \"definitions\": { \"dependent\": { \"type\": \"object\", \"properties\": { \"firstName\": { \"type\": \"string\" }, \"lastName\": { \"type\": \"string\" }, \"zipCode\": { \"type\": [\"string\", \"null\"], \"pattern\": \"^[0-9]{5}(?:-[0-9]{4})?$\" }, \"countyID\": { \"type\": [\"string\", \"null\"] }, \"dateOfBirth\": { \"type\": \"string\", \"format\": \"date\" }, \"gender\": { \"type\": \"string\", \"enum\": [\"0\", \"1\", \"m\", \"f\", \"male\", \"female\"] }, \"lastUsedTobacco\": { \"type\": [\"string\", \"null\"], \"format\": \"date\" }, \"relationship\": { \"type\": \"string\", \"enum\": [\"spouse\", \"child\"] }, \"ssn\": { \"type\": [\"string\", \"null\"] } }, \"required\": [\"firstName\", \"lastName\", \"dateOfBirth\", \"gender\", \"relationship\"] } } }\n</JSON Schema>\n<csv>EID,First Name,Middle Name,Last Name,Location,Relationship,SSN,Birth Date,Sex,Race,Citizenship,Language,Address 1,Address 2,City,State,Zip Code,County,Personal Phone,Work Phone,Mobile Phone,Email,Personal Email,Marital Status,Hire Date,Pay Cycle,Tobacco User,Disabled,Manager,HR Manager,Department,Division,Job Title,Job Class,Employment Type,Status,Remaining Deduction Periods,Benefit Cost Factor,Compensation Amount,Compensation Type,Compensation Start Date,Compensation Reason,Base + Commission Compensation Amount,Base + Commission Compensation Start Date,Base + Commission Compensation Reason,W2 Wages,Scheduled Hours,Sick Hours,Personal Hours,Termination Date,COBRA Date,Rehire Date,Benefit Eligible Date,Dependent Employer,Unlock Enrollment Date,GI Plan Types,GI Amount,GI Date,Member Id Unum,Member Id Humana,Notes\n123456,Alice,Jane,Williams,Branch 2,Employee,987-65-4321,05-15-88,F,Asian,USA,Spanish,456 Oak Street,Unit 5,Chicago,IL,60601,Cook,555-987-6543,555-123-4567,555-789-1234,alice.williams@example.com,alice@gmail.com,Single,03-08-20,Monthly,No,Yes,Emily Smith,HR Director,Sales,Division B,Sales Manager,Entry,Part-Time,Active,40,9.8,22,10-15-20,,,12-08-19,,08-13-18,Life,\"$60,000.00\",07-01-19,DEF-789,XYZ-456,These are some random notes for the employee.</csv>\n<Parsed Result>[\n    {\n        \"email\": \"alice.williams@example.com\",\n        \"gender\": \"f\",\n        \"name\": \"Alice Jane Williams\",\n        \"zip_code\": \"60601\",\n        \"date_of_birth\": \"1988-05-15\",\n        \"last_tobacco_use_date\": null,\n        \"salary\": null,\n        \"address1\": \"456 Oak Street\",\n        \"address2\": \"Unit 5\",\n        \"city\": \"Chicago\",\n        \"phone_number\": \"555-987-6543\",\n        \"hireDate\": \"2020-03-08\",\n        \"jobTitle\": \"Sales Manager\",\n        \"ssn\": \"987-65-4321\",\n        \"employment_type\": \"partTime\",\n        \"dependents\": []\n    }\n]</Parsed Result>", csvString)

	// Add a user message to the messages
	*messages = append(*messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userInstruction,
	})

	// Call OpenAI's Chat Completion API
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:          openai.GPT4TurboPreview,
			ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
			Messages:       *messages,
		},
	)
	if err != nil {
		fmt.Println("Openai Err", err)
		return nil, err
	}

	// Parse the response
	var resultJSON map[string]interface{}
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &resultJSON)
	if err != nil {
		return nil, err
	}

	// Append the result to messages as a system message
	*messages = append(*messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: resp.Choices[0].Message.Content,
	})

	return resultJSON, nil
}

func readBatch(reader *csv.Reader, batchSize int) ([][]string, error) {
	var batch [][]string
	for i := 0; i < batchSize; i++ {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				fmt.Println("End of file reached, batch size:", len(batch))
				return batch, err
			}
			fmt.Println("Error reading file:", err)
			return nil, err
		}
		batch = append(batch, record)
	}
	return batch, nil
}

func convertCSVDataToString(csvData [][]string) (string, error) {
	buffer := bytes.NewBufferString("")
	csvWriter := csv.NewWriter(buffer)

	for _, row := range csvData {
		if err := csvWriter.Write(row); err != nil {
			return "", err
		}
	}
	csvWriter.Flush()

	return buffer.String(), nil
}
