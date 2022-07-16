package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"time"

	"github.com/hugolgst/rich-go/client"
	"github.com/sacOO7/gowebsocket"
)

var (
	streaming  bool = true
	streamPipe io.Writer
	received   int64

	optionNoPresence bool
	optionVideo      bool
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	rand.Seed(time.Now().UnixNano())
	sessionId := strconv.Itoa(rand.Intn(2147483648))
	for _, argument := range os.Args {
		if argument == "--no-presence" {
			optionNoPresence = true
		} else if argument == "--video" {
			optionVideo = true
		}
	}

	if !optionNoPresence {
		err := client.Login("997806917588095066")
		if err != nil {
			fmt.Println("Unable to set status: " + err.Error())
		}
		startTime := time.Now()
		err = client.SetActivity(client.Activity{
			State:      "Listening to lofi music",
			LargeImage: "image",
			Timestamps: &client.Timestamps{
				Start: &startTime,
			},
		})
		if err != nil {
			fmt.Println("Unable to set status: " + err.Error())
		}
	}

	socket := gowebsocket.New("ws://lofi-server.herokuapp.com/" + sessionId)
	socket.OnConnected = func(socket gowebsocket.Socket) {
		fmt.Println("Successfully connected to server")
		go playAudio()
	}
	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		if streaming {
			fmt.Println("\nDisconnected from server: " + err.Error())
			socket.Connect()
		}
	}
	socket.OnBinaryMessage = func(data []byte, socket gowebsocket.Socket) {
		fmt.Printf("\r[%v] Receiving stream data...", received)
		io.WriteString(streamPipe, string(data))
		socket.SendText("pong")
		received++
	}
	socket.Connect()

	for {
		select {
		case <-interrupt:
			fmt.Println("\nStopping...")
			streaming = false
			socket.Close()
			return
		}
	}
}

func playAudio() {
	for {
		command := exec.Command("mpv", "--no-video", "-")
		if optionVideo {
			command = exec.Command("mpv", "-")
		}
		streamPipe, _ = command.StdinPipe()
		command.Run()
		if !streaming {
			return
		}
		fmt.Println("\nRestarting audio player...")
	}
}
