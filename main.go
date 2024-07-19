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
		if datos, exists := productos[proveedor]; exists && registro.Cancelado == "No" {

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
