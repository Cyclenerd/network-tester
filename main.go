/*
Copyright 2025 Nils Knieling. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections for demo purposes
	},
}

type CommandRequest struct {
	Command  string `json:"command"`
	Target   string `json:"target"`
	Option   string `json:"option"`
	Parallel string `json:"parallel"`
	Duration string `json:"duration"`
}

type NetworkInterface struct {
	Name       string   `json:"name"`
	Addresses  []string `json:"addresses"`
	MACAddress string   `json:"macAddress"`
}

func main() {
	r := mux.NewRouter()

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// API endpoints
	r.HandleFunc("/ws", handleWebSocket)

	// Serve the main page
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	// Serve the favicon
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/favicon.ico")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var cmdReq CommandRequest
		if err := json.Unmarshal(p, &cmdReq); err != nil {
			log.Println(err)
			continue
		}

		go executeCommand(conn, messageType, cmdReq)
	}
}

func executeCommand(conn *websocket.Conn, messageType int, cmdReq CommandRequest) {
	var cmd *exec.Cmd

	switch cmdReq.Command {
	case "ping":
		cmd = exec.Command("ping", "-c", "10", cmdReq.Target)
	case "curl":
		cmd = exec.Command("curl", "-s", "-i", cmdReq.Target)
	case "nslookup":
		cmd = exec.Command("nslookup", cmdReq.Target)
	case "nmap":
		cmd = exec.Command("nmap", cmdReq.Target)
	case "dig":
		cmd = exec.Command("dig", cmdReq.Option, cmdReq.Target)
	case "traceroute":
		cmd = exec.Command("traceroute", cmdReq.Target)
	case "iperf3":
		args := []string{"-c", cmdReq.Target}
		if cmdReq.Parallel != "" {
			args = append(args, "-P", cmdReq.Parallel)
		}
		if cmdReq.Duration != "" {
			args = append(args, "-t", cmdReq.Duration)
		}
		cmd = exec.Command("iperf3", args...)
	case "ifconfig":
		cmd = exec.Command("ifconfig")
	default:
		_ = conn.WriteMessage(messageType, []byte("Unsupported command"))
		return
	}

	// Send command being executed
	_ = conn.WriteMessage(messageType, []byte(fmt.Sprintf("$ %s\n\n", strings.Join(cmd.Args, " "))))

	// Create pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = conn.WriteMessage(messageType, []byte("Error creating stdout pipe: "+err.Error()))
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		_ = conn.WriteMessage(messageType, []byte("Error creating stderr pipe: "+err.Error()))
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		_ = conn.WriteMessage(messageType, []byte("Error starting command: "+err.Error()))
		return
	}

	// Stream in real-time
	go streamOutput(conn, messageType, stdout)
	go streamOutput(conn, messageType, stderr)

	// Wait for command to finish
	if err := cmd.Wait(); err != nil {
		_ = conn.WriteMessage(messageType, []byte("\nCommand exited with error: "+err.Error()))
		return
	}
}

func streamOutput(conn *websocket.Conn, messageType int, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		_ = conn.WriteMessage(messageType, []byte(line+"\n"))
		// time.Sleep(10 * time.Millisecond) // Small delay to prevent flooding
	}
}
