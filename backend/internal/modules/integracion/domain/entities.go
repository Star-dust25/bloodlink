package domain

import "errors"

var (
	ErrDNINoEncontrado     = errors.New("ciudadano no encontrado en el padrón nacional RENIEC")
	ErrDNIMenorDeEdad      = errors.New("el ciudadano es menor de 18 años y no está habilitado para donar")
	ErrDNIFormatoInvalido  = errors.New("el DNI debe contener exactamente 8 dígitos numéricos")
	ErrMapasNoDisponible   = errors.New("servicio externo de mapas no disponible (Contingencia E1)")
	ErrLegacyStockFaltante = errors.New("stock insuficiente en banco legado SOAP")
)

type CiudadanoReniec struct {
	DNI             string `json:"dni"`
	DNIValido       bool   `json:"dniValido"`
	NombreCompleto  string `json:"nombreCompleto"`
	FechaNacimiento string `json:"fechaNacimiento"`
	MayorEdad       bool   `json:"mayorEdad"`
}

type CalculoRuta struct {
	BancoDestinoID    string  `json:"bancoDestinoId"`
	DistanciaKm       float64 `json:"distanciaKm"`
	TiempoEstimadoMin int     `json:"tiempoEstimadoMin"`
}

type ReservaLegacyResultado struct {
	Exito             bool     `json:"exito"`
	BancoID           string   `json:"bancoId"`
	CodigoTransaccion string   `json:"codigoTransaccion"`
	UnidadesIDs       []string `json:"unidadesIds"`
	CodigosBarras     []string `json:"codigosBarras"`
	Mensaje           string   `json:"mensaje"`
}

type UDDIDirectorio struct {
	Servicio       string          `json:"servicio"`
	Version        string          `json:"version"`
	Namespace      string          `json:"namespace"`
	WSDLEndpoint   string          `json:"wsdlEndpoint"`
	PaginasBlancas UDDIPaginaBlanca `json:"paginasBlancas"`
	PaginasAmarillas []UDDIPaginaAmarilla `json:"paginasAmarillas"`
	PaginasVerdes  []UDDIPaginaVerde `json:"paginasVerdes"`
}

type UDDIPaginaBlanca struct {
	Organizacion string `json:"organizacion"`
	Entidad      string `json:"entidad"`
	Contacto     string `json:"contacto"`
	RUC          string `json:"ruc"`
	Jurisdiccion string `json:"jurisdiccion"`
}

type UDDIPaginaAmarilla struct {
	Categoria  string `json:"categoria"`
	CodigoUNSPSC string `json:"codigoUnspsc"`
	Estándar   string `json:"estandar"`
}

type UDDIPaginaVerde struct {
	Operacion   string `json:"operacion"`
	Protocolo   string `json:"protocolo"`
	BindingURL  string `json:"bindingUrl"`
	SOAPAction  string `json:"soapAction"`
	Descripcion string `json:"descripcion"`
}
