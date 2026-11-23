package security

import "testing"

func TestAESGCMCryptoService_EncryptDecryptHash(t *testing.T) {
	keyHex := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	svc, err := NewAESGCMCryptoService(keyHex)
	if err != nil {
		t.Fatalf("error inicializando AESGCMCryptoService: %v", err)
	}

	// Fallback de llave no hex de 64 chars
	svcFallback, err := NewAESGCMCryptoService("short_key")
	if err != nil || svcFallback == nil {
		t.Fatalf("fallback key falló: %v", err)
	}

	dni := "76543210"
	enc1, err := svc.Encrypt(dni)
	if err != nil || enc1 == "" || enc1 == dni {
		t.Fatalf("Encrypt falló: %v, enc=%s", err, enc1)
	}

	enc2, _ := svc.Encrypt(dni)
	if enc1 == enc2 {
		t.Errorf("AES-256-GCM debe usar nonce aleatorio por operación")
	}

	dec, err := svc.Decrypt(enc1)
	if err != nil || dec != dni {
		t.Fatalf("Decrypt obtuvo %s (err=%v), esperaba %s", dec, err, dni)
	}

	// Casos de error en Decrypt
	if _, err := svc.Decrypt("not-base64!!!"); err == nil {
		t.Error("esperaba error para base64 inválido")
	}
	if _, err := svc.Decrypt("YWJj"); err == nil {
		t.Error("esperaba error para ciphertext demasiado corto")
	}

	hash1 := svc.HashSHA256(dni)
	hash2 := svc.HashSHA256(dni)
	if len(hash1) != 64 || hash1 != hash2 {
		t.Errorf("HashSHA256 inconsistente: %s", hash1)
	}
}
