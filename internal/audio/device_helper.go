package audio

import (
	"errors"

	"github.com/gordonklaus/portaudio"
)

func getInputDeviceByName(name string) (*portaudio.DeviceInfo, error) {
	devices, err := portaudio.Devices()
	if err != nil {
		return nil, err
	}

	for _, d := range devices {
		if d.Name == name && d.MaxInputChannels > 0 {
			return d, nil
		}
	}

	return nil, errors.New("microphone not found: " + name)
}
