package main

import (
	"testing"
)

const mockAcpiListen = "./mock_acpi_listen.sh"

func TestAcpiEvent(t *testing.T) {
	acpiChan, _ := AcpiEvent(mockAcpiListen, acpiLidOpenEvent)
	event := <-acpiChan

	if event != acpiLidOpenEvent {
		t.Errorf("got %q, want %q", event, acpiLidOpenEvent)
	}
}
