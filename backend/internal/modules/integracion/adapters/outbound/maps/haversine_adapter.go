package maps

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/bloodlink/backend/internal/modules/integracion/domain"
	"github.com/bloodlink/backend/internal/modules/integracion/ports"
)

const (
	earthRadiusKm          = 6371.0
	urbanTortuosityFactor  = 1.22
	avgAmbulanceSpeedKmH   = 45.0
	minDispatchOverheadMin = 3
)

type InMemoryGeoCache struct {
	mu    sync.RWMutex
	store map[string]*domain.CalculoRuta
}

func NewInMemoryGeoCache() *InMemoryGeoCache {
	return &InMemoryGeoCache{
		store: make(map[string]*domain.CalculoRuta),
	}
}

func (c *InMemoryGeoCache) Get(key string) (*domain.CalculoRuta, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.store[key]
	return val, ok
}

func (c *InMemoryGeoCache) Set(key string, ruta *domain.CalculoRuta) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = ruta
}

type HaversineMapAdapter struct {
	cache ports.IGeoCache
}

func NewHaversineMapAdapter(cache ports.IGeoCache) *HaversineMapAdapter {
	return &HaversineMapAdapter{cache: cache}
}

func (a *HaversineMapAdapter) CalcularDistanciaYTiempo(_ context.Context, coordOrigen string, latDestino, lonDestino float64, simularFallo bool) (*domain.CalculoRuta, error) {
	if simularFallo {
		return nil, domain.ErrMapasNoDisponible
	}

	latOrigen, lonOrigen, err := ParseCoordenadas(coordOrigen)
	if err != nil {
		// Default fallback to Hospital Santa Rosa Piura if coordinates string is malformed
		latOrigen, lonOrigen = -5.1945, -80.6328
	}

	cacheKey := fmt.Sprintf("%.4f,%.4f->%.4f,%.4f", latOrigen, lonOrigen, latDestino, lonDestino)
	if a.cache != nil {
		if cached, ok := a.cache.Get(cacheKey); ok {
			return cached, nil
		}
	}

	rawDist := haversineKm(latOrigen, lonOrigen, latDestino, lonDestino)
	roadDist := math.Round(rawDist*urbanTortuosityFactor*100) / 100
	if roadDist < 0.1 {
		roadDist = 0.1
	}

	tiempoMin := int(math.Ceil((roadDist/avgAmbulanceSpeedKmH)*60.0)) + minDispatchOverheadMin
	res := &domain.CalculoRuta{
		DistanciaKm:       roadDist,
		TiempoEstimadoMin: tiempoMin,
	}

	if a.cache != nil {
		a.cache.Set(cacheKey, res)
	}
	return res, nil
}

func ParseCoordenadas(coord string) (float64, float64, error) {
	parts := strings.Split(strings.TrimSpace(coord), ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("formato de coordenadas inválido: %s", coord)
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, err
	}
	lon, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, err
	}
	return lat, lon, nil
}

func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	lat1Rad := lat1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}
