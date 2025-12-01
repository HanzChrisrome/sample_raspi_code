package audio

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gordonklaus/portaudio"
)

type WAVAudioFormat struct {
	AudioFile       *os.File
	FilePath        string
	Channel         int16
	BitsPerSample   int16
	SampleRate      float64
	NumberOfSamples int32
	InputBufferSize uint8
	RecControlSig   *RecondControlSignal
}

func NewWAVAudioFormat() IAudioFormat {
	return &WAVAudioFormat{}
}

func (wf *WAVAudioFormat) CreateFilePath(sysPath, filename string) string {
	return filepath.Join(sysPath, fmt.Sprintf("%s.%s", filename, wf.GetFileType()))
}

func (wf *WAVAudioFormat) Init(recordControlSig *RecondControlSignal, sysPath, filename string, targetChannel int16, sampleRate float64, inputBufSize uint8) {
	wf.RecControlSig = recordControlSig
	wf.Channel = targetChannel
	wf.SampleRate = sampleRate
	wf.BitsPerSample = 32
	wf.InputBufferSize = inputBufSize
	wf.NumberOfSamples = 0
	// save target path for creation at Record() time so headers can
	// reflect actual device settings
	wf.FilePath = wf.CreateFilePath(sysPath, filename)
}

func (wf *WAVAudioFormat) GetFileType() string { return WAV_EXT_NAME }

func (wf *WAVAudioFormat) Record() {
	fmt.Println("Starting WAV recording...")
	defer wf.WrapUp()

	portaudio.Initialize()
	defer portaudio.Terminate()

	in := make([]int32, wf.InputBufferSize)

	// ensure device/sample/channel are set before creating header
	// find default device if not provided elsewhere (keeps behaviour similar)
	// NOTE: WAV Record used OpenDefaultStream previously; retain similar behaviour

	// create file & header if not already created
	if wf.AudioFile == nil {
		if wf.FilePath == "" {
			wf.FilePath = wf.CreateFilePath("", "recording")
		}
		_file, err := os.Create(wf.FilePath)
		must(err)

		// WAV header
		_file.WriteString("RIFF")
		must(binary.Write(_file, binary.LittleEndian, int32(0))) // placeholder for file size
		_file.WriteString("WAVE")
		_file.WriteString("fmt ")
		must(binary.Write(_file, binary.LittleEndian, int32(16)))
		must(binary.Write(_file, binary.LittleEndian, int16(1)))
		must(binary.Write(_file, binary.LittleEndian, wf.Channel))
		must(binary.Write(_file, binary.LittleEndian, int32(wf.SampleRate)))
		byteRate := int32(wf.SampleRate) * int32(wf.Channel) * int32(wf.BitsPerSample) / 8
		blockAlign := int16(wf.Channel) * wf.BitsPerSample / 8
		must(binary.Write(_file, binary.LittleEndian, byteRate))
		must(binary.Write(_file, binary.LittleEndian, blockAlign))
		must(binary.Write(_file, binary.LittleEndian, wf.BitsPerSample))
		_file.WriteString("data")
		must(binary.Write(_file, binary.LittleEndian, int32(0))) // data size placeholder

		wf.AudioFile = _file
	}

	stream, err := portaudio.OpenDefaultStream(1, 0, wf.SampleRate, len(in), in)
	must(err)
	defer stream.Close()
	must(stream.Start())

	for {
		must(stream.Read())
		must(binary.Write(wf.AudioFile, binary.LittleEndian, in))
		wf.NumberOfSamples += int32(len(in))

		select {
		case ctl := <-wf.RecControlSig.Sig:
			if ctl == AUDIO_CTL_STOP_REC {
				must(stream.Stop())
				wf.RecControlSig.Sig <- AUDIO_CTL_REC_FULLY_STOPPED
				return
			}
			if ctl == AUDIO_GRACE_KILL_SIG_REQ {
				must(stream.Stop())
				wf.WrapUp()
				wf.RecControlSig.Sig <- AUDIO_GRACE_KILL_SIG_PROC
			}
		default:
		}
	}
}

func (wf *WAVAudioFormat) WrapUp() {
	if wf.AudioFile == nil {
		log.Fatal("audio file empty")
	}

	dataSize := wf.NumberOfSamples * int32(wf.Channel) * int32(wf.BitsPerSample) / 8
	totalFileSize := 36 + dataSize

	_, err := wf.AudioFile.Seek(4, 0)
	must(err)
	must(binary.Write(wf.AudioFile, binary.LittleEndian, totalFileSize-8))

	_, err = wf.AudioFile.Seek(40, 0)
	must(err)
	must(binary.Write(wf.AudioFile, binary.LittleEndian, dataSize))

	must(wf.AudioFile.Close())
	fmt.Println("WAV recording finished")
}
