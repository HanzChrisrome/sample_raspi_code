package audio

import (
	"sync"

	"github.com/gordonklaus/portaudio"
)

type MultiMicRecorder struct {
	recorders []*AIFFAudioFormat
	wg        sync.WaitGroup
}

func NewMultiMicRecorder() *MultiMicRecorder {
	return &MultiMicRecorder{
		recorders: []*AIFFAudioFormat{},
	}
}

func (m *MultiMicRecorder) AddRecorder(rec *AIFFAudioFormat) {
	m.recorders = append(m.recorders, rec)
}

func (m *MultiMicRecorder) AddDevice(device *portaudio.DeviceInfo, sysPath, filename string) {
	rec := NewAIFFAudioFormat().(*AIFFAudioFormat)
	rec.InputDevice = device

	rec.RecControlSig = &RecondControlSignal{
		Sig: make(chan uint8, 1),
	}

	rec.Init(rec.RecControlSig, sysPath, filename, int16(device.MaxInputChannels), device.DefaultSampleRate, 255)
	m.AddRecorder(rec)
}

func (m *MultiMicRecorder) StartAll() {
	for _, rec := range m.recorders {
		m.wg.Add(1)
		go func(r *AIFFAudioFormat) {
			defer m.wg.Done()
			r.Record()
		}(rec)
	}
}

// Stop all recorders
func (m *MultiMicRecorder) StopAll() {
	for _, rec := range m.recorders {
		rec.RecControlSig.Sig <- AUDIO_CTL_STOP_REC
	}
}

// Wait for all recorders to finish
func (m *MultiMicRecorder) WaitUntilFinished() {
	m.wg.Wait()
}
