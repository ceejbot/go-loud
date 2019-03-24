package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"github.com/nlopes/slack"
)

// Slacking off with global vars
var db *redis.Client
var api *slack.Client
var rtm *slack.RTM
var channelsByName map[string]string
var rediskey string
var emojiPattern *regexp.Regexp
var slackUserPattern *regexp.Regexp
var puncPattern *regexp.Regexp
var fuckityPattern *regexp.Regexp
var malcolmPattern *regexp.Regexp

func makeRedis() (r *redis.Client) {
	address, found := os.LookupEnv("REDIS_ADDRESS")
	if !found {
		address = "localhost:6379"
	}
	log.Printf("using redis @ %s to store our data", address)
	client := redis.NewClient(&redis.Options{Addr: address})
	return client
}

func makeChannelMap() {
	channelsByName = make(map[string]string)
	channels, err := api.GetChannels(true)
	if err != nil {
		return
	}

	for _, v := range channels {
		channelsByName[v.Name] = v.ID
	}

	address, found := os.LookupEnv("WELCOME_CHANNEL")
	if found {
		yell(findChannelByName(address), "WITNESS THE POWER OF THIS FULLY-OPERATIONAL LOUDBOT.")
	}
}

func findChannelByName(name string) string {
	// This feels unidiomatic.
	val, ok := channelsByName[name]
	if ok {
		return val
	}
	return ""
}

func handleMessage(event *slack.MessageEvent) {
	if event.SubType == "bot_message" {
		return
	}

	if strings.EqualFold(event.Text, "LOUDBOT REPORT") {
		report(event.Channel)
		return
	}

	if fuckityPattern.MatchString(event.Text) {
		log.Println("FUCKITY BYE!")
		yell(event.Channel, "https://cldup.com/NtvUeudPtg.gif")
		return
	}

	if malcolmPattern.MatchString(event.Text) {
		log.Println("MALCOLM RUNS!")
		yell(event.Channel, "https://cldup.com/w_exMqXKlT.gif")
		return
	}

	if !isLoud(event.Text) {
		return
	}

	// Your basic shout.
	remember(event.Text)
	rejoinder, err := db.SRandMember(rediskey).Result()
	if err != nil {
		log.Printf("error selecting yell: %s", err)
		return
	}
	yell(event.Channel, rejoinder)
	db.Incr(fmt.Sprintf("%s:count", rediskey)).Result()
}

func report(channel string) {
	counter, err := db.Get(fmt.Sprintf("%s:count", rediskey)).Result()
	if err != nil {
		counter = "AN UNKNOWN NUMBER OF"
	}

	card, _ := db.SCard(rediskey).Result()

	reply := fmt.Sprintf("I HAVE YELLED %s TIMES. ", counter)
	reply += fmt.Sprintf("I HAVE %d THINGS TO YELL AT YOU.", card)

	yell(channel, reply)
}

func stripWhitespace(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func isLoud(msg string) bool {
	// strip tags & emoji
	input := stripWhitespace(msg)
	input = emojiPattern.ReplaceAllLiteralString(input, "")
	input = slackUserPattern.ReplaceAllLiteralString(input, "")
	input = puncPattern.ReplaceAllLiteralString(input, "")

	if len(input) == 0 {
		return false
	}

	return strings.ToUpper(input) == input
}

func remember(msg string) {
	db.SAdd(rediskey, msg).Result()
}

func yell(channel string, msg string) {
	channelID, _, err := api.PostMessage(channel,
		slack.MsgOptionText(msg, false),
		slack.MsgOptionUsername("LOUDBOT"),
		slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{
			IconURL:     "https://cldup.com/XjiGTeey6i.png",
			UnfurlLinks: true,
			UnfurlMedia: true,
		}))

	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	log.Printf("YELLED to %s: `%s`", channelID, msg)
}

func main() {
	err := godotenv.Load(".env", "../../.env")

	slacktoken, ok := os.LookupEnv("SLACK_TOKEN")
	if !ok {
		log.Fatal("You must provide an access token in SLACK_TOKEN")
	}

	var found bool
	rediskey, found = os.LookupEnv("REDIS_KEY")
	if !found {
		rediskey = "LOUDBOT_YELLS"
	}

	db = makeRedis()
	card, err := db.SCard(rediskey).Result()
	if err != nil {
		// We fail NOW if we can't find our DB.
		log.Fatal(err)
	}
	log.Printf("LOUDIE HAS %d THINGS TO YELL", card)

	// Regular expressions we'll use a whole lot.
	emojiPattern = regexp.MustCompile(`:[^\t\n\f\r ]+:`)
	slackUserPattern = regexp.MustCompile(`<@[^\t\n\f\r ]+>`)
	puncPattern = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	fuckityPattern = regexp.MustCompile(`(?i)FUCKITY *BYE`)
	malcolmPattern = regexp.MustCompile(`(?i)MALCOLM +TUCKER`)

	api = slack.New(slacktoken)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			makeChannelMap()
			break

		case *slack.MessageEvent:
			// fmt.Printf("Message: %v\n", ev)
			handleMessage(ev)
			break

		case *slack.PresenceChangeEvent:
			// fmt.Printf("Presence Change: %v\n", ev)
			break

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())
			break

		case *slack.InvalidAuthEvent:
			log.Fatal("Invalid credentials")
			break

		default:
			// Ignore other events..
			// fmt.Printf("Event: %v\n", msg.Data.type)
		}
	}
}
