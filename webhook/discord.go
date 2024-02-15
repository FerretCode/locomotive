package webhook

import (
	"fmt"
	"strings"

	"github.com/Coaltergeist/discordgo-embeds/colors"
	"github.com/Coaltergeist/discordgo-embeds/embed"
	"github.com/bwmarrin/discordgo"
	"github.com/ferretcode/locomotive/config"
	"github.com/ferretcode/locomotive/graphql"
	"github.com/ferretcode/locomotive/logline"
	"github.com/ferretcode/locomotive/railway"
)

func SendDiscordWebhook(log *graphql.EnvironmentLog, embedLog bool, cfg *config.Config) error {
	if cfg.DiscordWebhookUrl == "" {
		return nil
	}

	if len(log.MessageRaw) == 0 {
		return nil
	}

	jsonObject, err := logline.ReconstructLogLine(log)

	if err != nil {
		return err
	}

	split := strings.Split(cfg.DiscordWebhookUrl, "/")[5:]

	webhookId := split[0]
	webhookToken := split[1]

	s, err := discordgo.New("Webhook " + cfg.DiscordWebhookUrl)

	if err != nil {
		return err
	}

	webhookParams := &discordgo.WebhookParams{}

	if embedLog {
		em := embed.New().
			SetTitle(strings.ToUpper(log.Severity)).
			SetDescription(fmt.Sprintf("```%s```", log.Message)).
			AddField("⠀", fmt.Sprintf("```%s```", (jsonObject)), false).
			SetColor(getColor(log.Severity))

		webhookParams.Embeds = []*discordgo.MessageEmbed{em.MessageEmbed}
	} else {
		webhookParams.Content = string(log.Message)
	}

	if _, err := s.WebhookExecute(
		webhookId,
		webhookToken,
		true,
		webhookParams,
	); err != nil {
		return err
	}

	return nil
}

func getColor(severity string) *colors.Color {
	var color *colors.Color

	switch severity {
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
