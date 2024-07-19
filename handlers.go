package main

import (
	"elarcacafe/models"
	"elarcacafe/services"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/tealeg/xlsx"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(handler.Filename)
	if ext != ".csv" && ext != ".xlsx" && ext != ".xls" {
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}

	// Save the uploaded file to a local directory
	tempFile, err := os.Create("uploaded" + ext)
	if err != nil {
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	_, err = tempFile.ReadFrom(file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	var listaDeGrupos models.ListaDeGrupos
	filePathXLSX := "uploaded.xlsx"
	if _, err := os.Stat(filePathXLSX); err == nil {
		file, err := xlsx.OpenFile(filePathXLSX)
		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}

		listaDeGrupos = services.ProcesarArchivoXLSX(file)
	} else {
		http.Error(w, "No uploaded file found", http.StatusBadRequest)
		return
	}
	jsonData, err := json.MarshalIndent(listaDeGrupos, "", "  ")
	if err != nil {
		http.Error(w, "Error converting to JSON", http.StatusInternalServerError)
		return
	}

	// Mostrar resultados en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
