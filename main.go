package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"sync"
)

type Id struct {
	id      string
	guildId string
}

type Data struct {
	name      string
	guildName string
	online    float64
}

type discordCollector struct {
	userVoiceMetric *prometheus.Desc
}

func newDiscordCollector() *discordCollector {
	return &discordCollector{
		userVoiceMetric: prometheus.NewDesc(
			prometheus.BuildFQName("discord_exporter", "voice", "connected"),
			"Is a specific user is connected!",
			[]string{"id", "name", "guildId", "guildName"}, nil,
		)}
}

func (collector *discordCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.userVoiceMetric
}

func (collector *discordCollector) Collect(ch chan<- prometheus.Metric) {
	for key, data := range voiceState {
		m := prometheus.MustNewConstMetric(collector.userVoiceMetric, prometheus.GaugeValue, data.online, key.id, data.name, key.guildId, data.guildName)
		ch <- m
	}
}

var (
	token      string
	stateMutex sync.RWMutex
	voiceState = make(map[Id]Data)
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Print(err)
	}
	token = os.Getenv("DISCORD_TOKEN")
}

func main() {
	dCollector := newDiscordCollector()
	err := prometheus.Register(dCollector)
	if err != nil {
		return
	}
	http.Handle("/metrics", promhttp.Handler())
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}
	discord.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates
	discord.AddHandler(voiceStateUpdate)
	err = discord.Open()
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":9101", nil), discord.Close())
}

func voiceStateUpdate(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	id := Id{
		id:      vsu.UserID,
		guildId: vsu.GuildID,
	}

	user, err := s.User(vsu.UserID)
	if err != nil {
		log.Print(err)
	}
	guild, err := s.Guild(vsu.GuildID)
	if err != nil {
		log.Print(err)
	}
	delete(voiceState, id)
	if vsu.ChannelID == "" {
		voiceState[id] = Data{
			name:      user.Username,
			guildName: guild.Name,
			online:    0,
		}
	} else {
		voiceState[id] = Data{
			name:      user.Username,
			guildName: guild.Name,
			online:    1,
		}
	}
}
