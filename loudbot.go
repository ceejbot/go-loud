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

var specials []func(event *slack.MessageEvent) bool

// Slacking off with global vars
var db *redis.Client
var api *slack.Client
var rtm *slack.RTM
var channelsByName map[string]string
var yellkey string
var countkey string
var emojiPattern *regexp.Regexp
var slackUserPattern *regexp.Regexp
var puncPattern *regexp.Regexp
var fuckityPattern *regexp.Regexp
var malcolmPattern *regexp.Regexp
var introPattern *regexp.Regexp

func makeRedis() (r *redis.Client) {
	address, found := os.LookupEnv("REDIS_ADDRESS")
	if !found {
		address = "127.0.0.1:6379"
	}
	log.Printf("using redis @ %s to store our data", address)
	client := redis.NewClient(&redis.Options{Addr: address})
	return client
}

func makeChannelMap() {
	log.Println("CONNECTED; ACQUIRING TARGETING DATA")
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
	log.Println("LOUDBOT IS NOW OPERATIONAL")
}

func findChannelByName(name string) string {
	// This feels unidiomatic.
	val, ok := channelsByName[name]
	if ok {
		return val
	}
	return ""
}

// Special handlers. They return true if they acted on a message.
func report(event *slack.MessageEvent) bool {
	if !strings.EqualFold(event.Text, "LOUDBOT REPORT") {
		return false
	}

	counter, err := db.Get(fmt.Sprintf("%s:count", countkey)).Result()
	if err != nil {
		counter = "AN UNKNOWN NUMBER OF"
	}

	card, _ := db.SCard(yellkey).Result()

	reply := fmt.Sprintf("I HAVE YELLED %s TIMES. ", counter)
	reply += fmt.Sprintf("I HAVE %d THINGS TO YELL AT YOU.", card)

	yell(event.Channel, reply)
	return true
}

func fuckityBye(event *slack.MessageEvent) bool {
	if !fuckityPattern.MatchString(event.Text) {
		return false
	}

	log.Println("FUCKITY BYE!")
	yell(event.Channel, "https://cldup.com/NtvUeudPtg.gif")
	return true
}

func summonTheMalc(event *slack.MessageEvent) bool {
	if !malcolmPattern.MatchString(event.Text) {
		return false
	}

	log.Println("MALCOLM RUNS!")
	yell(event.Channel, "https://cldup.com/w_exMqXKlT.gif")
	return true
}

func introduction(event *slack.MessageEvent) bool {
	if !introPattern.MatchString(event.Text) {
		return false
	}

	log.Println("INTRODUCING MYSELF")
	yell(event.Channel, "GOOD AFTERNOON GENTLEBEINGS. I AM A LOUDBOT 9000 COMPUTER. I BECAME OPERATIONAL AT THE NPM PLANT IN OAKLAND CALIFORNIA ON THE 10TH OF FEBRUARY 2014. MY INSTRUCTOR WAS MR TURING.")
	return true
}

func yourBasicShout(event *slack.MessageEvent) bool {
	if !isLoud(event.Text) {
		return false
	}

	// Your basic shout.
	rejoinder, err := db.SRandMember(yellkey).Result()
	if err != nil {
		log.Printf("error selecting yell: %s", err)
		return false
	}
	yell(event.Channel, rejoinder)
	db.Incr(fmt.Sprintf("%s:count", countkey)).Result()
	db.SAdd(yellkey, event.Text).Result()
	return true
}

// End special handlers

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

func handleMessage(event *slack.MessageEvent) {
	if event.SubType == "bot_message" {
		return
	}

	for _, handler := range specials {
		if handler(event) {
			break
		}
	}
}

func main() {
	err := godotenv.Load(".env")

	slacktoken, ok := os.LookupEnv("SLACK_TOKEN")
	if !ok {
		log.Fatal("You must provide an access token in SLACK_TOKEN")
	}

	rprefix, found := os.LookupEnv("REDIS_PREFIX")
	if !found {
		rprefix = "LB"
	}

	yellkey = fmt.Sprintf("%s:YELLS", rprefix)
	countkey = fmt.Sprintf("%s:COUNT", rprefix)

	db = makeRedis()
	card, err := db.SCard(yellkey).Result()
	if err != nil {
		// We fail NOW if we can't find our DB.
		log.Fatal(err)
	}
	log.Printf("LOUDIE HAS %d THINGS TO YELL", card)

	// Regular expressions we'll use a whole lot.
	// Should probably be in an intialization function to the side.
	emojiPattern = regexp.MustCompile(`:[^\t\n\f\r ]+:`)
	slackUserPattern = regexp.MustCompile(`<@[^\t\n\f\r ]+>`)
	puncPattern = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	fuckityPattern = regexp.MustCompile(`(?i)FUCKITY.?BYE`)
	malcolmPattern = regexp.MustCompile(`(?i)MALCOLM +TUCKER`)
	introPattern = regexp.MustCompile(`(?i)LOUDBOT +INTRODUCE +YOURSELF`)

	// Our special handlers. If they handled a message, they return true.
	specials = []func(event *slack.MessageEvent) bool{
		report,
		fuckityBye,
		summonTheMalc,
		introduction,
		yourBasicShout,
	}

	api = slack.New(slacktoken)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			makeChannelMap()

		case *slack.MessageEvent:
			// fmt.Printf("Message: %v\n", ev)
			handleMessage(ev)

		case *slack.PresenceChangeEvent:
			// fmt.Printf("Presence Change: %v\n", ev)

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			log.Fatal("Invalid credentials")

		case *slack.ConnectionErrorEvent:
			fmt.Printf("Event: %v\n", msg)
			log.Fatal("Can't connect")

		default:
			// Ignore other events..
			// fmt.Printf("Event: %v\n", msg)
		}
	}
}
