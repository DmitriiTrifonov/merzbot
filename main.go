package main

import (
	"github.com/DmitriiTrifonov/merzbowfier-bot/noise"
	tb "gopkg.in/tucnak/telebot.v2"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	errorMessage = "Cannot process the data!"
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
		tempFile, err := os.Create("in.ogg")
		_, err = io.Copy(tempFile, resp.Body)
		fWrite, err := os.Create("out.wav")
		err = noise.ProcessVoice(tempFile, fWrite)
		if err != nil {
			log.Println(err)
		}
		p := &tb.Document{File: tb.FromDisk("out.wav"), FileName: "tmp.wav"}
		_, _ = b.Send(m.Sender, p)
		err = os.Remove("in.ogg")
		err = os.Remove("out.ogg")
	})

	b.Start()
}
