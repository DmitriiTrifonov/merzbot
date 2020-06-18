package noise

import (
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
	"io"
)

func ProcessVoice(in io.ReadCloser, out io.WriteSeeker) error {
	streamer, format, err := vorbis.Decode(in)
	defer func() error {
		if streamer != nil {
			err := streamer.Close()
			return err
		}
		return nil
	}()

	gain := effects.Gain{Streamer: streamer, Gain: 1000}
	err = wav.Encode(out, &gain, format)
	if err != nil {
		return err
	}
	return nil
}
