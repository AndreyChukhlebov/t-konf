package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Структуры для запросов и ответов
type SignatureRequest struct {
	Message string `json:"message"`
}

type CryptoResponse struct {
	Encrypted string `json:"encrypted"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type LoadTestStats struct {
	TotalRequests    int64
	SuccessRequests  int64
	FailedRequests   int64
	TotalDuration    time.Duration
	MinResponseTime  time.Duration
	MaxResponseTime  time.Duration
	ResponseTimes    []time.Duration
	Mutex            sync.Mutex
}

func NewLoadTestStats() *LoadTestStats {
	return &LoadTestStats{
		MinResponseTime: time.Hour, // Инициализируем большим значением
		MaxResponseTime: 0,
		ResponseTimes:   make([]time.Duration, 0),
	}
}

func (s *LoadTestStats) AddSuccess(duration time.Duration) {
	atomic.AddInt64(&s.TotalRequests, 1)
	atomic.AddInt64(&s.SuccessRequests, 1)

	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.TotalDuration += duration
	if duration < s.MinResponseTime {
		s.MinResponseTime = duration
	}
	if duration > s.MaxResponseTime {
		s.MaxResponseTime = duration
	}
	s.ResponseTimes = append(s.ResponseTimes, duration)
}

func (s *LoadTestStats) AddFailure() {
	atomic.AddInt64(&s.TotalRequests, 1)
	atomic.AddInt64(&s.FailedRequests, 1)
}

func (s *LoadTestStats) CalculatePercentiles() (p50, p95, p99 time.Duration) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if len(s.ResponseTimes) == 0 {
		return 0, 0, 0
	}

	times := make([]time.Duration, len(s.ResponseTimes))
	copy(times, s.ResponseTimes)
	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })

	p50 = times[int(float64(len(times))*0.50)]
	p95 = times[int(float64(len(times))*0.95)]
	p99 = times[int(float64(len(times))*0.99)]

	return p50, p95, p99
}

func (s *LoadTestStats) PrintSummary() {
	total := atomic.LoadInt64(&s.TotalRequests)
	success := atomic.LoadInt64(&s.SuccessRequests)
	failed := atomic.LoadInt64(&s.FailedRequests)

	fmt.Println("\n=== Load Test Summary ===")
	fmt.Printf("Total requests: %d\n", total)
	fmt.Printf("Successful: %d\n", success)
	fmt.Printf("Failed: %d\n", failed)

	if success > 0 {
		s.Mutex.Lock()
		avgDuration := s.TotalDuration / time.Duration(success)
		s.Mutex.Unlock()

		p50, p95, p99 := s.CalculatePercentiles()

		fmt.Printf("Average response time: %v\n", avgDuration)
		fmt.Printf("Min response time: %v\n", s.MinResponseTime)
		fmt.Printf("Max response time: %v\n", s.MaxResponseTime)
		fmt.Printf("50th percentile: %v\n", p50)
		fmt.Printf("95th percentile: %v\n", p95)
		fmt.Printf("99th percentile: %v\n", p99)
		fmt.Printf("Success rate: %.2f%%\n", float64(success)/float64(total)*100)
	}
}

func (s *LoadTestStats) PrintProgress() {
	total := atomic.LoadInt64(&s.TotalRequests)
	success := atomic.LoadInt64(&s.SuccessRequests)
	failed := atomic.LoadInt64(&s.FailedRequests)

	if success > 0 {
		p50, p95, p99 := s.CalculatePercentiles()

		fmt.Printf("[%s] Req: %d | OK: %d | ERR: %d | P50: %v | P95: %v | P99: %v\n",
			time.Now().Format("15:04:05.000"),
			total, success, failed,
			p50, p95, p99)
	} else {
		fmt.Printf("[%s] Req: %d | OK: %d | ERR: %d\n",
			time.Now().Format("15:04:05.000"),
			total, success, failed)
	}
}

func sendEncryptRequest(client *http.Client, url string, message string) error {
	requestBody := SignatureRequest{
		Message: message,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return fmt.Errorf("HTTP %d: failed to parse error response", resp.StatusCode)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, errorResp.Error)
	}

	var cryptoResp CryptoResponse
	if err := json.NewDecoder(resp.Body).Decode(&cryptoResp); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	// Убрали логирование каждого успешного запроса
	_ = cryptoResp.Encrypted // используем переменную, чтобы избежать warning
	return nil
}

func waitForService(client *http.Client, healthURL string, timeout time.Duration) error {
	fmt.Printf("Waiting for service to be available at %s...\n", healthURL)

	start := time.Now()
	client = &http.Client{Timeout: 1 * time.Second}

	for {
		elapsed := int(time.Since(start).Seconds())

		// Проверяем таймаут
		if time.Since(start) > timeout {
			return fmt.Errorf("service health check timeout after %v", timeout)
		}

		// Делаем запрос как в health checker
		resp, err := client.Get(healthURL)
		if err != nil {
			fmt.Printf("Time: %ds - FAILED: %s\n", elapsed, err)
		} else {
			if resp.StatusCode == http.StatusOK {
				fmt.Printf("Time: %ds - OK (HTTP %d)\n", elapsed, resp.StatusCode)
				resp.Body.Close()
				fmt.Println("✅ Service is now available and healthy!")
				return nil
			} else {
				fmt.Printf("Time: %ds - FAILED (HTTP %d)\n", elapsed, resp.StatusCode)
			}
			resp.Body.Close()
		}

		time.Sleep(1 * time.Second)
	}
}

func runWorker(id int, client *http.Client, url string, stats *LoadTestStats,
	ticker *time.Ticker, done <-chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			message := fmt.Sprintf("Test message %d %d", id, time.Now().UnixNano())

			err := sendEncryptRequest(client, url, message)
			if err != nil {
				// Убрали логирование каждой ошибки
				stats.AddFailure()
			} else {
				duration := time.Since(start)
				stats.AddSuccess(duration)
			}

		case <-done:
			return
		}
	}
}

func runProgressReporter(stats *LoadTestStats, done <-chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	// Прогресс каждые 1 секунды
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats.PrintProgress()
		case <-done:
			return
		}
	}
}

func runLoadTest(url string, healthURL string, requestsPerSecond int, duration time.Duration) *LoadTestStats {
	stats := NewLoadTestStats()

	// Создаем HTTP клиент с таймаутами
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	fmt.Printf("Starting load test:\n")
	fmt.Printf("  Target: %s\n", url)
	fmt.Printf("  Rate: %d requests per second\n", requestsPerSecond)
	fmt.Printf("  Duration: %v\n", duration)
	fmt.Println("Progress updates every 1 seconds:")
	fmt.Println("Format: [Time] Req: total | OK: success | ERR: failed | P50: 50th percentile | P95: 95th percentile | P99: 99th percentile")

	// Ждем пока сервис не станет доступен (таймаут 5 минут)
	err := waitForService(client, healthURL, 5*time.Minute)
	if err != nil {
		log.Fatalf("❌ Service health check failed: %v", err)
	}

	// Создаем тикер для равномерного распределения запросов
	interval := time.Second / time.Duration(requestsPerSecond)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Канал для остановки воркеров и репортера
	done := make(chan bool)
	var wg sync.WaitGroup

	// Запускаем один воркер (Go эффективно обрабатывает конкурентность)
	wg.Add(1)
	go runWorker(1, client, url, stats, ticker, done, &wg)

	// Запускаем репортер прогресса
	wg.Add(1)
	go runProgressReporter(stats, done, &wg)

	// Ждем указанное время
	time.Sleep(duration)

	// Останавливаем воркеров и репортер
	close(done)
	wg.Wait()

	return stats
}

func main() {
	// Конфигурация теста
	config := struct {
		URL                string
		HealthURL          string
		RequestsPerSecond  int
		Duration           time.Duration
	}{
		URL:               "http://envoy:8090/api/crypto/encrypt",
		HealthURL:         "http://envoy:8090/api/crypto/health",
		RequestsPerSecond: 20,
		Duration:          60 * 10 * time.Second, // 3 минуты тестирования
	}

	fmt.Println("Encrypt Load Tester started")

	// Запускаем нагрузочный тест
	stats := runLoadTest(config.URL, config.HealthURL, config.RequestsPerSecond, config.Duration)

	// Выводим финальные результаты
	stats.PrintSummary()
}