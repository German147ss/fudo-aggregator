package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/tealeg/xlsx"
)

func main() {
	http.HandleFunc("/upload", cors(uploadHandler))
	http.HandleFunc("/process", cors(processHandler))
	log.Printf("Servidor corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

const STARTS_IN_ROW = 4
const PRICE_COLUMN = 9

type Registro struct {
	Id           string
	Caja         string
	Categoria    string
	Fecha        string
	Proveedor    string
	NumeroFiscal string
	TipoComp     string
	NroComp      string
	Descripcion  string
	Importe      float64
	MedioPago    string
	CreadoPor    string
	DeCaja       string
	Cancelado    string
}

type Grupo struct {
	Grupo    string
	Cantidad int
	Gasto    float64
}

type ListaDeGrupos []Grupo

func cors(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5500")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler(w, r)
	}
}

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
	var listaDeGrupos ListaDeGrupos
	filePathXLSX := "uploaded.xlsx"
	if _, err := os.Stat(filePathXLSX); err == nil {
		file, err := xlsx.OpenFile(filePathXLSX)
		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}

		listaDeGrupos = ProcesarArchivoXLSX(file)
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
func ProcesarArchivoXLSX(fileXLSX *xlsx.File) ListaDeGrupos {
	var registros []Registro
	var err error
	registros, err = leerArchivoXLSX(fileXLSX)
	if err != nil {
		log.Fatalf("Error al leer el archivo .xlsx: %v", err)
	}
	log.Println("Agrupando registros = ", len(registros))

	// Procesar y agrupar los productos
	productos := map[string]struct {
		Cantidad int
		Gasto    float64
	}{}
	for _, registro := range registros {
		proveedor := registro.Proveedor
		if registro.Cancelado == "No" {
			if datos, exists := productos[proveedor]; exists {
				datos.Cantidad++
				datos.Gasto += registro.Importe
				productos[proveedor] = datos
			} else {
				productos[proveedor] = struct {
					Cantidad int
					Gasto    float64
				}{
					Cantidad: 1,
					Gasto:    registro.Importe,
				}
			}
		}
	}

	grupos := agruparProductos(productos)

	// Convertir el mapa a una lista para ordenar
	var listaGrupos ListaDeGrupos
	for grupo, datos := range grupos {
		listaGrupos = append(listaGrupos, struct {
			Grupo    string
			Cantidad int
			Gasto    float64
		}{
			Grupo:    grupo,
			Cantidad: datos.Cantidad,
			Gasto:    datos.Gasto,
		})
	}

	// Ordenar la lista de grupos por gasto en orden descendente
	sort.Slice(listaGrupos, func(i, j int) bool {
		return listaGrupos[i].Gasto > listaGrupos[j].Gasto
	})
	return listaGrupos
}

/* func LeerArchivo() {
	log.Println("Intentando abrir archivo .xlsx")
	// Intentar abrir como .xlsx
	fileXLSX, errXLSX := xlsx.OpenFile("gastos/gastos.xlsx")
	if errXLSX != nil {
		log.Fatalf("Error al abrir el archivo: %v", errXLSX)
	}

	ProcesarArchivoXLSX(fileXLSX)
} */

func leerArchivoXLSX(file *xlsx.File) ([]Registro, error) {
	var registros []Registro
	for _, sheet := range file.Sheets {
		for index, row := range sheet.Rows {
			if index < STARTS_IN_ROW {
				continue // Saltar la primera fila
			}
			if len(row.Cells) < 14 {
				continue // Saltar filas incompletas
			}

			importe, err := row.Cells[PRICE_COLUMN].Float()
			if err != nil {
				log.Printf("Error al convertir el importe: %v", err)
				continue
			}

			registro := Registro{
				Id:           row.Cells[0].String(),
				Caja:         row.Cells[1].String(),
				Categoria:    row.Cells[2].String(),
				Fecha:        row.Cells[3].String(),
				Proveedor:    row.Cells[4].String(),
				NumeroFiscal: row.Cells[5].String(),
				TipoComp:     row.Cells[6].String(),
				NroComp:      row.Cells[7].String(),
				Descripcion:  row.Cells[8].String(),
				Importe:      importe,
				MedioPago:    row.Cells[10].String(),
				CreadoPor:    row.Cells[11].String(),
				DeCaja:       row.Cells[12].String(),
				Cancelado:    row.Cells[13].String(),
			}

			registros = append(registros, registro)
		}
	}
	return registros, nil
}

func agruparProductos(productos map[string]struct {
	Cantidad int
	Gasto    float64
}) map[string]struct {
	Cantidad  int
	Gasto     float64
	Productos []string
} {
	grupos := map[string]struct {
		Cantidad  int
		Gasto     float64
		Productos []string
	}{}
	for producto, datos := range productos {
		grupo := encontrarGrupoSimilar(producto, grupos)
		if grupo == "" {
			grupos[producto] = struct {
				Cantidad  int
				Gasto     float64
				Productos []string
			}{
				Cantidad:  datos.Cantidad,
				Gasto:     datos.Gasto,
				Productos: []string{producto},
			}
		} else {
			d := grupos[grupo]
			d.Cantidad += datos.Cantidad
			d.Gasto += datos.Gasto
			d.Productos = append(d.Productos, producto)
			grupos[grupo] = d
		}
	}
	return grupos
}

func encontrarGrupoSimilar(producto string, grupos map[string]struct {
	Cantidad  int
	Gasto     float64
	Productos []string
}) string {
	for key := range grupos {
		if fuzzy.Match(strings.ToLower(producto), strings.ToLower(key)) {
			return key
		}
	}
	return ""
}
