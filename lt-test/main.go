package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
	ProgressFile     *os.File
	SummaryFile      *os.File
}

func NewLoadTestStats(progressFilename, summaryFilename string) *LoadTestStats {
	// Создаем директорию если её нет
	os.MkdirAll(filepath.Dir(progressFilename), 0755)
	os.MkdirAll(filepath.Dir(summaryFilename), 0755)

	// Открываем файл для записи прогресса (добавляем в конец)
	progressFile, err := os.OpenFile(progressFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Warning: failed to open progress file: %v", err)
		progressFile = nil
	}

	// Открываем файл для итоговой статистики
	summaryFile, err := os.OpenFile(summaryFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Printf("Warning: failed to open summary file: %v", err)
		summaryFile = nil
	}

	return &LoadTestStats{
		MinResponseTime: time.Hour,
		MaxResponseTime: 0,
		ResponseTimes:   make([]time.Duration, 0),
		ProgressFile:    progressFile,
		SummaryFile:     summaryFile,
	}
}

func (s *LoadTestStats) CloseFiles() {
	if s.ProgressFile != nil {
		s.ProgressFile.Close()
	}
	if s.SummaryFile != nil {
		s.SummaryFile.Close()
	}
}

func (s *LoadTestStats) WriteProgressLine(line string) {
	if s.ProgressFile != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		fullLine := fmt.Sprintf("[%s] %s\n", timestamp, line)
		s.ProgressFile.WriteString(fullLine)
		s.ProgressFile.Sync() // Сбрасываем буфер на диск
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

// Структура для JSON вывода итоговой статистики
type TestSummary struct {
	TotalRequests    int64   `json:"total_requests"`
	SuccessRequests  int64   `json:"success_requests"`
	FailedRequests   int64   `json:"failed_requests"`
	SuccessRate      float64 `json:"success_rate"`
	AverageDuration  string  `json:"average_duration"`
	MinDuration      string  `json:"min_duration"`
	MaxDuration      string  `json:"max_duration"`
	Percentile50     string  `json:"p50_duration"`
	Percentile95     string  `json:"p95_duration"`
	Percentile99     string  `json:"p99_duration"`
	TestDuration     string  `json:"test_duration"`
	RequestsPerSecond int    `json:"requests_per_second"`
	Timestamp        string  `json:"timestamp"`
	TestConfig       struct {
		TargetURL   string `json:"target_url"`
		HealthURL   string `json:"health_url"`
		Service     string `json:"service"`
		TestNumber  string `json:"test_number"`
	} `json:"test_config"`
}

func (s *LoadTestStats) SaveSummaryToJSON(filename string, testDuration time.Duration, rps int, targetURL, healthURL, service, testNumber string) error {
	total := atomic.LoadInt64(&s.TotalRequests)
	success := atomic.LoadInt64(&s.SuccessRequests)
	failed := atomic.LoadInt64(&s.FailedRequests)

	summary := TestSummary{
		TotalRequests:    total,
		SuccessRequests:  success,
		FailedRequests:   failed,
		RequestsPerSecond: rps,
		TestDuration:     testDuration.String(),
		Timestamp:        time.Now().Format(time.RFC3339),
	}

	// Добавляем конфигурацию теста
	summary.TestConfig.TargetURL = targetURL
	summary.TestConfig.HealthURL = healthURL
	summary.TestConfig.Service = service
	summary.TestConfig.TestNumber = testNumber

	if success > 0 {
		s.Mutex.Lock()
		avgDuration := s.TotalDuration / time.Duration(success)
		s.Mutex.Unlock()

		p50, p95, p99 := s.CalculatePercentiles()

		summary.SuccessRate = float64(success)/float64(total)*100
		summary.AverageDuration = avgDuration.String()
		summary.MinDuration = s.MinResponseTime.String()
		summary.MaxDuration = s.MaxResponseTime.String()
		summary.Percentile50 = p50.String()
		summary.Percentile95 = p95.String()
		summary.Percentile99 = p99.String()
	}

	// Создаем директорию если её нет
	os.MkdirAll(filepath.Dir(filename), 0755)

	// Создаем JSON с красивым форматированием
	jsonData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Сохраняем в файл
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

func (s *LoadTestStats) PrintSummary() {
	total := atomic.LoadInt64(&s.TotalRequests)
	success := atomic.LoadInt64(&s.SuccessRequests)
	failed := atomic.LoadInt64(&s.FailedRequests)

	summaryText := "\n=== Load Test Summary ===\n"
	summaryText += fmt.Sprintf("Total requests: %d\n", total)
	summaryText += fmt.Sprintf("Successful: %d\n", success)
	summaryText += fmt.Sprintf("Failed: %d\n", failed)

	if success > 0 {
		s.Mutex.Lock()
		avgDuration := s.TotalDuration / time.Duration(success)
		s.Mutex.Unlock()

		p50, p95, p99 := s.CalculatePercentiles()

		summaryText += fmt.Sprintf("Average response time: %v\n", avgDuration)
		summaryText += fmt.Sprintf("Min response time: %v\n", s.MinResponseTime)
		summaryText += fmt.Sprintf("Max response time: %v\n", s.MaxResponseTime)
		summaryText += fmt.Sprintf("50th percentile: %v\n", p50)
		summaryText += fmt.Sprintf("95th percentile: %v\n", p95)
		summaryText += fmt.Sprintf("99th percentile: %v\n", p99)
		summaryText += fmt.Sprintf("Success rate: %.2f%%\n", float64(success)/float64(total)*100)
	}

	// Выводим в консоль
	fmt.Print(summaryText)
}

func (s *LoadTestStats) PrintProgress() {
	total := atomic.LoadInt64(&s.TotalRequests)
	success := atomic.LoadInt64(&s.SuccessRequests)
	failed := atomic.LoadInt64(&s.FailedRequests)

	var progressLine string

	if success > 0 {
		p50, p95, p99 := s.CalculatePercentiles()

		progressLine = fmt.Sprintf("Req: %d | OK: %d | ERR: %d | P50: %v | P95: %v | P99: %v",
			total, success, failed, p50, p95, p99)
	} else {
		progressLine = fmt.Sprintf("Req: %d | OK: %d | ERR: %d",
			total, success, failed)
	}

	// Выводим в консоль
	fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05.000"), progressLine)

	// Записываем в файл прогресса
	s.WriteProgressLine(progressLine)
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

func waitForService(client *http.Client, healthURL string, timeout time.Duration, stats *LoadTestStats) error {
	startMessage := fmt.Sprintf("Waiting for service to be available at %s...", healthURL)
	fmt.Println(startMessage)
	stats.WriteProgressLine(startMessage)

	start := time.Now()
	client = &http.Client{Timeout: 1 * time.Second}

	for {
		elapsed := int(time.Since(start).Seconds())

		// Проверяем таймаут
		if time.Since(start) > timeout {
			timeoutMessage := fmt.Sprintf("service health check timeout after %v", timeout)
			stats.WriteProgressLine("❌ " + timeoutMessage)
			return fmt.Errorf(timeoutMessage)
		}

		// Делаем запрос как в health checker
		resp, err := client.Get(healthURL)
		if err != nil {
			errorMessage := fmt.Sprintf("Time: %ds - FAILED: %s", elapsed, err)
			fmt.Println(errorMessage)
			stats.WriteProgressLine(errorMessage)
		} else {
			if resp.StatusCode == http.StatusOK {
				successMessage := fmt.Sprintf("Time: %ds - OK (HTTP %d)", elapsed, resp.StatusCode)
				fmt.Println(successMessage)
				stats.WriteProgressLine(successMessage)
				resp.Body.Close()

				finalMessage := "✅ Service is now available and healthy!"
				fmt.Println(finalMessage)
				stats.WriteProgressLine(finalMessage)
				return nil
			} else {
				errorMessage := fmt.Sprintf("Time: %ds - FAILED (HTTP %d)", elapsed, resp.StatusCode)
				fmt.Println(errorMessage)
				stats.WriteProgressLine(errorMessage)
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

func runLoadTest(url string, healthURL string, requestsPerSecond int, duration time.Duration, progressFile, summaryFile string) *LoadTestStats {
	stats := NewLoadTestStats(progressFile, summaryFile)
	defer stats.CloseFiles()

	// Записываем начало теста в файл прогресса
	stats.WriteProgressLine("=== Load Test Started ===")
	stats.WriteProgressLine(fmt.Sprintf("Target: %s", url))
	stats.WriteProgressLine(fmt.Sprintf("Rate: %d requests per second", requestsPerSecond))
	stats.WriteProgressLine(fmt.Sprintf("Duration: %v", duration))

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
	err := waitForService(client, healthURL, 5*time.Minute, stats)
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

	// Записываем завершение теста
	stats.WriteProgressLine("=== Load Test Completed ===")

	return stats
}

// Вспомогательные функции для работы с переменными окружения
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// Функция для создания пути к результатам на основе TEST_NUMBER_ENV и SERVICE
func getResultsPath() string {
	testNumber := getEnv("TEST_NUMBER_ENV", "1")
	service := getEnv("SERVICE", "unknown-service")
	basePath := getEnv("RESULTS_BASE_PATH", "/results")

	// Создаем путь в формате: /results/test_{testNumber}/{service}/
	return filepath.Join(basePath, fmt.Sprintf("test_%s", testNumber), service)
}

func main() {
	// Получаем конфигурацию из переменных окружения
	url := getEnv("TARGET_URL", "http://envoy:8090/api/crypto/encrypt")
	healthURL := getEnv("HEALTH_URL", "http://envoy:8090/api/crypto/health")
	requestsPerSecond := getEnvInt("REQUESTS_PER_SECOND", 20)
	duration := getEnvDuration("TEST_DURATION", 10*time.Minute)

	// Получаем параметры для пути
	testNumber := getEnv("TEST_NUMBER_ENV", "1")
	service := getEnv("SERVICE", "unknown-service")
	resultsPath := getResultsPath()

	// Файлы для вывода
	progressFile := filepath.Join(resultsPath, "load_test_progress.log")
	summaryFile := filepath.Join(resultsPath, "load_test_summary.json")
	jsonFile := filepath.Join(resultsPath, "load_test_results.json")

	fmt.Println("Encrypt Load Tester started")
	fmt.Printf("Configuration:\n")
	fmt.Printf("  URL: %s\n", url)
	fmt.Printf("  Health URL: %s\n", healthURL)
	fmt.Printf("  RPS: %d\n", requestsPerSecond)
	fmt.Printf("  Duration: %v\n", duration)
	fmt.Printf("  Test Number: %s\n", testNumber)
	fmt.Printf("  Service: %s\n", service)
	fmt.Printf("  Results path: %s\n", resultsPath)
	fmt.Printf("  Progress file: %s\n", progressFile)
	fmt.Printf("  Summary file: %s\n", summaryFile)
	fmt.Printf("  JSON file: %s\n", jsonFile)

	// Запускаем нагрузочный тест
	stats := runLoadTest(url, healthURL, requestsPerSecond, duration, progressFile, summaryFile)

	// Выводим финальные результаты в консоль
	stats.PrintSummary()

	// Сохраняем итоговую статистику в JSON файл
	err := stats.SaveSummaryToJSON(summaryFile, duration, requestsPerSecond, url, healthURL, service, testNumber)
	if err != nil {
		log.Printf("❌ Failed to save summary to JSON: %v", err)
	} else {
		fmt.Printf("✅ Summary saved to: %s\n", summaryFile)
	}

	// Сохраняем результаты в дополнительный JSON файл (дублирование для совместимости)
	err = stats.SaveSummaryToJSON(jsonFile, duration, requestsPerSecond, url, healthURL, service, testNumber)
	if err != nil {
		log.Printf("❌ Failed to save results to JSON: %v", err)
	} else {
		fmt.Printf("✅ JSON results saved to: %s\n", jsonFile)
	}

	fmt.Printf("✅ Progress log saved to: %s\n", progressFile)
}