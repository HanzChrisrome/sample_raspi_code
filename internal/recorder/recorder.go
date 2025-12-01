package recorder

import (
	"errors"

	"github.com/otis-co-ltd/aihub-recorder/internal/audio"
	"github.com/otis-co-ltd/aihub-recorder/internal/config"
)

var currentRecorder audio.IAudioFormat
var recControl *audio.RecondControlSignal
var cfg *config.Config

func init() {
	cfg = config.Load()
}

func Start() error {
	if currentRecorder != nil {
		return errors.New("recording already in progress")
	}

	recControl = audio.NewRecControlSig()
	currentRecorder = audio.NewAudioInstance(cfg.SYS_AUDIO_TYPE)
	currentRecorder.Init(
		recControl,
		cfg.SYS_RECORD_PATH,
		"pi_recording",
		cfg.SYS_AUDIO_CHANNEL,
		cfg.SYS_AUDIO_SAMPLE_RATE,
		cfg.SYS_AUDIO_INPUT_BUFFER_SIZE,
	)

	go currentRecorder.Record()
	return nil
}

func Stop() error {
	if currentRecorder == nil || recControl == nil {
		return errors.New("no recording in progress")
	}

	recControl.Sig <- audio.AUDIO_CTL_STOP_REC

	<-recControl.Sig

	currentRecorder = nil
	recControl = nil
	return nil
}
