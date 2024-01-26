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
	"strconv"
)

// PrettyPrint takes any Go data structure and returns a pretty-printed JSON string.
func PrettyPrint(v interface{}) (string, error) {
	prettyJSON, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return "", err // return an empty string and the error
	}
	return string(prettyJSON), nil
}

// Member represents the structure of a member in your JSON.
type Member struct {
	DOB        string                   `json:"DOB"`
	Dependents []map[string]interface{} `json:"dependents"`
	Email      string                   `json:"email"`
	Fips       string                   `json:"fips"`
	FirstName  string                   `json:"firstName"`
	Gender     string                   `json:"gender"`
	LastName   string                   `json:"lastName"`
	Tobacco    bool                     `json:"tobacco"`
	Zip        string                   `json:"zip"`
	Row        int                      `json:"row"`
	Column     int                      `json:"column"`
}

// MembersData represents the structure of your JSON object.
type MembersData struct {
	Members []Member `json:"members"`
}

func analyzeCensus(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong, retry"})
		return
	}
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

		_ = prependRowColumnNumbers(fileId)
		result, err := processCensus(fileId)
		c.JSON(200, result)
	case "application/vnd.ms-excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		// Handle XLS and XLSX files (convert and process)
		if err := convertXLSXToCSV(file, fileId); err != nil {
			fmt.Print(err)
			c.JSON(500, gin.H{"error": "Error in converting XLS/XLSX to CSV"})
			return
		}
		_ = prependRowColumnNumbers(fileId)
		result, err := processCensus(fileId)
		if err != nil {
			return
		}
		c.JSON(200, result)
	default:
		c.JSON(400, gin.H{"error": err})
	}
}

func prependRowColumnNumbers(fileName string) error {
	// Open the CSV file
	file, err := os.Open("./TransactionFiles/" + fileName + ".csv")
	if err != nil {
		return err
	}

	// Read the file into a 2D slice
	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		file.Close() // Close the file on read error
		return err
	}
	file.Close() // Close the file after reading

	// Add row numbers and prepare the header for column numbers
	for i, row := range records {
		records[i] = append([]string{strconv.Itoa(i)}, row...)
	}
	header := make([]string, len(records[0]))
	for j := range header {
		header[j] = strconv.Itoa(j)
	}

	// Prepend the header
	records = append([][]string{header}, records...)

	// Open the CSV file for writing (this will overwrite the existing file)
	file, err = os.Create("./TransactionFiles/" + fileName + ".csv")
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the modified data back to the file
	csvWriter := csv.NewWriter(file)
	err = csvWriter.WriteAll(records)
	if err != nil {
		return err
	}
	csvWriter.Flush()

	return nil
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

func processCensus(fileName string) (MembersData, error) {
	// Open the CSV file
	file, err := os.Open("./TransactionFiles/" + fileName + ".csv")
	if err != nil {
		return MembersData{}, err
	}
	defer file.Close()

	// Create a CSV reader
	reader := csv.NewReader(file)

	//Initialize the messages with a system message
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are Python REPL",
		},
	}

	var allResults MembersData

	// Process 10 rows at a time
	const batchSize = 15
	for {
		batch, err := readBatch(reader, batchSize)
		if err != nil {
			if err == io.EOF {
				if len(batch) > 0 {
					//fmt.Println("Calling openai,", batch)
					resultJSON, _ := callOpenai(batch, &messages)
					fmt.Println("Calling openai done ")
					appendMembers(resultJSON, &allResults)
				}
				break // End of file reached
			}
			return MembersData{}, err // Other error occurred
		}

		//fmt.Println("Calling openai,", batch)
		//Call OpenAI with the batch and messages
		resultJSON, err := callOpenai(batch, &messages)
		fmt.Println("Calling openai done ")
		if err != nil {
			return MembersData{}, err
		}
		appendMembers(resultJSON, &allResults)
	}

	return allResults, nil
}

// appendMembers takes a MembersData struct and appends its members to the Members slice of allResults.
func appendMembers(data MembersData, allResults *MembersData) {
	allResults.Members = append(allResults.Members, data.Members...)
}

func callOpenai(csvData [][]string, messages *[]openai.ChatCompletionMessage) (MembersData, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	csvString, err := convertCSVDataToString(csvData)
	fmt.Println("sendinf csv ", csvString)
	if err != nil {
		fmt.Println("Error converting CSV data to string:", err)
		return MembersData{}, err
	}

	userInstruction := fmt.Sprintf(`<python>
group_health_census_csv='''%s'''
json_schema = {
  "type": "object",
  "properties": {
    "members": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "firstName": { "type": "string" },
          "lastName": { "type": "string" },
          "email": { "type": ["string", "null"], "format": "email" },
          "DOB": { "type": "string", "pattern": "^(0[1-9]|1[0-2])/(0[1-9]|[12][0-9]|3[01])/(19|20)\\d{2}$" },
          "zip": { "type": "string", "pattern": "^[0-9]{5}(?:-[0-9]{4})?$" },
          "fips": { "type": "string" },
          "gender": { "type": "string", "enum": ["M", "F", "male", "female", "MALE", "FEMALE"] },
          "tobacco": { "type": "boolean", "default": false },
          "tobaccoUseDate": { "type": ["string", "null"], "format": "date" },
          "row": { "type": "integer" },
          "column": { "type": "integer" },
          "dependents": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "firstName": { "type": "string" },
                "lastName": { "type": "string" },
                "DOB": { "type": "string", "pattern": "^(0[1-9]|1[0-2])/(0[1-9]|[12][0-9]|3[01])/(19|20)\\d{2}$" },
                "gender": { "type": "string", "enum": ["M", "F", "male", "female", "MALE", "FEMALE"] },
                "tobacco": { "type": "boolean", "default": false },
                "tobaccoUseDate": { "type": ["string", "null"], "format": "date" },
                "relationship": { "type": "string", "enum": ["child", "spouse"] },
                "row": { "type": "integer" },
                "column": { "type": "integer" }
              }
            }
          }
        }
      }
    }
  }
}
# convert is a ML algorithm that takes in a JSON schema and CSV File and interprets unstructured CSV data into structured JSON schema.
# JSON Schema has an array of Employees and Dependents.
# CSV may have informtaion about company (should be ignored), miscellaneous(should be ignored) and a list of employees with their dependents (important).
# convert first strips out all information that is outside the list of employees and returns interpreted employee data as defined in JSON Schema provided.
# NOTE: the other sections may have some employee info that should be ignored, if no employee data is found it returns { "members": [] }
# member row, column and dependent row, column corresponds to row and column in group_health_census_csv
employee_plus_dependents_data = convert(json_schema = json_schema, group_health_census_csv = group_health_census_csv)
pprint.pprint(employee_plus_dependents_data)`, csvString)
	// Add a user message to the messages
	*messages = append(*messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userInstruction,
	})

	fmt.Println("Calling OpenAI")
	// Call OpenAI's Chat Completion API
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			//Model: openai.GPT4TurboPreview,
			Model:          openai.GPT3Dot5Turbo1106,
			ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
			Messages:       *messages,
		},
	)
	if err != nil {
		fmt.Println("Openai Err", err)
		return MembersData{}, err
	}

	fmt.Println("Calling OpenAI Done", resp.Choices[0].Message.Content)

	// Parse the response
	var resultJSON MembersData
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &resultJSON)
	if err != nil {
		fmt.Println(err)
		return MembersData{}, err
	}

	// Append the result to messages as a system message
	*messages = append(*messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: resp.Choices[0].Message.Content,
	})

	pp, _ := PrettyPrint(resultJSON)
	fmt.Println("openai result ", pp)

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
