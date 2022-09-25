package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"lab3_2/src/proto"
	"net/http"

	log "github.com/mgutz/logxi/v1"
)

// Обработчик подписки на этот пир другим пиром
func handleSub(w http.ResponseWriter, r *http.Request) {
	var conf proto.ConfigForShare
	body, err := bufio.NewReader(r.Body).ReadBytes('\u00C6')
	if err != nil && err != io.EOF {
		log.Error(err.Error())
		fmt.Fprintf(w, "I'm chereshnya")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(body, &conf); err != nil {
		log.Error(err.Error())
		fmt.Fprintf(w, "I'm teepot")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	subsTable[conf.Name] = true
	fmt.Print("\r\r\r\r\r\r")
	log.Info("\nYou've got new subsciber", "name", conf.Name)
	fmt.Print("command = ")
}

// Обработчик отписки от этого пира другим пиром
func handleUnsub(w http.ResponseWriter, r *http.Request) {
	var conf proto.ConfigForShare
	body, err := bufio.NewReader(r.Body).ReadBytes('\u00C6')
	if err != nil && err != io.EOF {
		log.Error(err.Error())
		fmt.Fprintf(w, "I'm chereshnya")
		return
	}
	if err := json.Unmarshal(body, &conf); err != nil {
		log.Error(err.Error())
		fmt.Fprintf(w, "I'm teepot")
		return
	}
	subsTable[conf.Name] = false
	fmt.Print("\r\r\r\r\r\r")
	log.Info("\nYou've lost subsciber", "name", conf.Name)
	fmt.Print("command = ")
}

// Обработчик приема сообщения
func handleMessage(w http.ResponseWriter, r *http.Request) {
	var msg proto.Message
	body, err := bufio.NewReader(r.Body).ReadBytes('\u00C6')
	if err != nil && err != io.EOF {
		log.Error(err.Error())
		fmt.Fprintf(w, "I'm chereshnya")
		return
	}
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusTeapot)
		fmt.Fprintf(w, "I'm teepot")
		return
	}
	fmt.Print("\r\r\r\r\r\r")
	log.Info("\nYou've got new message", "From", msg.Sender, "Content", msg.Content)
	fmt.Print("command = ")
}
