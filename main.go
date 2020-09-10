package main

import (
	"github.com/DmitriiTrifonov/merzbowfier-bot/noise"
	tb "gopkg.in/tucnak/telebot.v2"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

const (
	errorMessage = "Cannot process the data!"
	message      = "Please send me a voice message"
)

func main() {
	port := os.Getenv("PORT")
	publicURL := os.Getenv("PUBLIC_URL")
	token := os.Getenv("TOKEN")

	webhook := &tb.Webhook{
		Listen:   ":" + port,
		Endpoint: &tb.WebhookEndpoint{PublicURL: publicURL},
	}

	pref := tb.Settings{
		Token:  token,
		Poller: webhook,
	}

	b, err := tb.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	b.Handle("/start", func(m *tb.Message) {
		_, _ = b.Send(m.Sender, message)
	})

	b.Handle(tb.OnVoice, func(m *tb.Message) {
		voiceId := m.Voice.FileID
		url, err := b.FileURLByID(voiceId)
		if err != nil {
			_, _ = b.Send(m.Sender, errorMessage)
			return
		}
		log.Println(url)
		resp, err := http.Get(url)
		if err != nil {
			_, _ = b.Send(m.Sender, errorMessage)
			return
		}
		tempFile, err := os.Create("in.oga")
		_, err = io.Copy(tempFile, resp.Body)
		cmdOgaToWav := exec.Command("ffmpeg", "-i", "in.oga", "in.wav")
		err = cmdOgaToWav.Run()
		if err != nil {
			log.Println(err)
		}
		tempWav, err := os.Open("in.wav")
		defer tempWav.Close()
		fWrite, err := os.Create("out.wav")
		err = noise.ProcessVoice(tempWav, fWrite)
		if err != nil {
			log.Println(err)
		}
		cmdWavToOga := exec.Command("ffmpeg", "-i", "out.wav", "-acodec", "libopus", "out.oga")
		err = cmdWavToOga.Run()
		p := &tb.Voice{
			File: tb.FromDisk("out.oga"),
		}
		_, err = b.Send(m.Sender, p)
		if err != nil {
			log.Println(err)
		}
		err = os.Remove("in.oga")
		err = os.Remove("out.oga")
		err = os.Remove("in.wav")
		err = os.Remove("out.wav")
	})

	b.Start()
}
