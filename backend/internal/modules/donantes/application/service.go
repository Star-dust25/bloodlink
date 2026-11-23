package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bloodlink/backend/internal/modules/donantes/domain"
	"github.com/bloodlink/backend/internal/modules/donantes/ports"
	integDomain "github.com/bloodlink/backend/internal/modules/integracion/domain"
	integPorts "github.com/bloodlink/backend/internal/modules/integracion/ports"
	"github.com/google/uuid"
)

type SrvRegistrarDonante struct {
	repo   ports.IDonanteRepository
	reniec integPorts.IReniec
	crypto ports.ICryptoService
}

func NewSrvRegistrarDonante(repo ports.IDonanteRepository, reniec integPorts.IReniec, crypto ports.ICryptoService) *SrvRegistrarDonante {
	return &SrvRegistrarDonante{
		repo:   repo,
		reniec: reniec,
		crypto: crypto,
	}
}

func (s *SrvRegistrarDonante) ValidarDNIReniec(ctx context.Context, dni string) (*integDomain.CiudadanoReniec, error) {
	return s.reniec.ValidarIdentidad(ctx, strings.TrimSpace(dni))
}

func (s *SrvRegistrarDonante) RegistrarDonante(ctx context.Context, req ports.RegistrarDonanteRequest) (*ports.DonanteResponseDTO, error) {
	if !req.AceptaPoliticaPrivacidad {
		return nil, domain.ErrPoliticaPrivacidadRequerida
	}

	dniClean := strings.TrimSpace(req.DNI)
	ciudadano, err := s.reniec.ValidarIdentidad(ctx, dniClean)
	if err != nil {
		return nil, err
	}

	dniHash := s.crypto.HashSHA256(dniClean)
	existe, err := s.repo.ExistePorDNIHash(ctx, dniHash)
	if err != nil {
		return nil, err
	}
	if existe {
		return nil, domain.ErrDonanteYaRegistrado
	}

	now := time.Now()
	shortUUID := strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", "")[:8])
	donanteID := fmt.Sprintf("DON-%d-%s", now.Year(), shortUUID)

	donante, err := domain.NuevoDonante(
		donanteID,
		dniClean,
		ciudadano.NombreCompleto,
		strings.TrimSpace(req.Celular),
		strings.TrimSpace(req.Correo),
		strings.ToUpper(strings.TrimSpace(req.TipoSangre)),
		strings.TrimSpace(req.FactorRh),
		strings.TrimSpace(req.BancoPreferidoID),
		req.AceptaPoliticaPrivacidad,
		now,
	)
	if err != nil {
		return nil, err
	}

	dniEnc, err := s.crypto.Encrypt(donante.DocumentoIdentidad)
	if err != nil {
		return nil, err
	}
	celEnc, err := s.crypto.Encrypt(donante.Celular)
	if err != nil {
		return nil, err
	}

	donante.DNICifrado = dniEnc
	donante.DNIHash = dniHash
	donante.CelularCifrado = celEnc

	if err := s.repo.Guardar(ctx, donante); err != nil {
		return nil, err
	}

	return &ports.DonanteResponseDTO{
		DonanteID:               donante.ID,
		DNI:                     donante.DocumentoIdentidad,
		NombreCompleto:          donante.NombreCompleto,
		Celular:                 domain.EnmascararCelular(donante.Celular),
		Correo:                  donante.Correo,
		TipoSangre:              donante.TipoSangre,
		FactorRh:                donante.FactorRh,
		BancoPreferidoID:        donante.BancoPreferidoID,
		ValidadoReniec:          donante.ValidadoReniec,
		HabilitadoConvocatorias: donante.HabilitadoConvocatorias,
		FechaRegistro:           donante.FechaRegistro.Format(time.RFC3339),
	}, nil
}

func (s *SrvRegistrarDonante) ListarDonantes(ctx context.Context, tipoSangre, factorRh string) ([]ports.DonanteResponseDTO, error) {
	list, err := s.repo.ListarActivos(ctx, tipoSangre, factorRh)
	if err != nil {
		return nil, err
	}

	res := make([]ports.DonanteResponseDTO, 0, len(list))
	for _, d := range list {
		dniPlain, errDNI := s.crypto.Decrypt(d.DNICifrado)
		if errDNI != nil {
			dniPlain = "00000000"
		}
		celPlain, errCel := s.crypto.Decrypt(d.CelularCifrado)
		if errCel != nil {
			celPlain = "+51900000000"
		}

		res = append(res, ports.DonanteResponseDTO{
			DonanteID:               d.ID,
			DNI:                     domain.EnmascararDNI(dniPlain),
			NombreCompleto:          d.NombreCompleto,
			Celular:                 domain.EnmascararCelular(celPlain),
			Correo:                  d.Correo,
			TipoSangre:              d.TipoSangre,
			FactorRh:                d.FactorRh,
			BancoPreferidoID:        d.BancoPreferidoID,
			ValidadoReniec:          d.ValidadoReniec,
			HabilitadoConvocatorias: d.HabilitadoConvocatorias,
			FechaRegistro:           d.FechaRegistro.Format(time.RFC3339),
		})
	}
	return res, nil
}

func (s *SrvRegistrarDonante) ObtenerPerfilDonante(ctx context.Context, donanteID string) (*ports.DonanteResponseDTO, error) {
	d, err := s.repo.ObtenerPorID(ctx, donanteID)
	if err != nil {
		return nil, err
	}
	dniPlain, _ := s.crypto.Decrypt(d.DNICifrado)
	celPlain, _ := s.crypto.Decrypt(d.CelularCifrado)

	return &ports.DonanteResponseDTO{
		DonanteID:               d.ID,
		DNI:                     domain.EnmascararDNI(dniPlain),
		NombreCompleto:          d.NombreCompleto,
		Celular:                 domain.EnmascararCelular(celPlain),
		Correo:                  d.Correo,
		TipoSangre:              d.TipoSangre,
		FactorRh:                d.FactorRh,
		BancoPreferidoID:        d.BancoPreferidoID,
		ValidadoReniec:          d.ValidadoReniec,
		HabilitadoConvocatorias: d.HabilitadoConvocatorias,
		FechaRegistro:           d.FechaRegistro.Format(time.RFC3339),
	}, nil
}

func (s *SrvRegistrarDonante) ActualizarContacto(ctx context.Context, donanteID string, req ports.ActualizarContactoDonanteRequest) (*ports.DonanteResponseDTO, error) {
	d, err := s.repo.ObtenerPorID(ctx, donanteID)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.Celular) != "" {
		encCel, err := s.crypto.Encrypt(strings.TrimSpace(req.Celular))
		if err != nil {
			return nil, err
		}
		d.CelularCifrado = encCel
	}
	if strings.TrimSpace(req.Correo) != "" {
		d.Correo = strings.TrimSpace(req.Correo)
	}
	if strings.TrimSpace(req.BancoPreferidoID) != "" {
		d.BancoPreferidoID = strings.TrimSpace(req.BancoPreferidoID)
	}
	if req.HabilitadoConvocatorias != nil {
		d.HabilitadoConvocatorias = *req.HabilitadoConvocatorias
	}

	if err := s.repo.Actualizar(ctx, d); err != nil {
		return nil, err
	}
	return s.ObtenerPerfilDonante(ctx, donanteID)
}

func (s *SrvRegistrarDonante) DarDeBaja(ctx context.Context, donanteID string) error {
	return s.repo.DarDeBajaLogica(ctx, donanteID)
}
