package ports

import (
	"context"

	"github.com/bloodlink/backend/internal/modules/integracion/domain"
)

// IMapas defines the outbound geospatial calculation port (Sección 7.4.1 I_Mapas).
type IMapas interface {
	CalcularDistanciaYTiempo(ctx context.Context, coordOrigen string, latDestino, lonDestino float64, simularFallo bool) (*domain.CalculoRuta, error)
}

// IGeoCache defines the thread-safe cache port for geospatial distance calculations (RNF-01 / Riesgo R1).
type IGeoCache interface {
	Get(key string) (*domain.CalculoRuta, bool)
	Set(key string, ruta *domain.CalculoRuta)
}

// IReniec defines the outbound port for citizen identity validation via SOAP 1.1 (Sección 7.4.1 I_Reniec).
type IReniec interface {
	ValidarIdentidad(ctx context.Context, dni string) (*domain.CiudadanoReniec, error)
}

// ILegacySoapGateway defines the outbound port to query and reserve stock in legacy SOAP hospitals (EsSalud).
type ILegacySoapGateway interface {
	ConsultarYReservarStockLegacy(ctx context.Context, bancoID, tipoSangre, factorRh string, cantidad int) (*domain.ReservaLegacyResultado, error)
}

// ILegacyStockStore defines the database port used by the inbound SOAP server to lock units in legacy banks.
type ILegacyStockStore interface {
	ReservarUnidadesLegacyEnBD(ctx context.Context, bancoID, tipoSangre, factorRh string, cantidad int) (*domain.ReservaLegacyResultado, error)
}
