package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"time"

	"github.com/sacOO7/gowebsocket"
)

var streaming bool = false
var streamPath string = "/tmp/lofi-stream.ts"

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	sessionId := strconv.Itoa(rand.Intn(2147483647))
	environmentPath := os.Getenv("TEMP")
	if environmentPath != "" {
		streamPath = environmentPath + "\\lofi-stream.ts"
	}
	ioutil.WriteFile(streamPath, []byte(""), 0644)
	streamFile, errorObject := os.OpenFile(streamPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 6400)
	if errorObject != nil {
		fmt.Println("Unable to open stream file")
	}
	defer streamFile.Close()

	socket := gowebsocket.New("ws://lofi-server.herokuapp.com/" + sessionId)
	socket.OnConnected = func(socket gowebsocket.Socket) {
		fmt.Println("Successfully connected to server")
		go playAudio()
	}
	socket.OnDisconnected = func(_ error, socket gowebsocket.Socket) {
		fmt.Println("Disconnected from server")
		socket.Connect()
	}
	socket.OnBinaryMessage = func(data []byte, socket gowebsocket.Socket) {
		streaming = true
		fmt.Println("Receiving stream data...")
		streamFile.WriteString(string(data))
		socket.SendText("pong")
	}
	socket.Connect()

	for {
		select {
		case <-interrupt:
			fmt.Println("Stopping...")
			socket.Close()
			return
		}
	}
}

func playAudio() {
	for !streaming {
		time.Sleep(1 * time.Second)
	}
	for {
		exec.Command("mpv", "--no-video", streamPath).Output()
		fmt.Println("Restarting audio player...")
	}
}
