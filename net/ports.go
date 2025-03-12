package net

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Port struct {
	Address string
	Port    uint16
	IsOpen  bool
	IsTCP   bool
	IsUDP   bool
}

func isTCPPortOpen(address string, port int, timeout time.Duration) bool {
	target := fmt.Sprintf("%s:%d", address, port)
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func isUDPPortOpen(address string, port int, timeout time.Duration) bool {
	target := fmt.Sprintf("%s:%d", address, port)
	conn, err := net.DialTimeout("udp", target, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()

	_, err = conn.Write([]byte("Ping"))
	if err != nil {
		return false
	}

	conn.SetReadDeadline(time.Now().Add(timeout))
	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	if err != nil {
		return true
	}
	return true
}

func IsPortOpen(address string, port uint16) *Port {
	p := &Port{
		Address: address,
		Port:    port,
	}

	if isTCPPortOpen(address, int(port), 1*time.Second) {
		p.IsOpen = true
		p.IsTCP = true
		return p
	}

	if isUDPPortOpen(address, int(port), 1*time.Second) {
		p.IsOpen = true
		p.IsUDP = true
		return p
	}

	return p
}

func ScanPorts(address string, ports []uint16, threads uint8) []*Port {
	var openPorts []*Port
	var threadsChan = make(chan struct{}, threads)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, port := range ports {
		wg.Add(1)
		threadsChan <- struct{}{}

		go func(p uint16) {
			defer wg.Done()
			defer func() { <-threadsChan }() // Release the slot in the semaphore

			tPort := IsPortOpen(address, p)
			if tPort.IsOpen {
				mu.Lock()
				openPorts = append(openPorts, tPort)
				mu.Unlock()
			}
		}(port)
	}

	wg.Wait() // Wait for all goroutines to finish
	return openPorts
}
