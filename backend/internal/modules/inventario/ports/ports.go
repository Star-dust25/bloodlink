package ports

import (
	"context"

	"github.com/bloodlink/backend/internal/modules/inventario/domain"
)

type FiltroStock struct {
	TipoSangre string
	FactorRh   string
	Estado     string
	HospitalID string
}

type BancoStockResumenDTO struct {
	BancoID               string `json:"bancoId"`
	Nombre                string `json:"nombre"`
	TipoInstitucion       string `json:"tipoInstitucion"`
	Coordenadas           string `json:"coordenadas"`
	Jurisdiccion          string `json:"jurisdiccion"`
	EsLegacySoap          bool   `json:"esLegacySoap"`
	UnidadesDisponibles   int    `json:"unidadesDisponibles"`
	ProximasACaducar7Dias int    `json:"proximasACaducar7Dias"`
}

type StockRedResponseDTO struct {
	TipoSangre            string                 `json:"tipoSangre"`
	FactorRh              string                 `json:"factorRh"`
	StockGlobalDisponible int                    `json:"stockGlobalDisponible"`
	NivelAbastecimiento   string                 `json:"nivelAbastecimiento"`
	Bancos                []BancoStockResumenDTO `json:"bancos"`
	ConsultadoEn          string                 `json:"consultadoEn"`
}

type RegistrarUnidadRequest struct {
	BancoID         string `json:"bancoId" binding:"required"`
	TipoSangre      string `json:"tipoSangre" binding:"required"`
	FactorRh        string `json:"factorRh" binding:"required"`
	VolumenMl       int    `json:"volumenMl"`
	FechaExtraccion string `json:"fechaExtraccion" binding:"required"`
	Origen          string `json:"origen" binding:"required"`
}

type UnidadSangreResponseDTO struct {
	UnidadID        string `json:"unidadId"`
	CodigoBarras    string `json:"codigoBarras"`
	BancoID         string `json:"bancoId"`
	TipoSangre      string `json:"tipoSangre"`
	FactorRh        string `json:"factorRh"`
	VolumenMl       int    `json:"volumenMl"`
	Origen          string `json:"origen"`
	FechaExtraccion string `json:"fechaExtraccion"`
	FechaCaducidad  string `json:"fechaCaducidad"`
	Estado          string `json:"estado"`
	AlertaCaducidad string `json:"alertaCaducidad"`
	DiasRestantes   int    `json:"diasRestantes"`
}

type ActualizarEstadoRequest struct {
	NuevoEstado string `json:"nuevoEstado" binding:"required"`
	Motivo      string `json:"motivo"`
}

// IGestorInventario defines the inbound port for MS-Inventario (Sección 7.4.1 I_GestorInventario).
type IGestorInventario interface {
	ConsultarStockRed(ctx context.Context, filtro FiltroStock) (*StockRedResponseDTO, error)
	ListarUnidades(ctx context.Context, filtro FiltroStock) ([]UnidadSangreResponseDTO, error)
	ObtenerUnidad(ctx context.Context, unidadID string) (*UnidadSangreResponseDTO, error)
	RegistrarUnidad(ctx context.Context, req RegistrarUnidadRequest, usuarioID string) (*UnidadSangreResponseDTO, error)
	ActualizarEstado(ctx context.Context, unidadID string, req ActualizarEstadoRequest, usuarioID string) (*UnidadSangreResponseDTO, error)
}

// IInventarioRepository defines the outbound persistence port for banks, blood units, and audit logs.
type IInventarioRepository interface {
	ListarBancos(ctx context.Context, hospitalID string) ([]domain.BancoSangre, error)
	ObtenerBancoPorID(ctx context.Context, bancoID string) (*domain.BancoSangre, error)
	ListarUnidades(ctx context.Context, filtro FiltroStock) ([]domain.UnidadSangre, error)
	ObtenerUnidadPorID(ctx context.Context, unidadID string) (*domain.UnidadSangre, error)
	CrearUnidadConAuditoria(ctx context.Context, u *domain.UnidadSangre, usuarioID, detalle string) error
	ActualizarEstadoConAuditoria(ctx context.Context, u *domain.UnidadSangre, estadoAnterior, usuarioID, motivo string) error
}
