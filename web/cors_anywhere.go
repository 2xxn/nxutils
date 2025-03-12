package web

import (
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
)

type CorsAnywherePort struct {
	Port           uint16
	AsUrl          string
	IsOpen         bool
	IsHttp         bool
	IsCorsAnywhere bool
}

type CorsAnywhere struct {
	url        string
	httpClient http.Client
	openPorts  []*CorsAnywherePort
}

func NewCorsAnywhere(url string) *CorsAnywhere {
	client := http.Client{}

	url = strings.TrimSuffix(url, "/") + "/"

	return &CorsAnywhere{url: url, httpClient: client}
}

func (c *CorsAnywhere) TestPort(port uint16) *CorsAnywherePort {
	caPort := &CorsAnywherePort{
		Port:           port,
		IsOpen:         false,
		IsHttp:         false,
		IsCorsAnywhere: false,
		AsUrl:          c.url + "http://127.0.0.1:" + strconv.Itoa(int(port)),
	}

	request, _ := http.NewRequest("GET", caPort.AsUrl, nil)
	request.Header.Set("Origin", "127.0.0.1")

	response, err := c.httpClient.Do(request)
	if err != nil {
		c.openPorts = append(c.openPorts, caPort)
		return caPort
	}

	defer response.Body.Close()

	contentBytes, _ := io.ReadAll(response.Body)
	content := string(contentBytes)
	isCorsAnywhere := strings.Contains(content, "This API enables cross-origin requests to anywhere.")

	caPort.IsOpen = true
	caPort.IsCorsAnywhere = isCorsAnywhere

	if response.StatusCode == 200 {
		caPort.IsHttp = true
		return caPort
	}

	httpChecks := []bool{
		strings.Contains(content, "Parse Error"),
		strings.Contains(content, "socket hang up"),
		strings.Contains(content, "ECONNRESET"),
	}

	caPort.IsHttp = !slices.Contains(httpChecks, true)

	return caPort
}

func (c *CorsAnywhere) TestPorts(ports []uint16, threads uint8) []*CorsAnywherePort {
	var caPorts []*CorsAnywherePort
	var wg sync.WaitGroup
	var mu sync.Mutex

	threadsChan := make(chan struct{}, threads)

	for _, port := range ports {
		wg.Add(1)
		threadsChan <- struct{}{} // Acquire a slot in the semaphore

		go func(p uint16) {
			defer wg.Done()
			defer func() { <-threadsChan }() // Release the slot in the semaphore

			portResult := c.TestPort(p)

			mu.Lock()
			caPorts = append(caPorts, portResult)
			mu.Unlock()
		}(port)
	}

	wg.Wait() // Wait for all goroutines to finish

	for _, port := range caPorts {
		if port.IsOpen {
			c.openPorts = append(c.openPorts, port)
		}
	}

	return caPorts
}

func (c *CorsAnywhere) GetOpenPorts() []*CorsAnywherePort {
	return c.openPorts
}
