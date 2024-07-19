package services

import (
	"elarcacafe/models"
	"log"
	"sort"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/tealeg/xlsx"
)

func ProcesarArchivoXLSX(fileXLSX *xlsx.File) models.ListaDeGrupos {
	var registros []models.Registro
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
	var listaGrupos models.ListaDeGrupos
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

func leerArchivoXLSX(file *xlsx.File) ([]models.Registro, error) {
	var registros []models.Registro
	for _, sheet := range file.Sheets {
		for index, row := range sheet.Rows {
			if index < models.STARTS_IN_ROW {
				continue // Saltar la primera fila
			}
			if len(row.Cells) < 14 {
				continue // Saltar filas incompletas
			}

			importe, err := row.Cells[models.PRICE_COLUMN].Float()
			if err != nil {
				log.Printf("Error al convertir el importe: %v", err)
				continue
			}

			registro := models.Registro{
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
