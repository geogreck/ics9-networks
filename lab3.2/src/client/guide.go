package main

import log "github.com/mgutz/logxi/v1"

func printGuide() {
	log.Info("Available commands are:",
		"exit", "leave network",
		"show peers", "show all peers online",
		"sub", "sub to the (name) peer",
		"unsub", "unsub of the (name) peer",
		"send", "send message to all your followers",
		"help", "show this guide",
	)
}
