package main

import (
	"net/url"
)

func main() {
	u := `%2fReplayFile%2f2014-11-8%2f141108_%5bUD%5dLusiria_VS_%5bNE%5dG-one_TwistedMeadows_RN.w3g`
	u2, _ := url.QueryUnescape(u)
	println(u2)
}
