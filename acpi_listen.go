package main

import (
	"bufio"
	"io"
	"os/exec"
)

// AcpiEvent returns a channel from which you can receive ACPI events
// by reading from acpi_listen output.
// Arguments are like printed by acpi_listen, for example:
//
//	"button/lid LID close"
//	"button/lid LID open"
//	"video/brightnessdown BRTDN 00000087 00000000"
//	"wmi PNP0C14:05 000000d0 00000000"
//
// Passing one or more of these strings will make the returned channel
// only react to the corresponding events. Passing "" will make it react
// to all events.
// The item returned from the channel will be the ACPI event string in
// the same format as the filter argument.
// AcpiEvent will create a goroutine that spawns an acpi_listen instance
// and wait for output from it to fill the event channel.
func AcpiEvent(cmdName string, watchedEvent string) (eventsOut chan string, err error) {
	cmd := exec.Command(cmdName)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	return acpiEventScanner(out, watchedEvent), err
}

func acpiEventScanner(cmdIo io.ReadCloser, watchedEvent string) (eventsOut chan string) {
	eventsOut = make(chan string, 100)
	scanner := bufio.NewScanner(cmdIo)
	go func() {
		for scanner.Scan() {
			text := scanner.Text()
			if text == watchedEvent || watchedEvent == "" {
				line := scanner.Text()
				eventsOut <- line
			}
		}
	}()
	return eventsOut
}
