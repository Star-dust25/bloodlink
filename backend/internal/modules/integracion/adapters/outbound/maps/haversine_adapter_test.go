package maps

import (
	"context"
	"errors"
	"testing"

	"github.com/bloodlink/backend/internal/modules/integracion/domain"
)

func TestHaversineMapAdapter_CalculoYCache(t *testing.T) {
	cache := NewInMemoryGeoCache()
	adapter := NewHaversineMapAdapter(cache)
	ctx := context.Background()

	// Distancia entre Hospital Santa Rosa (-5.1945,-80.6328) y Hospital Cayetano Heredia (-5.1821,-80.6190)
	ruta, err := adapter.CalcularDistanciaYTiempo(ctx, "-5.1945,-80.6328", -5.1821, -80.6190, false)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if ruta.DistanciaKm < 1.5 || ruta.DistanciaKm > 3.5 {
		t.Errorf("distancia esperada ~2.49 km, obtuvo %.2f km", ruta.DistanciaKm)
	}
	if ruta.TiempoEstimadoMin <= 0 {
		t.Errorf("tiempo estimado inválido: %d", ruta.TiempoEstimadoMin)
	}

	// Segunda llamada debe venir de caché
	rutaCached, err := adapter.CalcularDistanciaYTiempo(ctx, "-5.1945,-80.6328", -5.1821, -80.6190, false)
	if err != nil || rutaCached.DistanciaKm != ruta.DistanciaKm {
		t.Fatalf("lectura de caché falló: %v", err)
	}

	// Simulación de Contingencia E1 (caída de API de mapas)
	_, errFallo := adapter.CalcularDistanciaYTiempo(ctx, "-5.1945,-80.6328", -5.1821, -80.6190, true)
	if !errors.Is(errFallo, domain.ErrMapasNoDisponible) {
		t.Fatalf("esperaba ErrMapasNoDisponible, obtuvo %v", errFallo)
	}
}
