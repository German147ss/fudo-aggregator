package services

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/extrame/xls"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/tealeg/xlsx"
)

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

	// Intentar abrir como .xls
	fileXLS, errXLS := xls.Open("gastos/gastos.xls", "utf-8")
	if errXLS == nil {
		log.Println("Leyendo archivo .xls")
		registros, err = leerArchivoXLS(fileXLS)
		if err != nil {
			log.Fatalf("Error al leer el archivo .xls: %v", err)
		}
	} else {
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
	fmt.Println("---|---|---|")
	for _, grupo := range listaGrupos {
		fmt.Printf("%s | %d | %.2f\n", grupo.Grupo, grupo.Cantidad, grupo.Gasto)
	}
}

func leerArchivoXLS(file *xls.WorkBook) ([]Registro, error) {
	var registros []Registro
	for i := 0; i < file.NumSheets(); i++ {
		sheet := file.GetSheet(i)
		if sheet == nil {
			continue // Saltar si la hoja es nil
		}
		log.Printf("Leyendo hoja %d", i)
		log.Printf("Hoja tiene %d filas", sheet.MaxRow)
		for rowIndex := STARTS_IN_ROW; rowIndex <= int(sheet.MaxRow); rowIndex++ {
			log.Printf("Leyendo fila %d", rowIndex)
			row := sheet.Row(rowIndex)
			if row == nil || row.LastCol() < 14 {
				continue // Saltar filas incompletas
			}

			importeStr := row.Col(PRICE_COLUMN)
			log.Printf("Importe: %s", importeStr)
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

func leerArchivoXLSX(file *xlsx.File) ([]Registro, error) {
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
