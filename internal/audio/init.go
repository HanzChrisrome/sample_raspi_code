package audio

type IAudioFormat interface {
	CreateFilePath(sysPath, filename string) string
	Init(recordControlSig *RecondControlSignal, sysPath, filename string, targetChannel int16, sampleRate float64, inputBufSize uint8)
	GetFileType() string
	Record()
	WrapUp()
}

// Graceful Stop/Start of Recording
type RecondControlSignal struct {
	Sig chan uint8
}

func NewRecControlSig() *RecondControlSignal {
	return &RecondControlSignal{
		Sig: make(chan uint8),
	}
}

const (
	AUDIO_CTL_STOP_REC          = 0
	AUDIO_CTL_START_REC         = 1
	AUDIO_CTL_REC_FULLY_STOPPED = 2
	AUDIO_GRACE_KILL_SIG_REQ    = 3
	AUDIO_GRACE_KILL_SIG_PROC   = 4
)

func NewAudioInstance(audioType uint8) IAudioFormat {
	switch audioType {
	case AUDIO_AIFF:
		return NewAIFFAudioFormat()
	case AUDIO_WAV:
		return NewWAVAudioFormat()
	default:
		return nil
	}
}
