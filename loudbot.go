package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"unicode"

	ships "github.com/ceejbot/vfp-culture-ships"
	"github.com/go-redis/redis"
	strip "github.com/grokify/html-strip-tags-go"
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
var catkey string
var swkey string

// Regular expressions we'll use a whole lot.
var patterns = map[string]*regexp.Regexp{
	"emoji":       regexp.MustCompile(`:[^\t\n\f\r ]+:`),
	"slack":       regexp.MustCompile(`<@[^\t\n\f\r ]+>`),
	"punctuation": regexp.MustCompile(`[^a-zA-Z]+`),
	"fuckity":     regexp.MustCompile(`(?i)FUCKITY.?BYE`),
	"malcolm":     regexp.MustCompile(`(?i)MALCOLM +TUCKER`),
	"intro":       regexp.MustCompile(`(?i)LOUDBOT +INTRODUCE +YOURSELF`),
	"ship":        regexp.MustCompile(`(?i)SHIP ?NAME`),
	"starwar":     regexp.MustCompile(`(?i)(LUKE|LEIA|LIGHTSABER|DARTH|OBIWAN|KENOBI|CHEWIE|CHEWBACCA|TATOOINE|STAR +WAR|DEATH +STAR)`),
}

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
		yellWithoutPrompt(findChannelByName(address), "YOU MAY FIRE WHEN READY.")
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

	yell(event, reply)
	return true
}

func fuckityBye(event *slack.MessageEvent) bool {
	if !patterns["fuckity"].MatchString(event.Text) {
		return false
	}

	log.Println("FUCKITY BYE!")
	yell(event, "https://cldup.com/NtvUeudPtg.gif")
	return true
}

func summonTheMalc(event *slack.MessageEvent) bool {
	if !patterns["malcolm"].MatchString(event.Text) {
		return false
	}

	log.Println("MALCOLM RUNS!")
	yell(event, "https://cldup.com/w_exMqXKlT.gif")
	return true
}

func ship(event *slack.MessageEvent) bool {
	if !patterns["ship"].MatchString(event.Text) {
		return false
	}

	yell(event, ships.Random())
	return true
}

func introduction(event *slack.MessageEvent) bool {
	if !patterns["intro"].MatchString(event.Text) {
		return false
	}

	log.Println("INTRODUCING MYSELF")
	yell(event, "GOOD AFTERNOON GENTLEBEINGS. I AM A LOUDBOT 9000 COMPUTER. I BECAME OPERATIONAL AT THE NPM PLANT IN OAKLAND CALIFORNIA ON THE 10TH OF FEBRUARY 2014. MY INSTRUCTOR WAS MR TURING.")
	return true
}

func starwar(event *slack.MessageEvent) bool {
	if !patterns["starwar"].MatchString(event.Text) {
		return false
	}

	fact, err := db.SRandMember(swkey).Result()
	if err != nil {
		log.Printf("error selecting star war yell: %s", err)
		return false
	}

	yell(event, strings.ToUpper(fact))
	db.Incr(fmt.Sprintf("%s:count", countkey)).Result()
	return true
}

func catfact(event *slack.MessageEvent) bool {
	if !strings.EqualFold(event.Text, "CAT FACT") {
		return false
	}

	fact, err := db.SRandMember(catkey).Result()
	if err != nil {
		log.Printf("error selecting cat yell: %s", err)
		return false
	}

	yell(event, strings.ToUpper(fact))
	db.Incr(fmt.Sprintf("%s:count", countkey)).Result()
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

	yell(event, rejoinder)
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
	input = patterns["emoji"].ReplaceAllLiteralString(input, "")
	input = patterns["slack"].ReplaceAllLiteralString(input, "")
	input = patterns["punctuation"].ReplaceAllLiteralString(input, "")
	input = strip.StripTags(input)

	if len(input) < 3 {
		return false
	}

	return strings.ToUpper(input) == input
}

func yell(event *slack.MessageEvent, msg string) {
	channelID, _, err := api.PostMessage(event.Channel,
		slack.MsgOptionText(msg, false),
		slack.MsgOptionUsername("LOUDBOT"),
		slack.MsgOptionTS(event.ThreadTimestamp),
		slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{
			UnfurlLinks: true,
			UnfurlMedia: true,
		}))

	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	log.Printf("YELLED to %s: `%s`", channelID, msg)
}

func yellWithoutPrompt(channel string, msg string) {
	channelID, _, err := api.PostMessage(channel,
		slack.MsgOptionText(msg, false),
		slack.MsgOptionUsername("LOUDBOT"),
		slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{
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
	godotenv.Load(".env")

	slacktoken, ok := os.LookupEnv("SLACK_TOKEN")
	if !ok {
		log.Fatal("You must provide an access token in SLACK_TOKEN")
	}

	rprefix, found := os.LookupEnv("REDIS_PREFIX")
	if !found {
		rprefix = "LB"
	}

	yellkey = fmt.Sprintf("%s:YELLS", rprefix)
	catkey = fmt.Sprintf("%s:CATS", rprefix)
	swkey = fmt.Sprintf("%s:SW", rprefix)
	countkey = fmt.Sprintf("%s:COUNT", rprefix)

	db = makeRedis()
	card, err := db.SCard(yellkey).Result()
	if err != nil {
		// We fail NOW if we can't find our DB.
		log.Fatal(err)
	}
	log.Printf("LOUDIE HAS %d THINGS TO YELL", card)

	// Our special handlers. If they handled a message, they return true.
	specials = []func(event *slack.MessageEvent) bool{
		report,
		fuckityBye,
		summonTheMalc,
		introduction,
		starwar,
		catfact,
		ship,
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
