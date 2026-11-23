package soap

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bloodlink/backend/internal/modules/integracion/domain"
	"github.com/gin-gonic/gin"
)

type mockLegacyStore struct{}

func (m *mockLegacyStore) ReservarUnidadesLegacyEnBD(_ context.Context, bancoID, _, _ string, _ int) (*domain.ReservaLegacyResultado, error) {
	if bancoID == "SIN-STOCK" {
		return nil, domain.ErrLegacyStockFaltante
	}
	return &domain.ReservaLegacyResultado{
		Exito:             true,
		BancoID:           bancoID,
		CodigoTransaccion: "SOAP-TX-TEST",
		UnidadesIDs:       []string{"UND-2026-0401", "UND-2026-0402"},
		CodigosBarras:     []string{"BL-ESSALUD04-01", "BL-ESSALUD04-02"},
		Mensaje:           "Reserva exitosa",
	}, nil
}

func TestSOAPServer_WSDL_Reniec_Legacy_Y_UDDI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	apiV1 := router.Group("/api/v1")

	handler := NewSOAPServerHandler(&mockLegacyStore{}, "http://localhost:8080/soap/v1/interoperabilidad")
	handler.RegisterRoutes(router, apiV1)

	// 1. WSDL Endpoint (httptest.NewRecorder no requiere abrir socket TCP)
	wWSDL := httptest.NewRecorder()
	reqWSDL, _ := http.NewRequest(http.MethodGet, "/soap/v1/interoperabilidad?wsdl", nil)
	router.ServeHTTP(wWSDL, reqWSDL)
	if wWSDL.Code != http.StatusOK || !strings.Contains(wWSDL.Body.String(), "ServicioBloodLinkSOAP") {
		t.Fatalf("GET WSDL falló: status=%d", wWSDL.Code)
	}

	// 2. UDDI Endpoint
	wUDDI := httptest.NewRecorder()
	reqUDDI, _ := http.NewRequest(http.MethodGet, "/api/v1/uddi/servicios", nil)
	router.ServeHTTP(wUDDI, reqUDDI)
	if wUDDI.Code != http.StatusOK || !strings.Contains(wUDDI.Body.String(), "PaginasVerdes") && !strings.Contains(wUDDI.Body.String(), "paginasVerdes") {
		t.Fatalf("GET UDDI falló: status=%d", wUDDI.Code)
	}

	// 3. SOAP POST ValidarIdentidadReniec (DNI 76543210 -> JUAN CARLOS PEREZ GOMEZ)
	soapReniecOK := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tns="http://bloodlink.minsa.gob.pe/soap/interoperabilidad">
  <soap:Body>
    <tns:ValidarIdentidadReniecRequest>
      <dni>76543210</dni>
    </tns:ValidarIdentidadReniecRequest>
  </soap:Body>
</soap:Envelope>`
	wReniec := httptest.NewRecorder()
	reqReniec, _ := http.NewRequest(http.MethodPost, "/soap/v1/interoperabilidad", bytes.NewBufferString(soapReniecOK))
	router.ServeHTTP(wReniec, reqReniec)
	if wReniec.Code != http.StatusOK || !strings.Contains(wReniec.Body.String(), "JUAN CARLOS PEREZ GOMEZ") {
		t.Fatalf("SOAP ValidarIdentidadReniec 76543210 falló: %d %s", wReniec.Code, wReniec.Body.String())
	}

	// 4. SOAP POST ValidarIdentidadReniec con DNI inexistente -> SOAP Fault 404
	soapReniec404 := strings.ReplaceAll(soapReniecOK, "76543210", "99999999")
	w404 := httptest.NewRecorder()
	req404, _ := http.NewRequest(http.MethodPost, "/soap/v1/interoperabilidad", bytes.NewBufferString(soapReniec404))
	router.ServeHTTP(w404, req404)
	if w404.Code != http.StatusNotFound || !strings.Contains(w404.Body.String(), "soap:Fault") {
		t.Fatalf("esperaba 404 soap:Fault, obtuvo %d", w404.Code)
	}

	// 5. SOAP POST ConsultarYReservarStockLegacy exitoso
	soapLegacyOK := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tns="http://bloodlink.minsa.gob.pe/soap/interoperabilidad">
  <soap:Body>
    <tns:ConsultarYReservarStockLegacyRequest>
      <bancoId>HOSP-ESSALUD-04</bancoId>
      <tipoSangre>O</tipoSangre>
      <factorRh>+</factorRh>
      <cantidad>2</cantidad>
    </tns:ConsultarYReservarStockLegacyRequest>
  </soap:Body>
</soap:Envelope>`
	wLeg := httptest.NewRecorder()
	reqLeg, _ := http.NewRequest(http.MethodPost, "/soap/v1/interoperabilidad", bytes.NewBufferString(soapLegacyOK))
	router.ServeHTTP(wLeg, reqLeg)
	if wLeg.Code != http.StatusOK || !strings.Contains(wLeg.Body.String(), "UND-2026-0401") {
		t.Fatalf("SOAP ConsultarYReservarStockLegacy falló: %d %s", wLeg.Code, wLeg.Body.String())
	}

	// 6. XML malformado, body vacío y DNI con formato inválido -> SOAP Fault 400
	wBad := httptest.NewRecorder()
	reqBad, _ := http.NewRequest(http.MethodPost, "/soap/v1/interoperabilidad", bytes.NewBufferString("<invalid>"))
	router.ServeHTTP(wBad, reqBad)
	if wBad.Code != http.StatusBadRequest {
		t.Errorf("esperaba 400 ante XML inválido, obtuvo %d", wBad.Code)
	}

	wEmpty := httptest.NewRecorder()
	reqEmpty, _ := http.NewRequest(http.MethodPost, "/soap/v1/interoperabilidad", bytes.NewBufferString(""))
	router.ServeHTTP(wEmpty, reqEmpty)
	if wEmpty.Code != http.StatusBadRequest {
		t.Errorf("esperaba 400 ante body vacío, obtuvo %d", wEmpty.Code)
	}

	soapBadDNI := strings.ReplaceAll(soapReniecOK, "76543210", "123")
	wBadDNI := httptest.NewRecorder()
	reqBadDNI, _ := http.NewRequest(http.MethodPost, "/soap/v1/interoperabilidad", bytes.NewBufferString(soapBadDNI))
	router.ServeHTTP(wBadDNI, reqBadDNI)
	if wBadDNI.Code != http.StatusBadRequest {
		t.Errorf("esperaba 400 ante DNI de 3 dígitos, obtuvo %d", wBadDNI.Code)
	}

	soapLegacyFail := strings.ReplaceAll(soapLegacyOK, "HOSP-ESSALUD-04", "SIN-STOCK")
	wLegFail := httptest.NewRecorder()
	reqLegFail, _ := http.NewRequest(http.MethodPost, "/soap/v1/interoperabilidad", bytes.NewBufferString(soapLegacyFail))
	router.ServeHTTP(wLegFail, reqLegFail)
	if wLegFail.Code != http.StatusConflict {
		t.Errorf("esperaba 409 ante fallo de stock legacy SOAP, obtuvo %d", wLegFail.Code)
	}
}
