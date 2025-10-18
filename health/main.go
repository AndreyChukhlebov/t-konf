package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Health checker started")
	start := time.Now()
	client := &http.Client{Timeout: 1 * time.Second}

	// URL для проверки
	healthURL := "http://envoy:8090/api/crypto/health"

	for {
		elapsed := int(time.Since(start).Seconds())

		// Делаем запрос
		resp, err := client.Get(healthURL)
		if err != nil {
			fmt.Printf("Time: %ds - FAILED: %s\n", elapsed, err)
		} else {
			if resp.StatusCode == http.StatusOK {
				fmt.Printf("Time: %ds - OK (HTTP %d)\n", elapsed, resp.StatusCode)
			} else {
				fmt.Printf("Time: %ds - FAILED (HTTP %d)\n", elapsed, resp.StatusCode)
			}
			resp.Body.Close()
		}

		time.Sleep(1 * time.Second)
	}
}