## LLM Powered API for Fraud Detection.

## Project Workflow Overview

The Fraud Detection API is designed to process financial transaction data for fraud analysis, specifically handling CSV files. The workflow comprises a series of steps that facilitate asynchronous processing and analysis.

### Workflow Steps:

1. **Upload Transactions**:
   - Users can upload a CSV file containing transaction data via the `/upload` endpoint. Uploaded files are temporarily stored in the `TransactionFiles` directory. (To-Do: Will be migrated to an AWS storage bucket).

2. **Analyze Transactions**:
    - The user then initiates the analysis by calling the `/analyze/:fileID` endpoint, passing the `file_id` obtained from the upload step. The analysis process is started asynchronously. This endpoint accepts a set of fraud detection rules in the request body.
   ```
      {
         "rules": [
            "Rule 1 description",
            "Rule 2 description",
            // additional rules...
         ]
      }
      ```

3. **Poll Analysis Status**:
    - While the analysis is ongoing, the user can check its status through the `/analyze/:analysisId/status` endpoint. The status provides insights into whether the analysis is `Processing`, `Complete`, or has `Failed`.

4. **Retrieve Analysis Results**:
   - Once the analysis is complete, the results can be fetched from the `/results/:analysisId` endpoint using the provided `analysisId`. Additional Results may be returned as JSON files, which are stored in the `ResultFiles` directory. (To-Do: Will be migrated to an AWS storage bucket) If present, the files will be referenced in the API Response.

5. **Health Check**:
    - The API also includes a `/health` endpoint, allowing users to verify if the API service is operational.

### How it works

![Flowchart Diagram](./assets/FraudDetectFlow.svg)

## Onboarding

Build the Image:
```
docker build -t frauddetection .
```
Run the Container
```
docker run -p 8080:8080 frauddetection
```
Access the application at ```http://localhost:8080```

## Swagger API Documentation

The Fraud Detection API comes with an integrated Swagger UI.

### Accessing Swagger UI:

- To access the Swagger UI, first ensure the API service is running.
- Navigate to `http://localhost:8080/swagger/index.html` in your web browser.
- The Swagger UI page will display a list of all available API endpoints with their expected parameters and responses.
- You can interact with the API directly from this page by expanding individual endpoint details, entering required parameters, and executing requests to see the responses in real-time.

## Tech Stack

### Golang

- The application is built in [Golang](https://golang.org/), selected for its robust support for concurrent and asynchronous processing using goroutines. This feature is crucial for efficiently handling the asynchronous tasks and large-scale data processing inherent in fraud detection.

### OpenAI Assistants API

- For fraud analysis, the project utilizes the [OpenAI Assistants API](https://platform.openai.com/docs/assistants/how-it-works). This Assistants API aids in analyzing transaction data to detect potential fraud patterns.
