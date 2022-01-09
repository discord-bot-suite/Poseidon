package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/discord-bot-suite/Poseidon/client"
	"github.com/discord-bot-suite/Poseidon/command"
	"github.com/discord-bot-suite/Poseidon/config"
	"github.com/google/safebrowsing"
	"github.com/iAtomPlaza/dgoc"
	"regexp"
)

var (
	conf  *config.Config
	stats *config.Statistics
)

func main() {

	c, err := config.New("./global.json")
	if err != nil {
		panic(err.Error())
	}

	conf = c
	bot, err := client.New(conf)
	if err != nil {
		panic(err.Error())
	}

	defer bot.Start()

	// add event listeners
	bot.Session.AddHandler(messageEvent)

	//register commands...
	commandHandler := dgoc.New(bot.Session)
	dgoc.SetPrefix(bot.Config.Prefix)
	err = commandHandler.AddCommand(&command.Help{})

	if err != nil {
		fmt.Println(err)
		/* no need to return if command could not be loaded */
	}

	// load cached url's into memory
	err = config.LoadCache("./cache.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	// create Statistics watcher
	stats, err = config.NewStatisticsWatcher("./statistics.json")
	if err != nil {
		fmt.Println(err)
		return
	}
}

func messageEvent(s *discordgo.Session, m *discordgo.MessageCreate) {

	stats.ScannedMessages++

	urls := getURLs(m.Content)
	if len(urls) > 0 {

		_ = s.MessageReactionAdd(m.ChannelID, m.ID, "👀")

		status := lookupURLs(urls)
		if status.IsUnsafe {
			_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
			_, _ = s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Color: 0xff0000,
				Title: "Spam detected!",
				Description: fmt.Sprintf(":octagonal_sign: %s, A message you sent was detected to contain an unsafe url!", m.Author.Mention()),
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Powered by google safe browsing api",
				},
			})
		}

		_ = s.MessageReactionsRemoveAll(m.ChannelID, m.ID)
		_ = s.MessageReactionAdd(m.ChannelID, m.ID, conf.EmojiID)
	}

	stats.Update()
}

func getURLs(message string) []string {

	pattern := regexp.MustCompile(`((ht|f)tp(s?)://)[\S]*`)
	x := pattern.FindAll([]byte(message), 3)

	// convert bytes to string slice
	var urls = []string{}
	for _, bytes := range x {
		urls = append(urls, string(bytes))
	}

	return urls
}

func lookupURLs(urls []string) *Status {

	stats.ScannedURLs += len(urls)

	sb, err := safebrowsing.NewSafeBrowser(safebrowsing.Config{
		ID: "poseidon",
		Version: "0.0.1",
		APIKey: conf.APIkey,
	})

	if err != nil {
		fmt.Println(err)
		return nil
	}

	threats, err := sb.LookupURLs(urls)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var result = false
	for _, threat := range threats {
		if len(threat) > 0 {
			result = true
			stats.UnsafeURLs++
		} else {
			stats.SafeURLs++
		}
	}

	return &Status{
		IsUnsafe: result,
	}
}

type Status struct {
	IsUnsafe bool
}