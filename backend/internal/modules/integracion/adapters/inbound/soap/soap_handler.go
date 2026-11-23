package soap

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/bloodlink/backend/internal/modules/integracion/domain"
	"github.com/bloodlink/backend/internal/modules/integracion/ports"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const BloodLinkWSDL = `<?xml version="1.0" encoding="UTF-8"?>
<definitions name="ServicioBloodLinkSOAP"
             targetNamespace="http://bloodlink.minsa.gob.pe/soap/interoperabilidad"
             xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="http://bloodlink.minsa.gob.pe/soap/interoperabilidad"
             xmlns:xsd="http://www.w3.org/2001/XMLSchema">

  <types>
    <xsd:schema targetNamespace="http://bloodlink.minsa.gob.pe/soap/interoperabilidad">
      <xsd:element name="ValidarIdentidadReniecRequest">
        <xsd:complexType>
          <xsd:sequence>
            <xsd:element name="dni" type="xsd:string"/>
          </xsd:sequence>
        </xsd:complexType>
      </xsd:element>
      <xsd:element name="ValidarIdentidadReniecResponse">
        <xsd:complexType>
          <xsd:sequence>
            <xsd:element name="dni" type="xsd:string"/>
            <xsd:element name="dniValido" type="xsd:boolean"/>
            <xsd:element name="nombreCompleto" type="xsd:string"/>
            <xsd:element name="fechaNacimiento" type="xsd:string"/>
            <xsd:element name="mayorEdad" type="xsd:boolean"/>
          </xsd:sequence>
        </xsd:complexType>
      </xsd:element>
      <xsd:element name="ConsultarYReservarStockLegacyRequest">
        <xsd:complexType>
          <xsd:sequence>
            <xsd:element name="bancoId" type="xsd:string"/>
            <xsd:element name="tipoSangre" type="xsd:string"/>
            <xsd:element name="factorRh" type="xsd:string"/>
            <xsd:element name="cantidad" type="xsd:int"/>
          </xsd:sequence>
        </xsd:complexType>
      </xsd:element>
      <xsd:element name="ConsultarYReservarStockLegacyResponse">
        <xsd:complexType>
          <xsd:sequence>
            <xsd:element name="exito" type="xsd:boolean"/>
            <xsd:element name="bancoId" type="xsd:string"/>
            <xsd:element name="codigoTransaccion" type="xsd:string"/>
            <xsd:element name="unidadesReservadas" type="xsd:string"/>
            <xsd:element name="codigosBarras" type="xsd:string"/>
            <xsd:element name="mensaje" type="xsd:string"/>
          </xsd:sequence>
        </xsd:complexType>
      </xsd:element>
    </xsd:schema>
  </types>

  <message name="ValidarIdentidadReniecInput">
    <part name="parameters" element="tns:ValidarIdentidadReniecRequest"/>
  </message>
  <message name="ValidarIdentidadReniecOutput">
    <part name="parameters" element="tns:ValidarIdentidadReniecResponse"/>
  </message>
  <message name="ConsultarYReservarStockLegacyInput">
    <part name="parameters" element="tns:ConsultarYReservarStockLegacyRequest"/>
  </message>
  <message name="ConsultarYReservarStockLegacyOutput">
    <part name="parameters" element="tns:ConsultarYReservarStockLegacyResponse"/>
  </message>

  <portType name="InteroperabilidadPortType">
    <operation name="ValidarIdentidadReniec">
      <input message="tns:ValidarIdentidadReniecInput"/>
      <output message="tns:ValidarIdentidadReniecOutput"/>
    </operation>
    <operation name="ConsultarYReservarStockLegacy">
      <input message="tns:ConsultarYReservarStockLegacyInput"/>
      <output message="tns:ConsultarYReservarStockLegacyOutput"/>
    </operation>
  </portType>

  <binding name="InteroperabilidadSOAPBinding" type="tns:InteroperabilidadPortType">
    <soap:binding style="document" transport="http://schemas.xmlsoap.org/soap/http"/>
    <operation name="ValidarIdentidadReniec">
      <soap:operation soapAction="http://bloodlink.minsa.gob.pe/soap/interoperabilidad/ValidarIdentidadReniec"/>
      <input><soap:body use="literal"/></input>
      <output><soap:body use="literal"/></output>
    </operation>
    <operation name="ConsultarYReservarStockLegacy">
      <soap:operation soapAction="http://bloodlink.minsa.gob.pe/soap/interoperabilidad/ConsultarYReservarStockLegacy"/>
      <input><soap:body use="literal"/></input>
      <output><soap:body use="literal"/></output>
    </operation>
  </binding>

  <service name="ServicioBloodLinkSOAP">
    <documentation>Bus SOAP 1.1 Oficial de Interoperabilidad BloodLink - RENIEC y Hospitales Legados EsSalud</documentation>
    <port name="InteroperabilidadPort" binding="tns:InteroperabilidadSOAPBinding">
      <soap:address location="http://localhost:8080/soap/v1/interoperabilidad"/>
    </port>
  </service>
</definitions>`

type SOAPEnvelopeInput struct {
	XMLName xml.Name      `xml:"Envelope"`
	Body    SOAPBodyInput `xml:"Body"`
}

type SOAPBodyInput struct {
	ValidarReniec       *ValidarReniecXMLReq       `xml:"ValidarIdentidadReniecRequest"`
	ConsultarStockLegacy *ConsultarStockLegacyXMLReq `xml:"ConsultarYReservarStockLegacyRequest"`
}

type ValidarReniecXMLReq struct {
	DNI string `xml:"dni"`
}

type ConsultarStockLegacyXMLReq struct {
	BancoID    string `xml:"bancoId"`
	TipoSangre string `xml:"tipoSangre"`
	FactorRh   string `xml:"factorRh"`
	Cantidad   int    `xml:"cantidad"`
}

var dniRegex = regexp.MustCompile(`^\d{8}$`)

// PadronReniecSimulado contiene ciudadanos peruanos para CP-03 y pruebas E2E.
var PadronReniecSimulado = map[string]domain.CiudadanoReniec{
	"76543210": {
		DNI:             "76543210",
		DNIValido:       true,
		NombreCompleto:  "JUAN CARLOS PEREZ GOMEZ",
		FechaNacimiento: "1995-04-12",
		MayorEdad:       true,
	},
	"71234567": {
		DNI:             "71234567",
		DNIValido:       true,
		NombreCompleto:  "MARIA FERNANDA QUISPE ROJAS",
		FechaNacimiento: "1998-08-23",
		MayorEdad:       true,
	},
	"72345678": {
		DNI:             "72345678",
		DNIValido:       true,
		NombreCompleto:  "LUIS ALBERTO VARGAS SEMINARIO",
		FechaNacimiento: "1990-11-05",
		MayorEdad:       true,
	},
	"45678901": {
		DNI:             "45678901",
		DNIValido:       true,
		NombreCompleto:  "ANA LUCIA CASTILLO CHUNGA",
		FechaNacimiento: "1987-02-19",
		MayorEdad:       true,
	},
	"48901234": {
		DNI:             "48901234",
		DNIValido:       true,
		NombreCompleto:  "JORGE EDUARDO PALACIOS MERINO",
		FechaNacimiento: "1993-07-30",
		MayorEdad:       true,
	},
	"70112233": {
		DNI:             "70112233",
		DNIValido:       true,
		NombreCompleto:  "CARMEN ROSA CORDOVA ESPINOZA",
		FechaNacimiento: "2001-01-15",
		MayorEdad:       true,
	},
	"61112222": {
		DNI:             "61112222",
		DNIValido:       true,
		NombreCompleto:  "DIEGO MATEO SANCHEZ HUAMAN",
		FechaNacimiento: "2011-09-10",
		MayorEdad:       false, // Menor de edad (15 años)
	},
}

type SOAPServerHandler struct {
	legacyStore ports.ILegacyStockStore
	baseURL     string
}

func NewSOAPServerHandler(legacyStore ports.ILegacyStockStore, baseURL string) *SOAPServerHandler {
	return &SOAPServerHandler{
		legacyStore: legacyStore,
		baseURL:     baseURL,
	}
}

func (h *SOAPServerHandler) RegisterRoutes(router *gin.Engine, apiV1 *gin.RouterGroup) {
	soapGroup := router.Group("/soap/v1")
	{
		soapGroup.GET("/interoperabilidad", h.GetWSDL)
		soapGroup.POST("/interoperabilidad", h.HandleSOAPRequest)
	}
	apiV1.GET("/uddi/servicios", h.GetUDDIDirectory)
}

func (h *SOAPServerHandler) GetWSDL(c *gin.Context) {
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(BloodLinkWSDL))
}

func (h *SOAPServerHandler) GetUDDIDirectory(c *gin.Context) {
	endpoint := h.baseURL
	if endpoint == "" {
		endpoint = "http://localhost:8080/soap/v1/interoperabilidad"
	}
	directorio := domain.UDDIDirectorio{
		Servicio:     "ServicioBloodLinkSOAP",
		Version:      "1.1.0",
		Namespace:    "http://bloodlink.minsa.gob.pe/soap/interoperabilidad",
		WSDLEndpoint: endpoint + "?wsdl",
		PaginasBlancas: domain.UDDIPaginaBlanca{
			Organizacion: "Ministerio de Salud del Perú (MINSA) - PRONAHEBAS / Red Nacional BloodLink",
			Entidad:      "Dirección General de Donaciones, Trasplantes y Banco de Sangre (DIGDOT)",
			Contacto:     "interoperabilidad@bloodlink.minsa.gob.pe",
			RUC:          "20131373237",
			Jurisdiccion: "Piura - Red Macro Regional Norte",
		},
		PaginasAmarillas: []domain.UDDIPaginaAmarilla{
			{
				Categoria:    "Servicios de Salud y Hemoterapia - Interoperabilidad de Bancos de Sangre",
				CodigoUNSPSC: "85121800",
				Estándar:     "HL7 / SOAP 1.1 / WSDL 1.1",
			},
			{
				Categoria:    "Validación Biométrica y Padrón de Identidad Ciudadana (RENIEC)",
				CodigoUNSPSC: "93151500",
				Estándar:     "SOAP 1.1 / XML Schema",
			},
		},
		PaginasVerdes: []domain.UDDIPaginaVerde{
			{
				Operacion:   "ValidarIdentidadReniec",
				Protocolo:   "SOAP 1.1 over HTTP POST",
				BindingURL:  endpoint,
				SOAPAction:  "http://bloodlink.minsa.gob.pe/soap/interoperabilidad/ValidarIdentidadReniec",
				Descripcion: "Verifica existencia y mayoría de edad de un DNI en el padrón RENIEC",
			},
			{
				Operacion:   "ConsultarYReservarStockLegacy",
				Protocolo:   "SOAP 1.1 over HTTP POST",
				BindingURL:  endpoint,
				SOAPAction:  "http://bloodlink.minsa.gob.pe/soap/interoperabilidad/ConsultarYReservarStockLegacy",
				Descripcion: "Consulta disponibilidad y ejecuta bloqueo transaccional en hospitales legados EsSalud",
			},
		},
	}
	c.JSON(http.StatusOK, directorio)
}

func (h *SOAPServerHandler) HandleSOAPRequest(c *gin.Context) {
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil || len(rawBody) == 0 {
		writeSOAPFault(c, http.StatusBadRequest, "soap:Client", "Cuerpo XML SOAP vacío o ilegible")
		return
	}

	var env SOAPEnvelopeInput
	if err := xml.Unmarshal(rawBody, &env); err != nil {
		writeSOAPFault(c, http.StatusBadRequest, "soap:Client", "Error de sintaxis en sobre XML SOAP: "+err.Error())
		return
	}

	if env.Body.ValidarReniec != nil {
		h.processValidarReniec(c, strings.TrimSpace(env.Body.ValidarReniec.DNI))
		return
	}

	if env.Body.ConsultarStockLegacy != nil {
		h.processConsultarStockLegacy(c, c.Request.Context(), env.Body.ConsultarStockLegacy)
		return
	}

	writeSOAPFault(c, http.StatusBadRequest, "soap:Client", "Operación SOAP no reconocida en soap:Body")
}

func (h *SOAPServerHandler) processValidarReniec(c *gin.Context, dni string) {
	if !dniRegex.MatchString(dni) {
		writeSOAPFault(c, http.StatusBadRequest, "soap:Client", "ReniecFault: Formato de DNI inválido (debe tener 8 dígitos)")
		return
	}

	ciudadano, exists := PadronReniecSimulado[dni]
	if !exists {
		writeSOAPFault(c, http.StatusNotFound, "soap:Client", fmt.Sprintf("ReniecFault: El DNI %s no existe en el padrón nacional RENIEC", dni))
		return
	}

	respXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tns="http://bloodlink.minsa.gob.pe/soap/interoperabilidad">
  <soap:Body>
    <tns:ValidarIdentidadReniecResponse>
      <dni>%s</dni>
      <dniValido>%t</dniValido>
      <nombreCompleto>%s</nombreCompleto>
      <fechaNacimiento>%s</fechaNacimiento>
      <mayorEdad>%t</mayorEdad>
    </tns:ValidarIdentidadReniecResponse>
  </soap:Body>
</soap:Envelope>`, ciudadano.DNI, ciudadano.DNIValido, ciudadano.NombreCompleto, ciudadano.FechaNacimiento, ciudadano.MayorEdad)

	c.Data(http.StatusOK, "text/xml; charset=utf-8", []byte(respXML))
}

func (h *SOAPServerHandler) processConsultarStockLegacy(c *gin.Context, ctx context.Context, req *ConsultarStockLegacyXMLReq) {
	if req.BancoID == "" || req.TipoSangre == "" || req.FactorRh == "" || req.Cantidad <= 0 {
		writeSOAPFault(c, http.StatusBadRequest, "soap:Client", "LegacyStockFault: Parámetros incompletos para reserva SOAP")
		return
	}

	if h.legacyStore == nil {
		shortUUID := strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", "")[:8])
		txID := fmt.Sprintf("RES-%d-%s", time.Now().Year(), shortUUID)
		respXML := buildLegacyStockSOAPResponse(true, req.BancoID, txID, []string{"UND-SOAP-01"}, []string{"BL-SOAP-01"}, "Reserva SOAP simulada exitosa")
		c.Data(http.StatusOK, "text/xml; charset=utf-8", []byte(respXML))
		return
	}

	res, err := h.legacyStore.ReservarUnidadesLegacyEnBD(ctx, req.BancoID, req.TipoSangre, req.FactorRh, req.Cantidad)
	if err != nil {
		writeSOAPFault(c, http.StatusConflict, "soap:Server", "LegacyStockFault: "+err.Error())
		return
	}

	respXML := buildLegacyStockSOAPResponse(res.Exito, res.BancoID, res.CodigoTransaccion, res.UnidadesIDs, res.CodigosBarras, res.Mensaje)
	c.Data(http.StatusOK, "text/xml; charset=utf-8", []byte(respXML))
}

func buildLegacyStockSOAPResponse(exito bool, bancoID, txID string, unidades, codigos []string, mensaje string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tns="http://bloodlink.minsa.gob.pe/soap/interoperabilidad">
  <soap:Body>
    <tns:ConsultarYReservarStockLegacyResponse>
      <exito>%t</exito>
      <bancoId>%s</bancoId>
      <codigoTransaccion>%s</codigoTransaccion>
      <unidadesReservadas>%s</unidadesReservadas>
      <codigosBarras>%s</codigosBarras>
      <mensaje>%s</mensaje>
    </tns:ConsultarYReservarStockLegacyResponse>
  </soap:Body>
</soap:Envelope>`, exito, bancoID, txID, strings.Join(unidades, ","), strings.Join(codigos, ","), mensaje)
}

func writeSOAPFault(c *gin.Context, statusCode int, faultCode, faultString string) {
	faultXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <soap:Fault>
      <faultcode>%s</faultcode>
      <faultstring>%s</faultstring>
    </soap:Fault>
  </soap:Body>
</soap:Envelope>`, faultCode, faultString)
	c.Data(statusCode, "text/xml; charset=utf-8", []byte(faultXML))
}
