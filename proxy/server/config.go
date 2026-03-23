package server

import (
	"github.com/thomasgame/trojan-go-extra/config"
	"github.com/thomasgame/trojan-go-extra/proxy/client"
)

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return new(client.Config)
	})
}
