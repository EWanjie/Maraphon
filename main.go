package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type SensorData struct {
	ID          string
	Temperature float64
}

func pollSensors(sensorURLs []string, timeout time.Duration) {

	resultChan := make(chan SensorData, len(sensorURLs))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	for _, sensorURL := range sensorURLs {
		wg.Add(1)
		go pollSensor(sensorURL, &wg, resultChan, ctx)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	select {
	case <-ctx.Done():
		fmt.Println("Cancelled from the external code")
	case <-time.After(timeout):
		fmt.Println("Timeout reached")
	default:
		for data := range resultChan {
			fmt.Printf("Sensor ID: %s, Temperature: %.2f\n", data.ID, data.Temperature)
		}
	}
}

func pollSensor(sensorURL string, wg *sync.WaitGroup, resultChan chan<- SensorData, ctx context.Context) {
	defer wg.Done()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sensorURL, nil)
	if err != nil {
		fmt.Printf("%s, Error creating request", sensorURL)
		return
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s, Error polling", sensorURL)
		return
	}
	defer resp.Body.Close()

	//Симуляция чтения данных
	sensorId := generate()
	temperature := rand.Float64() * 100

	resultChan <- SensorData{ID: sensorId, Temperature: temperature}
}

func generate() string {
	chars := []rune("abcdefghijklmnopqrstuvwxyz" + "0123456789" + "-")
	length := 36
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String()
	return str
}

func readFile() []string {
	file, err := os.Open("URL.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	lines := []string{}

	for fileScanner.Scan() {
		lines = append(lines, fileScanner.Text())
	}

	return lines
}

func main() {
	sensorURLs := readFile()
	timeout := 10 * time.Second

	pollSensors(sensorURLs, timeout)
}
