package database

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/bloodlink/backend/internal/config"
	"github.com/bloodlink/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitPostgres(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("error conectando a PostgreSQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo sql.DB: %w", err)
	}

	maxOpen := cfg.DBMaxOpenConns
	if maxOpen <= 0 || maxOpen > 40 {
		maxOpen = 40
	}
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifetimeM) * time.Minute)

	if err := AutoMigrateAndSeed(db, cfg.AESKeyHex); err != nil {
		return nil, err
	}

	return db, nil
}

func AutoMigrateAndSeed(db *gorm.DB, aesKeyHex string) error {
	if err := db.AutoMigrate(
		&models.BancoSangreORM{},
		&models.UsuarioPersonalORM{},
		&models.UnidadSangreORM{},
		&models.SolicitudEmergenciaORM{},
		&models.SolicitudUnidadORM{},
		&models.DonanteORM{},
		&models.AuditLogORM{},
		&models.NotificacionLogORM{},
	); err != nil {
		return fmt.Errorf("error en AutoMigrate: %w", err)
	}

	return SeedDatabase(db, aesKeyHex)
}

func SeedDatabase(db *gorm.DB, aesKeyHex string) error {
	now := time.Now().Truncate(24 * time.Hour)

	// 1. Bancos de Sangre en Piura (CP-01 a CP-05)
	bancos := []models.BancoSangreORM{
		{
			ID:              "HOSP-MINSA-01",
			Nombre:          "Hospital Nacional Santa Rosa - Piura",
			Direccion:       "Av. Grau Km 3.5, Veintiséis de Octubre, Piura",
			Latitud:         -5.1945,
			Longitud:        -80.6328,
			Coordenadas:     "-5.1945,-80.6328",
			Jurisdiccion:    "Piura Oeste",
			TipoInstitucion: "MINSA",
			EsLegacySoap:    false,
		},
		{
			ID:              "HOSP-ESSALUD-04",
			Nombre:          "Hospital III José Cayetano Heredia - EsSalud",
			Direccion:       "Av. Independencia s/n, Urb. Miraflores, Castilla, Piura",
			Latitud:         -5.1821,
			Longitud:        -80.6190,
			Coordenadas:     "-5.1821,-80.6190",
			Jurisdiccion:    "Castilla",
			TipoInstitucion: "EsSalud",
			EsLegacySoap:    true,
		},
		{
			ID:              "HOSP-MINSA-02",
			Nombre:          "Hospital de Apoyo II-2 Sullana",
			Direccion:       "Av. Santa Rosa s/n, Sullana, Piura",
			Latitud:         -4.9039,
			Longitud:        -80.6853,
			Coordenadas:     "-4.9039,-80.6853",
			Jurisdiccion:    "Sullana",
			TipoInstitucion: "MINSA",
			EsLegacySoap:    false,
		},
		{
			ID:              "CLIN-PRIV-01",
			Nombre:          "Clínica Miraflores - Piura",
			Direccion:       "Calle Las Dalias A-12, Urb. Miraflores, Castilla, Piura",
			Latitud:         -5.1880,
			Longitud:        -80.6250,
			Coordenadas:     "-5.1880,-80.6250",
			Jurisdiccion:    "Castilla",
			TipoInstitucion: "Privada",
			EsLegacySoap:    false,
		},
	}

	for _, b := range bancos {
		var count int64
		db.Model(&models.BancoSangreORM{}).Where("id = ?", b.ID).Count(&count)
		if count == 0 {
			if err := db.Create(&b).Error; err != nil {
				return fmt.Errorf("error insertando banco %s: %w", b.ID, err)
			}
		}
	}

	// 2. Usuarios de Prueba (RF-11)
	seedUsers := []struct {
		ID       uint
		Nombre   string
		Doc      string
		Email    string
		Password string
		Rol      string
		Hospital string
	}{
		{
			ID:       12,
			Nombre:   "Dr. Carlos Mendoza",
			Doc:      "41258963",
			Email:    "medico@bloodlink.gob.pe",
			Password: "Medico123!",
			Rol:      "MEDICO",
			Hospital: "HOSP-MINSA-01",
		},
		{
			ID:       13,
			Nombre:   "Lic. María Torres",
			Doc:      "43896512",
			Email:    "banco@bloodlink.gob.pe",
			Password: "Banco123!",
			Rol:      "PERSONAL_BANCO",
			Hospital: "HOSP-MINSA-01",
		},
		{
			ID:       14,
			Nombre:   "Mg. Roberto Paredes",
			Doc:      "40125874",
			Email:    "coordinador@bloodlink.gob.pe",
			Password: "Coord123!",
			Rol:      "COORDINADOR",
			Hospital: "HOSP-MINSA-01",
		},
	}

	for _, u := range seedUsers {
		var count int64
		db.Model(&models.UsuarioPersonalORM{}).Where("email = ?", u.Email).Count(&count)
		if count == 0 {
			hash, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
			userORM := models.UsuarioPersonalORM{
				ID:                 u.ID,
				NombreCompleto:     u.Nombre,
				DocumentoIdentidad: u.Doc,
				Email:              u.Email,
				PasswordHash:       string(hash),
				Rol:                u.Rol,
				HospitalID:         u.Hospital,
			}
			if err := db.Create(&userORM).Error; err != nil {
				return fmt.Errorf("error insertando usuario %s: %w", u.Email, err)
			}
		}
	}

	// 3. Unidades de Sangre iniciales (solo si la tabla está vacía)
	var totalUnidades int64
	db.Model(&models.UnidadSangreORM{}).Count(&totalUnidades)
	if totalUnidades == 0 {
		type unitSeed struct {
			ID         string
			Barcode    string
			BancoID    string
			Tipo       string
			Rh         string
			DaysToExp  int
		}

		var units []unitSeed
		// HOSP-MINSA-01: 0 unidades O+ (para que CP-01/CP-02 salte a red externa),
		// 3 unidades A-, 2 unidades B+, y 2 unidades O- que vencen en 4 días (alerta AMARILLA RF-08)
		units = append(units,
			unitSeed{"UND-2026-0101", "BL-MINSA01-20260101", "HOSP-MINSA-01", "A", "-", 25},
			unitSeed{"UND-2026-0102", "BL-MINSA01-20260102", "HOSP-MINSA-01", "A", "-", 28},
			unitSeed{"UND-2026-0103", "BL-MINSA01-20260103", "HOSP-MINSA-01", "A", "-", 30},
			unitSeed{"UND-2026-0104", "BL-MINSA01-20260104", "HOSP-MINSA-01", "B", "+", 20},
			unitSeed{"UND-2026-0105", "BL-MINSA01-20260105", "HOSP-MINSA-01", "B", "+", 22},
			unitSeed{"UND-2026-0106", "BL-MINSA01-20260106", "HOSP-MINSA-01", "O", "-", 4},
			unitSeed{"UND-2026-0107", "BL-MINSA01-20260107", "HOSP-MINSA-01", "O", "-", 4},
		)

		// HOSP-ESSALUD-04 (Legacy SOAP, a ~2 km): 10 unidades O+ disponibles
		for i := 1; i <= 10; i++ {
			units = append(units, unitSeed{
				ID:        fmt.Sprintf("UND-2026-04%02d", i),
				Barcode:   fmt.Sprintf("BL-ESSALUD04-202604%02d", i),
				BancoID:   "HOSP-ESSALUD-04",
				Tipo:      "O",
				Rh:        "+",
				DaysToExp: 30,
			})
		}

		// HOSP-MINSA-02 (Sullana): 6 unidades O+ disponibles
		for i := 1; i <= 6; i++ {
			units = append(units, unitSeed{
				ID:        fmt.Sprintf("UND-2026-02%02d", i),
				Barcode:   fmt.Sprintf("BL-MINSA02-202602%02d", i),
				BancoID:   "HOSP-MINSA-02",
				Tipo:      "O",
				Rh:        "+",
				DaysToExp: 35,
			})
		}

		// CLIN-PRIV-01 (Castilla): 4 unidades A+ y 3 unidades O-
		for i := 1; i <= 4; i++ {
			units = append(units, unitSeed{
				ID:        fmt.Sprintf("UND-2026-030%d", i),
				Barcode:   fmt.Sprintf("BL-CLIN01-2026030%d", i),
				BancoID:   "CLIN-PRIV-01",
				Tipo:      "A",
				Rh:        "+",
				DaysToExp: 28,
			})
		}
		for i := 5; i <= 7; i++ {
			units = append(units, unitSeed{
				ID:        fmt.Sprintf("UND-2026-030%d", i),
				Barcode:   fmt.Sprintf("BL-CLIN01-2026030%d", i),
				BancoID:   "CLIN-PRIV-01",
				Tipo:      "O",
				Rh:        "-",
				DaysToExp: 18,
			})
		}

		for _, u := range units {
			expDate := now.AddDate(0, 0, u.DaysToExp)
			extDate := expDate.AddDate(0, 0, -42)
			unitORM := models.UnidadSangreORM{
				ID:              u.ID,
				CodigoBarras:    u.Barcode,
				BancoID:         u.BancoID,
				TipoSangre:      u.Tipo,
				FactorRh:        u.Rh,
				VolumenMl:       450,
				Origen:          "DONACION_VOLUNTARIA",
				FechaExtraccion: extDate,
				FechaCaducidad:  expDate,
				Estado:          "Disponible",
			}
			if err := db.Create(&unitORM).Error; err != nil {
				return fmt.Errorf("error insertando unidad %s: %w", u.ID, err)
			}
			db.Create(&models.AuditLogORM{
				Entidad:        "UnidadSangre",
				EntidadID:      u.ID,
				Accion:         "REGISTRO_INICIAL",
				EstadoAnterior: "NINGUNO",
				EstadoNuevo:    "Disponible",
				UsuarioID:      "SYSTEM_SEED",
				Detalle:        fmt.Sprintf("Ingreso inicial en banco %s (%s%s)", u.BancoID, u.Tipo, u.Rh),
				Timestamp:      time.Now(),
			})
		}
	}

	// 4. Donantes pre-registrados cifrados con AES-256-GCM
	var totalDonantes int64
	db.Model(&models.DonanteORM{}).Count(&totalDonantes)
	if totalDonantes == 0 {
		seedDonors := []struct {
			ID     string
			DNI    string
			Nombre string
			Cel    string
			Correo string
			Tipo   string
			Rh     string
			Banco  string
		}{
			{
				ID:     "DON-2026-4401",
				DNI:    "71234567",
				Nombre: "MARIA FERNANDA QUISPE ROJAS",
				Cel:    "+51987654321",
				Correo: "mquispe@ciudadano.pe",
				Tipo:   "O",
				Rh:     "+",
				Banco:  "HOSP-MINSA-01",
			},
			{
				ID:     "DON-2026-4402",
				DNI:    "72345678",
				Nombre: "LUIS ALBERTO VARGAS SEMINARIO",
				Cel:    "+51912345678",
				Correo: "lvargas@ciudadano.pe",
				Tipo:   "AB",
				Rh:     "-",
				Banco:  "HOSP-MINSA-01",
			},
		}

		for _, d := range seedDonors {
			dniEnc, err := encryptAESGCMHelper(aesKeyHex, d.DNI)
			if err != nil {
				return err
			}
			celEnc, err := encryptAESGCMHelper(aesKeyHex, d.Cel)
			if err != nil {
				return err
			}
			hashBytes := sha256.Sum256([]byte(d.DNI))
			donORM := models.DonanteORM{
				ID:                      d.ID,
				DNICifrado:              dniEnc,
				DNIHash:                 hex.EncodeToString(hashBytes[:]),
				NombreCompleto:          d.Nombre,
				CelularCifrado:          celEnc,
				Correo:                  d.Correo,
				TipoSangre:              d.Tipo,
				FactorRh:                d.Rh,
				BancoPreferidoID:        d.Banco,
				ValidadoReniec:          true,
				HabilitadoConvocatorias: true,
				Activo:                  true,
				FechaRegistro:           time.Now(),
			}
			if err := db.Create(&donORM).Error; err != nil {
				return fmt.Errorf("error insertando donante %s: %w", d.ID, err)
			}
		}
	}

	slog.Info("base de datos PostgreSQL migrada y seed verificado correctamente")
	return nil
}

func encryptAESGCMHelper(keyHex, plaintext string) (string, error) {
	key, err := hex.DecodeString(keyHex)
	if err != nil || len(key) != 32 {
		key = []byte("0123456789abcdef0123456789abcdef")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
