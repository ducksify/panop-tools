package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/net/publicsuffix"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <domain>")
	}
	domain := os.Args[1]

	etld, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		log.Fatal(err)
	}

	if domain == etld {
		fmt.Println("is-apex")
	} else {
		fmt.Println("not-apex")
	}
}
