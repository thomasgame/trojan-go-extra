//go:build api || full
// +build api full

package build

import (
	_ "github.com/thomasgame/trojan-go-extra/api/control"
	_ "github.com/thomasgame/trojan-go-extra/api/service"
)
