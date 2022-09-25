package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"lab3_2/src/proto"
	"net/http"
	"os"
	"os/signal"
	"strings"

	log "github.com/mgutz/logxi/v1"
)

var conf proto.Config
var subsTable map[string]bool

const url string = "http://localhost:8000"

// Обновляет список известных пиров
func updatePeers() {
	resp, err := http.Get(url + "/updatepeers")
	if err != nil {
		log.Error("request failed", "reason", err.Error())
	} else {
		body, _ := bufio.NewReader(resp.Body).ReadBytes('\u00C6')
		var foo proto.Config
		if err := json.Unmarshal(body, &foo); err != nil {
			log.Error(err.Error())
			return
		}
		conf.KnownHosts = foo.KnownHosts
	}
}

// Подписывает этот пир на сообщения другого пира
func subscribe(peer proto.ConfigForShare) {
	buf, _ := json.Marshal(conf)
	body := strings.NewReader(string(buf))
	if _, err := http.Post("http://"+peer.Addr[:len(peer.Addr)-1]+"/sub", "application/json", body); err != nil {
		log.Error("failed to send sub request", "reason", err)
	} else {
		log.Info("succesfully subed to", "name", peer.Name)
	}
}

// Отписывает этот пир от сообщений другого
func unsubscribe(peer proto.ConfigForShare) {
	buf, _ := json.Marshal(conf)
	body := strings.NewReader(string(buf))
	if _, err := http.Post("http://"+peer.Addr[:len(peer.Addr)-1]+"/unsub", "application/json", body); err != nil {
		log.Error("failed to send sub request", "reason", err)
	} else {
		log.Info("succesfully subed to", "name", peer.Name)
	}
}

// Отправка сообщения подписчику
func sendMessage(peer proto.ConfigForShare, msg proto.Message) {
	buf, _ := json.Marshal(msg)
	body := strings.NewReader(string(buf))
	if _, err := http.Post("http://"+peer.Addr[:len(peer.Addr)-1]+"/message", "application/json", body); err != nil {
		log.Error("failed to send message", "dest", peer.Name, "reason", err)
	}
}

func main() {
	subsTable = make(map[string]bool)

	fmt.Print("Please enter your name: ")
	fmt.Scan(&conf.Name)

	if encoded, err := json.Marshal(conf); err != nil {
		fmt.Println(err)
	} else {
		resp, err := http.Post(url+"/register", "application/json", strings.NewReader(string(encoded)))
		if err != nil {
			log.Error(err.Error())
			return
		}
		body, _ := bufio.NewReader(resp.Body).ReadBytes('\u00C6')
		if err := json.Unmarshal(body, &conf); err != nil {
			log.Error(err.Error())
			return
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/sub", handleSub)
	mux.HandleFunc("/unsub", handleUnsub)
	mux.HandleFunc("/message", handleMessage)
	server := http.Server{
		Addr:    conf.Addr[:len(conf.Addr)-1],
		Handler: mux,
	}
	go server.ListenAndServe()

	printGuide()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Println(sig)
			http.Get(url + "/logout")
			os.Exit(1)
		}
	}()

	for {
		fmt.Printf("command = ")
		var command string
		reader := bufio.NewReader(os.Stdin)
		buf, _, _ := reader.ReadLine()
		command = string(buf)

		switch command {
		case "exit":
			http.Get(url + "/logout")
			return
		case "show peers":
			updatePeers()
			fmt.Println("Available peers:")
			for _, peer := range conf.KnownHosts {
				if peer.Name == conf.Name {
					fmt.Println(peer.Name + "(you)")
				} else {
					fmt.Println(peer.Name)
				}
			}
		case "sub":
			var name string
			var peer proto.ConfigForShare
			fmt.Print("enter peer name: ")
			fmt.Scan(&name)
			for _, host := range conf.KnownHosts {
				if host.Name == name && name != conf.Name {
					peer = host
				}
			}
			if peer != (proto.ConfigForShare{}) {
				subscribe(peer)
			} else {
				log.Error("failed to find peer", "name", name)
			}
		case "unsub":
			var name string
			var peer proto.ConfigForShare
			fmt.Print("enter peer name: ")
			fmt.Scan(&name)
			for _, host := range conf.KnownHosts {
				if host.Name == name && name != conf.Name {
					peer = host
				}
			}
			if peer != (proto.ConfigForShare{}) {
				unsubscribe(peer)
			} else {
				log.Error("failed to find peer", "name", name)
			}
		case "send":
			var message proto.Message
			reader := bufio.NewReader(os.Stdin)
			buf, _, _ := reader.ReadLine()
			message.Content = string(buf)
			for _, host := range conf.KnownHosts {
				if subsTable[host.Name] {
					message.Sender = conf.Name
					fmt.Println(message.Sender)
					sendMessage(host, message)
				}
			}
			log.Info("Sent to all subs succesfully")
		case "help":
			printGuide()
		default:
			log.Warn("unknown command, please try again")
		}

	}

}
