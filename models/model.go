package models

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
