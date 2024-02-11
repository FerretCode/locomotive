package webhook

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Coaltergeist/discordgo-embeds/colors"
	"github.com/Coaltergeist/discordgo-embeds/embed"
	"github.com/bwmarrin/discordgo"
	"github.com/ferretcode/locomotive/railway"
)

type Log struct {
	Message  string
	Severity string
	Embed    bool
}

func SendDiscordWebhook(log Log) error {
	webhookUrl := os.Getenv("DISCORD_WEBHOOK_URL")

	if webhookUrl == "" {
		return nil
	}

	if !strings.HasPrefix(webhookUrl, "https://discord.com/api/webhooks/") {
		return errors.New("Invalid Discord webhook URL")
	}

	split := strings.Split(webhookUrl, "/")[5:]

	webhookId := split[0]
	webhookToken := split[1]

	s, err := discordgo.New("Webhook " + os.Getenv("DISCORD_WEBHOOK_URL"))

	if err != nil {
		return err
	}

	webhookParams := &discordgo.WebhookParams{}

	if log.Embed {
		em := embed.New().
			SetTitle(strings.ToUpper(log.Severity)).
			SetDescription(fmt.Sprintf("```%s```", log.Message)).
			SetColor(getColor(log))

		webhookParams.Embeds = []*discordgo.MessageEmbed{em.MessageEmbed}
	} else {
		webhookParams.Content = log.Message
	}

	_, err = s.WebhookExecute(
		webhookId,
		webhookToken,
		true,
		webhookParams,
	)

	if err != nil {
		return err
	}

	return nil
}

func getColor(log Log) *colors.Color {
	var color *colors.Color

	switch log.Severity {
	case railway.SEVERITY_INFO:
		color = colors.White()
	case railway.SEVERITY_ERROR:
		color = colors.Red()
	default:
		color = colors.Black()
	}

	return color
}
