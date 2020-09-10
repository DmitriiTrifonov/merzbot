package main

import (
	"fmt"
	"github.com/DmitriiTrifonov/merzbowfier-bot/noise"
	tb "gopkg.in/tucnak/telebot.v2"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	errorMessage = "Cannot process the data!"
	message      = "Please send me a voice message"
)

func randInt() int {
	seed := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(seed)
	return rnd.Int()
}

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

	returnMessage := func(m *tb.Message) {
		if !m.Private() {
			return
		}
		_, _ = b.Send(m.Sender, message)
	}

	b.Handle("/start", returnMessage)

	b.Handle(tb.OnVoice, func(m *tb.Message) {
		id := randInt()

		voiceId := m.Voice.FileID
		url, err := b.FileURLByID(voiceId)
		log.Println("Got the url:", url)

		if err != nil {
			if _, err := b.Send(m.Sender, errorMessage); err != nil {
				log.Println(err)
			}
			return
		}

		resp, err := http.Get(url)
		if err != nil {
			log.Println(err)
			if _, err := b.Send(m.Sender, errorMessage); err != nil {
				log.Println(err)
			}
			return
		}

		inNameOga := fmt.Sprintf("in%d.oga", id)
		inNameWav := fmt.Sprintf("in%d.wav", id)

		inFileOga, err := os.Create(inNameOga)
		if err != nil {
			log.Println(err)
			if _, err := b.Send(m.Sender, errorMessage); err != nil {
				log.Println(err)
			}
			return
		}
		_, err = io.Copy(inFileOga, resp.Body)
		if err != nil {
			log.Println(err)
			if _, err := b.Send(m.Sender, errorMessage); err != nil {
				log.Println(err)
			}
			return
		}

		cmdOgaToWav := exec.Command("ffmpeg", "-i", inNameOga, inNameWav)
		err = cmdOgaToWav.Run()
		if err != nil {
			log.Println(err)
			if _, err := b.Send(m.Sender, errorMessage); err != nil {
				log.Println(err)
			}
			return
		}

		inFileWav, err := os.Open(inNameWav)
		if err != nil {
			log.Println(err)
			if _, err := b.Send(m.Sender, errorMessage); err != nil {
				log.Println(err)
			}
			return
		}

		outNameWav := fmt.Sprintf("out%d.wav", id)
		outFileWav, err := os.Create(outNameWav)
		if err != nil {
			log.Println(err)
			if _, err := b.Send(m.Sender, errorMessage); err != nil {
				log.Println(err)
			}
			return
		}

		err = noise.ProcessVoice(inFileWav, outFileWav)
		if err != nil {
			log.Println(err)
			if _, err := b.Send(m.Sender, errorMessage); err != nil {
				log.Println(err)
			}
			return
		}

		outNameOga := fmt.Sprintf("out%d.oga", id)
		cmdWavToOga := exec.Command("ffmpeg", "-i", outNameWav, "-acodec", "libopus", outNameOga)
		err = cmdWavToOga.Run()
		if err != nil {
			log.Println(err)
			if _, err := b.Send(m.Sender, errorMessage); err != nil {
				log.Println(err)
			}
			return
		}

		p := &tb.Voice{
			File: tb.FromDisk(outNameOga),
		}

		if m.Private() {
			_, err = b.Send(m.Sender, p)
			if err != nil {
				log.Println(err)
				if _, err := b.Send(m.Sender, errorMessage); err != nil {
					log.Println(err)
				}
				return
			}
		} else {
			_, err = b.Reply(m, p)
			if err != nil {
				log.Println(err)
				if _, err := b.Send(m.Sender, errorMessage); err != nil {
					log.Println(err)
				}
				return
			}
		}

		err = os.Remove(inNameOga)
		err = os.Remove(outNameOga)
		err = os.Remove(inNameWav)
		err = os.Remove(outNameWav)
		if err != nil {
			log.Println(err)
		}
	})

	b.Start()
}
