package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/tealeg/xlsx"
)

func main() {
	LeerArchivo()
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

func LeerArchivo() {
	var registros []Registro
	var err error

	log.Println("Intentando abrir archivo .xlsx")
	// Intentar abrir como .xlsx
	fileXLSX, errXLSX := xlsx.OpenFile("gastos/gastos.xlsx")
	if errXLSX != nil {
		log.Fatalf("Error al abrir el archivo: %v", errXLSX)
	}
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
	var listaGrupos []struct {
		Grupo    string
		Cantidad int
		Gasto    float64
	}
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

	// Mostrar resultados
	fmt.Println("Grupo | Cantidad | Gasto")
	fmt.Println("--------|--------|--------|")
	for _, grupo := range listaGrupos {
		fmt.Printf("%s | %d | %.2f\n", grupo.Grupo, grupo.Cantidad, grupo.Gasto)
	}
}

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

/*
import (
	"elarcacafe/services"
	"encoding/csv"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/extrame/xls"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/tealeg/xlsx"
)

type Registro struct {
	Id           string  `json:"id"`
	Caja         string  `json:"caja"`
	Categoria    string  `json:"categoria"`
	Fecha        string  `json:"fecha"`
	Proveedor    string  `json:"proveedor"`
	NumeroFiscal string  `json:"numero_fiscal"`
	TipoComp     string  `json:"tipo_comp"`
	NroComp      string  `json:"nro_comp"`
	Descripcion  string  `json:"descripcion"`
	Importe      float64 `json:"importe"`
	MedioPago    string  `json:"medio_pago"`
	CreadoPor    string  `json:"creado_por"`
	DeCaja       string  `json:"de_caja"`
	Cancelado    string  `json:"cancelado"`
}

type Producto struct {
	Grupo    string  `json:"grupo"`
	Cantidad int     `json:"cantidad"`
	Gasto    float64 `json:"gasto"`
}

func main() {
	services.LeerArchivo()
	http.HandleFunc("/upload", cors(uploadHandler))
	http.HandleFunc("/process", cors(processHandler))
	log.Printf("Servidor corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

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
	var registros []Registro
	var err error

	filePathCSV := "uploaded.csv"
	filePathXLSX := "uploaded.xlsx"
	filePathXLS := "uploaded.xls"

	if _, err := os.Stat(filePathCSV); err == nil {
		file, err := os.Open(filePathCSV)
		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		registros, err = leerCSV(file)
		if err != nil {
			http.Error(w, "Error processing file", http.StatusInternalServerError)
			return
		}
	} else if _, err := os.Stat(filePathXLSX); err == nil {
		file, err := xlsx.OpenFile(filePathXLSX)
		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}

		registros, err = leerXLSX(file)
		if err != nil {
			http.Error(w, "Error processing file", http.StatusInternalServerError)
			return
		}
	} else if _, err := os.Stat(filePathXLS); err == nil {
		file, err := xls.Open(filePathXLS, "utf-8")
		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}

		registros, err = leerXLS(file)
		if err != nil {
			http.Error(w, "Error processing file", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "No uploaded file found", http.StatusBadRequest)
		return
	}

	// Procesar y agrupar los productos
	productos := map[string]struct {
		Cantidad int
		Gasto    float64
	}{}
	for _, registro := range registros {
		proveedor := registro.Proveedor
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

	grupos := agruparProductos(productos)

	// Convertir el mapa a una lista para ordenar
	var listaGrupos []Producto
	for grupo, datos := range grupos {
		listaGrupos = append(listaGrupos, Producto{
			Grupo:    grupo,
			Cantidad: datos.Cantidad,
			Gasto:    datos.Gasto,
		})
	}

	// Ordenar la lista de grupos por gasto en orden descendente
	sort.Slice(listaGrupos, func(i, j int) bool {
		return listaGrupos[i].Gasto > listaGrupos[j].Gasto
	})

	// Convertir a JSON
	jsonData, err := json.MarshalIndent(listaGrupos, "", "  ")
	if err != nil {
		http.Error(w, "Error converting to JSON", http.StatusInternalServerError)
		return
	}

	// Mostrar resultados en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func leerCSV(file *os.File) ([]Registro, error) {
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var registros []Registro
	for i, record := range records {
		if i == 0 {
			continue // Skip header
		}

		importe, err := strconv.ParseFloat(record[9], 64)
		if err != nil {
			log.Printf("Error al convertir el importe: %v", err)
			continue
		}

		registro := Registro{
			Id:           record[0],
			Caja:         record[1],
			Categoria:    record[2],
			Fecha:        record[3],
			Proveedor:    record[4],
			NumeroFiscal: record[5],
			TipoComp:     record[6],
			NroComp:      record[7],
			Descripcion:  record[8],
			Importe:      importe,
			MedioPago:    record[10],
			CreadoPor:    record[11],
			DeCaja:       record[12],
			Cancelado:    record[13],
		}

		registros = append(registros, registro)
	}
	return registros, nil
}

func leerXLSX(file *xlsx.File) ([]Registro, error) {
	var registros []Registro
	for _, sheet := range file.Sheets {
		for index, row := range sheet.Rows {
			if index == 0 {
				continue // Saltar la primera fila
			}
			if len(row.Cells) < 14 {
				continue // Saltar filas incompletas
			}

			importe, err := row.Cells[9].Float()
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

func leerXLS(file *xls.WorkBook) ([]Registro, error) {
	var registros []Registro
	for i := 0; i < file.NumSheets(); i++ {
		sheet := file.GetSheet(i)
		if sheet == nil {
			continue // Saltar si la hoja es nil
		}
		for rowIndex := 1; rowIndex <= int(sheet.MaxRow); rowIndex++ {
			row := sheet.Row(rowIndex)
			if row == nil || row.LastCol() < 14 {
				continue // Saltar filas incompletas
			}

			importeStr := row.Col(9)
			importe, err := strconv.ParseFloat(importeStr, 64)
			if err != nil {
				log.Printf("Error al convertir el importe: %v", err)
				continue
			}

			registro := Registro{
				Id:           row.Col(0),
				Caja:         row.Col(1),
				Categoria:    row.Col(2),
				Fecha:        row.Col(3),
				Proveedor:    row.Col(4),
				NumeroFiscal: row.Col(5),
				TipoComp:     row.Col(6),
				NroComp:      row.Col(7),
				Descripcion:  row.Col(8),
				Importe:      importe,
				MedioPago:    row.Col(10),
				CreadoPor:    row.Col(11),
				DeCaja:       row.Col(12),
				Cancelado:    row.Col(13),
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
*/
