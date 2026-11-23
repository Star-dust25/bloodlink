package soap

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/bloodlink/backend/internal/modules/integracion/domain"
)

var clientDNIRegex = regexp.MustCompile(`^\d{8}$`)

type SOAPClientAdapter struct {
	endpointURL string
	httpClient  *http.Client
}

func NewSOAPClientAdapter(endpointURL string) *SOAPClientAdapter {
	return &SOAPClientAdapter{
		endpointURL: endpointURL,
		httpClient: &http.Client{
			Timeout: 4 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

type soapResponseEnvelope struct {
	XMLName xml.Name         `xml:"Envelope"`
	Body    soapResponseBody `xml:"Body"`
}

type soapResponseBody struct {
	ReniecResp *struct {
		DNI             string `xml:"dni"`
		DNIValido       bool   `xml:"dniValido"`
		NombreCompleto  string `xml:"nombreCompleto"`
		FechaNacimiento string `xml:"fechaNacimiento"`
		MayorEdad       bool   `xml:"mayorEdad"`
	} `xml:"ValidarIdentidadReniecResponse"`
	LegacyResp *struct {
		Exito              bool   `xml:"exito"`
		BancoID            string `xml:"bancoId"`
		CodigoTransaccion  string `xml:"codigoTransaccion"`
		UnidadesReservadas string `xml:"unidadesReservadas"`
		CodigosBarras      string `xml:"codigosBarras"`
		Mensaje            string `xml:"mensaje"`
	} `xml:"ConsultarYReservarStockLegacyResponse"`
	Fault *struct {
		FaultCode   string `xml:"faultcode"`
		FaultString string `xml:"faultstring"`
	} `xml:"Fault"`
}

func (c *SOAPClientAdapter) ValidarIdentidad(ctx context.Context, dni string) (*domain.CiudadanoReniec, error) {
	dni = strings.TrimSpace(dni)
	if !clientDNIRegex.MatchString(dni) {
		return nil, domain.ErrDNIFormatoInvalido
	}

	envelope := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tns="http://bloodlink.minsa.gob.pe/soap/interoperabilidad">
  <soap:Body>
    <tns:ValidarIdentidadReniecRequest>
      <dni>%s</dni>
    </tns:ValidarIdentidadReniecRequest>
  </soap:Body>
</soap:Envelope>`, dni)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpointURL, bytes.NewBufferString(envelope))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://bloodlink.minsa.gob.pe/soap/interoperabilidad/ValidarIdentidadReniec")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error invocando bus SOAP RENIEC: %w", err)
	}
	defer resp.Body.Close()

	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed soapResponseEnvelope
	if err := xml.Unmarshal(rawResp, &parsed); err != nil {
		return nil, fmt.Errorf("error parseando XML SOAP RENIEC: %w", err)
	}

	if parsed.Body.Fault != nil {
		if strings.Contains(parsed.Body.Fault.FaultString, "no existe") {
			return nil, domain.ErrDNINoEncontrado
		}
		return nil, fmt.Errorf("SOAP Fault: %s", parsed.Body.Fault.FaultString)
	}

	if parsed.Body.ReniecResp == nil {
		return nil, domain.ErrDNINoEncontrado
	}

	ciudadano := &domain.CiudadanoReniec{
		DNI:             parsed.Body.ReniecResp.DNI,
		DNIValido:       parsed.Body.ReniecResp.DNIValido,
		NombreCompleto:  parsed.Body.ReniecResp.NombreCompleto,
		FechaNacimiento: parsed.Body.ReniecResp.FechaNacimiento,
		MayorEdad:       parsed.Body.ReniecResp.MayorEdad,
	}

	if !ciudadano.MayorEdad {
		return ciudadano, domain.ErrDNIMenorDeEdad
	}

	return ciudadano, nil
}

func (c *SOAPClientAdapter) ConsultarYReservarStockLegacy(ctx context.Context, bancoID, tipoSangre, factorRh string, cantidad int) (*domain.ReservaLegacyResultado, error) {
	envelope := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tns="http://bloodlink.minsa.gob.pe/soap/interoperabilidad">
  <soap:Body>
    <tns:ConsultarYReservarStockLegacyRequest>
      <bancoId>%s</bancoId>
      <tipoSangre>%s</tipoSangre>
      <factorRh>%s</factorRh>
      <cantidad>%d</cantidad>
    </tns:ConsultarYReservarStockLegacyRequest>
  </soap:Body>
</soap:Envelope>`, bancoID, tipoSangre, factorRh, cantidad)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpointURL, bytes.NewBufferString(envelope))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://bloodlink.minsa.gob.pe/soap/interoperabilidad/ConsultarYReservarStockLegacy")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error invocando SOAP Hospital Legado: %w", err)
	}
	defer resp.Body.Close()

	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed soapResponseEnvelope
	if err := xml.Unmarshal(rawResp, &parsed); err != nil {
		return nil, fmt.Errorf("error parseando respuesta SOAP Hospital Legado: %w", err)
	}

	if parsed.Body.Fault != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrLegacyStockFaltante, parsed.Body.Fault.FaultString)
	}

	if parsed.Body.LegacyResp == nil || !parsed.Body.LegacyResp.Exito {
		return nil, domain.ErrLegacyStockFaltante
	}

	var unidades []string
	if parsed.Body.LegacyResp.UnidadesReservadas != "" {
		unidades = strings.Split(parsed.Body.LegacyResp.UnidadesReservadas, ",")
	}
	var codigos []string
	if parsed.Body.LegacyResp.CodigosBarras != "" {
		codigos = strings.Split(parsed.Body.LegacyResp.CodigosBarras, ",")
	}

	return &domain.ReservaLegacyResultado{
		Exito:             true,
		BancoID:           parsed.Body.LegacyResp.BancoID,
		CodigoTransaccion: parsed.Body.LegacyResp.CodigoTransaccion,
		UnidadesIDs:       unidades,
		CodigosBarras:     codigos,
		Mensaje:           parsed.Body.LegacyResp.Mensaje,
	}, nil
}
