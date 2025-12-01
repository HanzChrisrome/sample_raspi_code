package audio

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gordonklaus/portaudio"
)

type AIFFAudioFormat struct {
	AudioFile       *os.File
	FilePath        string
	Channel         int16
	ChunkSize       int32
	BitsPerSample   int16
	SampleRate      float64
	NumberOfSamples int32
	InputBufferSize uint8
	RecControlSig   *RecondControlSignal
	InputDevice     *portaudio.DeviceInfo
}

func NewAIFFAudioFormat() IAudioFormat {
	return &AIFFAudioFormat{}
}

func (af *AIFFAudioFormat) CreateFilePath(sysPath, filename string) string {
	return filepath.Join(sysPath, fmt.Sprintf("%s.%s", filename, af.GetFileType()))
}

func (af *AIFFAudioFormat) Init(recordControlSig *RecondControlSignal, sysPath, filename string, targetChannel int16, sampleRate float64, inputBufSize uint8) {
	af.RecControlSig = recordControlSig
	af.Channel = targetChannel
	af.SampleRate = sampleRate
	af.BitsPerSample = 32
	af.InputBufferSize = inputBufSize
	af.NumberOfSamples = 0
	af.ChunkSize = 18
	// save target path for creation at Record() time so headers can
	// reflect the actual device sample rate and channel settings
	af.FilePath = af.CreateFilePath(sysPath, filename)
	if sysPath != "" {
		os.MkdirAll(sysPath, os.ModePerm)
	}
}

func (af *AIFFAudioFormat) GetFileType() string { return AIFF_EXT_NAME }

func (af *AIFFAudioFormat) Record() {
	fmt.Println("Starting AIFF recording...")
	defer af.WrapUp()

	// Initialize PortAudio
	if err := portaudio.Initialize(); err != nil {
		panic(err)
	}
	defer portaudio.Terminate()

	// Create input buffer
	in := make([]int32, af.InputBufferSize)

	// Find the target Device
	device := af.InputDevice
	if device == nil {
		device, _ = portaudio.DefaultInputDevice()
	}

	fmt.Println("Using input device:", device.Name)
	// use actual device settings before creating file/header
	af.SampleRate = device.DefaultSampleRate
	af.Channel = int16(device.MaxInputChannels)

	// create file & header now that device sample rate/channel are known
	if af.AudioFile == nil {
		if af.FilePath == "" {
			af.FilePath = af.CreateFilePath("", "recording")
		}
		_file, err := os.Create(af.FilePath)
		must(err)

		_file.WriteString("FORM")
		must(binary.Write(_file, binary.BigEndian, int32(0)))
		_file.WriteString(strings.ToUpper(af.GetFileType()))
		_file.WriteString("COMM")
		must(binary.Write(_file, binary.BigEndian, af.ChunkSize))
		must(binary.Write(_file, binary.BigEndian, af.Channel))
		must(binary.Write(_file, binary.BigEndian, af.NumberOfSamples))
		must(binary.Write(_file, binary.BigEndian, af.BitsPerSample))
		_file.Write(SampleRateToByte(af.SampleRate))
		_file.WriteString("SSND")
		must(binary.Write(_file, binary.BigEndian, int32(0)))
		must(binary.Write(_file, binary.BigEndian, int32(0)))
		must(binary.Write(_file, binary.BigEndian, int32(0)))

		af.AudioFile = _file
	}

	// Stream parameters
	params := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   device,
			Channels: int(af.Channel),
			Latency:  device.DefaultLowInputLatency,
		},
		SampleRate:      af.SampleRate,
		FramesPerBuffer: len(in),
	}

	// Open stream
	stream, err := portaudio.OpenStream(params, in)
	must(err)
	defer stream.Close()

	// Start recording
	must(stream.Start())

	for {
		// Read audio from the device
		must(stream.Read())
		must(binary.Write(af.AudioFile, binary.BigEndian, in))
		af.NumberOfSamples += int32(len(in))

		// Check control signal
		select {
		case ctl := <-af.RecControlSig.Sig:
			if ctl == AUDIO_CTL_STOP_REC {
				must(stream.Stop())
				af.RecControlSig.Sig <- AUDIO_CTL_REC_FULLY_STOPPED
				return
			}
			if ctl == AUDIO_GRACE_KILL_SIG_REQ {
				must(stream.Stop())
				af.WrapUp()
				af.RecControlSig.Sig <- AUDIO_GRACE_KILL_SIG_PROC
				return
			}
		default:
			// continue recording if no signal
		}
	}
}

func (af *AIFFAudioFormat) WrapUp() {
	if af.AudioFile == nil {
		log.Fatal("audio file empty")
	}

	totalBytes := 4 + 8 + 18 + 8 + 8 + 4*af.NumberOfSamples
	_, err := af.AudioFile.Seek(4, 0)
	must(err)
	must(binary.Write(af.AudioFile, binary.BigEndian, totalBytes))

	_, err = af.AudioFile.Seek(22, 0)
	must(err)
	must(binary.Write(af.AudioFile, binary.BigEndian, af.NumberOfSamples))

	_, err = af.AudioFile.Seek(42, 0)
	must(err)
	must(binary.Write(af.AudioFile, binary.BigEndian, int32(4*af.NumberOfSamples+8)))

	must(af.AudioFile.Close())
	fmt.Println("AIFF recording finished")
}
