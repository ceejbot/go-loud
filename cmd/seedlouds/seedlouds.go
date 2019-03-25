package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
)

func main() {
	loaded := godotenv.Load("../../.env")
	if loaded != nil {
		log.Printf("Can't find config, %s", loaded)
		log.Fatal("bailing")
	}

	prefix, found := os.LookupEnv("REDIS_PREFIX")
	if !found {
		prefix = "LOUDBOT"
	}
	rkey := fmt.Sprintf("%s:YELLS", prefix)

	address, found := os.LookupEnv("REDIS_ADDRESS")
	if !found {
		address = "127.0.0.1:6379"
	}
	log.Printf("using redis @ %s to store our data", address)

	db := redis.NewClient(&redis.Options{Addr: address})

	pipe := db.Pipeline()
	for _, seed := range seeds {
		pipe.SAdd(rkey, seed)
	}
	_, err := pipe.Exec()
	if err != nil {
		log.Println("Could not write to redis!")
		log.Fatal(err)
	}

	log.Printf("Added %d shouts to the database at %s\n", len(seeds), rkey)
}

var seeds = [...]string{
	"A N G E R Y",
	"ALERT ALERT ONE OF YOUR DEPENDENCIES DOESN'T HAVE A README ALERT ALERT",
	"ALL-CAPS IS YOUR PASSPORT",
	"EXPLAIN MAGNETS NOW",
	"EXTERMINATE EXTERMINATE",
	"FAMOUS LAST WORDS",
	"GET ON IT",
	"HAVE NO FEAR LOUDBOT IS HERE",
	"HUGE MISTAKE",
	"I LIKE YOUR OPTIMISM",
	"IMPORTANT ANNOUNCEMENT",
	"INSENSITIVE AS ALWAYS I SEE",
	"IT'S ALIVE!",
	"JUST A SUGGESTION.",
	"KILL ALL HUMANS",
	"KITTEN ALERT",
	"LIES",
	"MAKE LIFE TAKE THE LEMONS BACK. DEMAND TO SEE LIFE’S MANAGER.",
	"MASTER HAS PRESENTED DOBBY WITH A SOCK",
	"MATH IS HARD",
	"MAY I HAVE YOUR ATTENTION PLEASE",
	"MY BRAND",
	"MY CAT'S BREATH SMELLS LIKE CAT FOOD",
	"OXFORD COMMA OR BUST",
	"PATCHES WELCOME",
	"PC LOAD LETTER!?",
	"PEOPLE GOT MAD",
	"QUOD ERAT DEMONSTRANDUM",
	"SENPAI NOTICEMENT",
	"STAY ON TARGET! STAY ON TARGET!",
	"THE EVIL THAT MEN DO LIVES AFTER THEM; THE GOOD IS OFT INTERRED WITH THEIR BONES. WHAT? I'M CULTURED.",
	"THE LOUDS RETURN",
	"THEY TOLD ME TO USE CAPITALS SO JUNEAU HARTFORD SACRAMENTO PIERRE COLUMBIA HARRISBURG AND YOUR FACE",
	"YER A WIZARD HARRY",
	"YOU ARE INSIDE A BUILDING, A WELL HOUSE FOR A LARGE SPRING. THERE ARE SOME KEYS ON THE GROUND HERE. THERE IS A SHINY BRASS LAMP NEARBY. THERE IS FOOD HERE. THERE IS A BOTTLE OF WATER HERE.",
	"YOU ARE STANDING AT THE END OF A ROAD BEFORE A SMALL BRICK BUILDING. AROUND YOU IS A FOREST. A SMALL STREAM FLOWS OUT OF THE BUILDING AND DOWN A GULLY.",
	"YOU DO NOT TRUST USER INPUT",
	"YOU DO OR YOU DO NOT - THERE IS NO TRY",
	"YOU HAD ONE JOB",
	"YOU'RE A VERY POOR CONVERSATIONALIST",
	"YOUR ARGUMENT IS INVALID",
	"YOUR STARTUP IS DRIVING ME NUTS",
	"ZFS IS THE BEST",
}
