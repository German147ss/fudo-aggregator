[![Go Dev](https://pkg.go.dev/badge/www.linkedin.com/in/mendietagerman/.svg)](https://pkg.go.dev/www.linkedin.com/in/mendietagerman/)

# FUDO Costs File Upload Service

Fudo Costs File Upload Service is a Go-based API that allows users to upload CSV and XLSX files, process them in memory, and return the results in JSON format. This service is designed to be lightweight, efficient, and easy to integrate into existing applications.

## Features

- **File Upload**: Allows uploading of XLSX files via an HTTP POST request.
- **In-Memory Processing**: Files are processed directly in memory without the need to save them to the file system.
- **JSON Response**: The processing results are returned in JSON format, making it easy to integrate with front-end applications.

## Requirements

- Go 1.22 or higher

## Installation

1. Clone this repository:
   ```sh
   git clone https://github.com/German147ss/fudo-aggregator.git
   cd fudo-aggregator
   ```
2. Install the dependencies:
   ```sh
   go mod download
   ```

## API Endpoints

```POST /upload```

**Uploads and processes a CSV or XLSX file**

- **Request**

### Headers:

Content-Type: multipart/form-data

### Body:

file: The CSV or XLSX file to upload.

### Response

#### Success (200 OK):

Returns the processed data in JSON format.

#### Error (4xx/5xx):

Returns an error message.