package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type loginResp struct {
	AccessToken string `json:"accessToken"`
}

func main() {
	baseURL := flag.String("url", "http://localhost:8080", "URL base del backend BloodLink")
	concurrency := flag.Int("c", 100, "Número de usuarios concurrentes (goroutines)")
	requestsPerWorker := flag.Int("n", 3, "Peticiones por cada usuario concurrente")
	mode := flag.String("mode", "mixed", "Modo de prueba: 'matching' (POST /api/v1/emergencias), 'stock' (GET /api/v1/inventario/stock) o 'mixed'")
	email := flag.String("email", "medico@bloodlink.gob.pe", "Correo del médico para autenticación JWT")
	password := flag.String("pass", "Medico123!", "Contraseña del médico")
	flag.Parse()

	fmt.Println("================================================================================")
	fmt.Println("  BLOODLINK PLATFORM - PRUEBA DE CARGA Y CONCURRENCIA CP-04 (RNF-01)")
	fmt.Println("================================================================================")
	fmt.Printf("Target URL           : %s\n", *baseURL)
	fmt.Printf("Usuarios Concurrentes: %d goroutines\n", *concurrency)
	fmt.Printf("Requests por Usuario : %d (Total: %d peticiones)\n", *requestsPerWorker, (*concurrency)*(*requestsPerWorker))
	fmt.Printf("Modo de Operación    : %s\n", *mode)
	fmt.Println("--------------------------------------------------------------------------------")

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 200,
		},
	}

	token, err := authenticate(client, *baseURL, *email, *password)
	if err != nil {
		fmt.Printf("[ERROR] No se pudo autenticar contra %s: %v\n", *baseURL, err)
		fmt.Println("Asegúrese de que el backend esté en ejecución (docker compose up -d).")
		os.Exit(1)
	}
	fmt.Println("[OK] Token JWT obtenido exitosamente. Iniciando ráfaga concurrente...")

	totalRequests := (*concurrency) * (*requestsPerWorker)
	latencies := make([]float64, 0, totalRequests)
	var latMu sync.Mutex
	var successCount int64
	var conflict409Count int64
	var errorCount int64

	startGlobal := time.Now()
	var wg sync.WaitGroup

	for w := 0; w < *concurrency; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for r := 0; r < *requestsPerWorker; r++ {
				reqStart := time.Now()
				var status int
				var reqErr error

				if *mode == "matching" || (*mode == "mixed" && r == 0) {
					status, reqErr = fireEmergencyMatching(client, *baseURL, token)
				} else {
					status, reqErr = fireStockQuery(client, *baseURL, token)
				}

				elapsedMs := float64(time.Since(reqStart).Microseconds()) / 1000.0
				latMu.Lock()
				latencies = append(latencies, elapsedMs)
				latMu.Unlock()

				if reqErr != nil || status >= 500 || status == 0 {
					atomic.AddInt64(&errorCount, 1)
				} else if status == http.StatusConflict {
					// 409 Conflict es una respuesta válida de negocio cuando el stock se agota sin errores 5xx ni doble asignación
					atomic.AddInt64(&conflict409Count, 1)
					atomic.AddInt64(&successCount, 1)
				} else if status >= 200 && status < 300 {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&errorCount, 1)
				}
			}
		}(w)
	}

	wg.Wait()
	totalDuration := time.Since(startGlobal)

	sort.Float64s(latencies)
	minLat, avgLat, p95Lat, p99Lat, maxLat := computeStats(latencies)
	errorRate := (float64(errorCount) / float64(totalRequests)) * 100.0
	rps := float64(totalRequests) / totalDuration.Seconds()

	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println("  RESULTADOS DE EVIDENCIA CP-04 (SLA RNF-01: Latencia <= 3000 ms, Error = 0%)")
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("Tiempo Total de Prueba : %.2f s (Throughput: %.1f req/s)\n", totalDuration.Seconds(), rps)
	fmt.Printf("Peticiones Exitosas    : %d / %d (Reservas/Consultas 2xx: %d | Sin Stock 409 Controlado: %d)\n",
		successCount, totalRequests, successCount-conflict409Count, conflict409Count)
	fmt.Printf("Errores de Servidor    : %d (Tasa de Error: %.2f%%)\n", errorCount, errorRate)
	fmt.Printf("Latencia Mínima        : %.2f ms\n", minLat)
	fmt.Printf("Latencia Promedio      : %.2f ms\n", avgLat)
	fmt.Printf("Latencia Percentil p95 : %.2f ms\n", p95Lat)
	fmt.Printf("Latencia Percentil p99 : %.2f ms\n", p99Lat)
	fmt.Printf("Latencia Máxima        : %.2f ms\n", maxLat)
	fmt.Println("--------------------------------------------------------------------------------")

	if p95Lat <= 3000.0 && errorCount == 0 {
		fmt.Println("VEREDICTO CP-04: [APROBADO] Cumple RNF-01 (<= 3000 ms bajo 100 usuarios concurrentes y 0% error).")
	} else {
		fmt.Println("VEREDICTO CP-04: [OBSERVADO] Revisar métricas de latencia o tasa de error.")
		os.Exit(1)
	}
}

func authenticate(client *http.Client, baseURL, email, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	resp, err := client.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(raw))
	}

	var lr loginResp
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return "", err
	}
	return lr.AccessToken, nil
}

func fireEmergencyMatching(client *http.Client, baseURL, token string) (int, error) {
	payload, _ := json.Marshal(map[string]any{
		"hospitalOrigenId":  "HOSP-MINSA-01",
		"coordenadasOrigen": "-5.1945,-80.6328",
		"medicoId":          12,
		"tipoSangreReq":     "O",
		"factorRhReq":       "+",
		"cantidadReq":       1,
		"urgencia":          "ALTA_VITAL",
	})
	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/emergencias", bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}

func fireStockQuery(client *http.Client, baseURL, token string) (int, error) {
	req, err := http.NewRequest(http.MethodGet, baseURL+"/api/v1/inventario/stock?tipoSangre=O&factorRh=%2B", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}

func computeStats(sorted []float64) (min, avg, p95, p99, max float64) {
	n := len(sorted)
	if n == 0 {
		return 0, 0, 0, 0, 0
	}
	min = sorted[0]
	max = sorted[n-1]
	sum := 0.0
	for _, v := range sorted {
		sum += v
	}
	avg = sum / float64(n)
	p95Idx := int(math.Min(float64(n-1), math.Ceil(0.95*float64(n))-1))
	p99Idx := int(math.Min(float64(n-1), math.Ceil(0.99*float64(n))-1))
	p95 = sorted[p95Idx]
	p99 = sorted[p99Idx]
	return
}
