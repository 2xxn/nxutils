package web

import (
	"testing"
)

func TestCorsAnywhere(t *testing.T) {
	ca := NewCorsAnywhere("http://149.104.25.177:11000")
	ports := []uint16{80, 8080, 21, 22, 23, 888, 3001, 3306, 6379, 9520, 11000, 27017}

	scanned := ca.TestPorts(ports, 5)
	t.Log("Scanned", len(scanned), "ports")

	passedCheck := false
	for _, port := range scanned {
		if port.IsCorsAnywhere {
			passedCheck = true
			break
		}
		t.Log("Port", port.Port, ", is open:", port.IsOpen, ", is HTTP:", port.IsHttp, ", is CorsAnywhere:", port.IsCorsAnywhere)
	}

	if !passedCheck {
		t.Error("No cors-anywhere found")
	}
}
