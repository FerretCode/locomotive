package webhook

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Coaltergeist/discordgo-embeds/colors"
	"github.com/Coaltergeist/discordgo-embeds/embed"
	"github.com/bwmarrin/discordgo"
	"github.com/ferretcode/locomotive/config"
	"github.com/ferretcode/locomotive/graphql"
	"github.com/ferretcode/locomotive/railway"
)

func SendDiscordWebhook(log graphql.Log, cfg config.Config) error {
	webhookUrl := cfg.DiscordWebhookUrl 

	if webhookUrl == "" {
		return nil
	}

	if !strings.HasPrefix(webhookUrl, "https://discord.com/api/webhooks/") {
		return errors.New("Invalid Discord webhook URL")
	}

	split := strings.Split(webhookUrl, "/")[5:]

	webhookId := split[0]
	webhookToken := split[1]

	s, err := discordgo.New("Webhook " + cfg.DiscordWebhookUrl)

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

func getColor(log graphql.Log) *colors.Color {
	var color *colors.Color

	switch log.Severity {
	case railway.SEVERITY_INFO:
		color = colors.White()
	case railway.SEVERITY_ERROR:
		color = colors.Red()
	case railway.SEVERITY_WARN:
		color = colors.Yellow()
	default:
		color = colors.Black()
	}

	return color
}
