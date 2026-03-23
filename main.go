package main

import (
	"flag"

	_ "github.com/thomasgame/trojan-go-extra/component"
	"github.com/thomasgame/trojan-go-extra/log"
	"github.com/thomasgame/trojan-go-extra/option"
)

func main() {
	flag.Parse()
	for {
		h, err := option.PopOptionHandler()
		if err != nil {
			log.Fatal("invalid options")
		}
		err = h.Handle()
		if err == nil {
			break
		}
	}
}
