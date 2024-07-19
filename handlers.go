package main

import (
	"elarcacafe/services"
	"encoding/json"
	"net/http"

	"github.com/tealeg/xlsx"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Leer el archivo cargado en memoria
	xlFile, err := xlsx.OpenReaderAt(file, r.ContentLength)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusBadRequest)
		return
	}

	// Procesar el archivo directamente en memoria
	listaDeGrupos := services.ProcesarArchivoXLSX(xlFile)

	// Convertir los datos a JSON
	jsonData, err := json.MarshalIndent(listaDeGrupos, "", "  ")
	if err != nil {
		http.Error(w, "Error converting to JSON", http.StatusInternalServerError)
		return
	}

	// Mostrar resultados en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
