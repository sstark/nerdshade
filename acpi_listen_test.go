package main

import (
	"testing"
	"io"
	"strings"
)

var mockAcpiEvents = `
video/brightnessdown BRTDN 00000087 00000000
video/brightnessup BRTUP 00000086 00000000
video/brightnessdown BRTDN 00000087 00000000
video/brightnessup BRTUP 00000086 00000000
button/lid LID close
button/lid LID open
video/brightnessdown BRTDN 00000087 00000000
wmi PNP0C14:05 000000d0 00000000
video/brightnessup BRTUP 00000086 00000000
wmi PNP0C14:05 000000d0 00000000
video/brightnessdown BRTDN 00000087 00000000
wmi PNP0C14:05 000000d0 00000000
`

func TestAcpiEventScanner(t *testing.T) {
	reader := io.NopCloser(strings.NewReader(mockAcpiEvents))
	acpiChan := acpiEventScanner(reader, acpiLidOpenEvent)
	event := <-acpiChan

	if event != acpiLidOpenEvent {
		t.Errorf("got %q, want %q", event, acpiLidOpenEvent)
	}
}
