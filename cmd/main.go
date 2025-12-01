package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/otis-co-ltd/aihub-recorder/internal/audio"
)

func main() {
	// Initialize PortAudio
	if err := portaudio.Initialize(); err != nil {
		log.Fatal(err)
	}
	defer portaudio.Terminate()

	// List all input devices
	devices, err := portaudio.Devices()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Available input devices:")
	for i, d := range devices {
		if d.MaxInputChannels > 0 {
			fmt.Printf("%d: %s (max channels: %d, default sample rate: %.2f)\n", i, d.Name, d.MaxInputChannels, d.DefaultSampleRate)
		}
	}

	// Pick a device (e.g., first USB mic)
	selectedDevice := devices[0]

	// Create MultiMicRecorder
	multiRec := audio.NewMultiMicRecorder()

	// Add recorder for this device
	multiRec.AddDevice(selectedDevice, "recordings", "test_mic2")

	// Start recording
	go multiRec.StartAll()

	recordDuration := 5 * time.Second // change duration as needed
	fmt.Printf("Recording for %v...\n", recordDuration)
	time.Sleep(recordDuration)

	// Stop recording
	multiRec.StopAll()
	multiRec.WaitUntilFinished()

	fmt.Println("Recording finished. Check recordings folder.")
}
