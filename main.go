package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type config struct {
	WebsocketHost string `json:"websocket_host"`
	WebsocketPath string `json:"websocket_path"`
}

type message struct {
	Nick      string        `json:"nick"`
	Features  []interface{} `json:"features"`
	Timestamp int64         `json:"timestamp"`
	Data      string        `json:"data"`
}

type users struct {
	Users []struct {
		Nick     string        `json:"nick"`
		Features []interface{} `json:"features"`
	} `json:"users"`
	Connectioncount int `json:"connectioncount"`
}

type quit struct {
	Nick      string        `json:"nick"`
	Features  []interface{} `json:"features"`
	Timestamp int64         `json:"timestamp"`
}

type join struct {
	Nick      string        `json:"nick"`
	Features  []interface{} `json:"features"`
	Timestamp int64         `json:"timestamp"`
}

func main() {
	msgCount := 0

	//load config
	var con config
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalln("no config file found")
	} else {
		err = json.Unmarshal(file, &con)
		if err != nil {
			log.Fatalf("malformed configuration file: %v\n", err)
		}
	}

	ws := connect(con.WebsocketHost, con.WebsocketPath)
	currentDate := time.Now().Format("2006-01-02")
	createfolder(`./` + currentDate + `/`)
	f, _ := os.OpenFile(`./`+currentDate+`/`+"logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	for {
		_, m, err := ws.ReadMessage()
		//handle error by trying to reconnect every 100ms
		for err != nil {
			ws = connect(con.WebsocketHost, con.WebsocketPath)
			_, m, err = ws.ReadMessage()
			time.Sleep(time.Millisecond * 100)
		}
		messageString := string(m)
		if len(messageString) > 4 && messageString[:4] == "MSG " {
			msgCount++
			fmt.Print("\rmessages logged: " + strconv.Itoa(msgCount))
			//create new folder and log file for new day
			if time.Now().Format("2006-01-02") != currentDate {
				currentDate = time.Now().Format("2006-01-02")
				createfolder(`./` + currentDate + `/`)
				f, _ = os.OpenFile(`./`+currentDate+`/`+"logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			}
			//string to struct
			var ms message
			json.Unmarshal([]byte(messageString[4:]), &ms)
			//write to logs.txt
			tm := time.Unix(ms.Timestamp/1000, 0)
			s := "[" + tm.String() + "] " + ms.Nick + ": " + ms.Data
			f.WriteString(s + "\n")
			//write userlog
			writeToFile(s, ms.Nick)
		}
	}
}

func createfolder(s string) {
	if _, err := os.Stat(s); err != nil {
		fmt.Println("\nmaking folder " + s)
		os.MkdirAll(s, os.ModePerm)
	}
}

func writeToFile(message string, nick string) {
	currentTime := time.Now()
	createfolder(`./` + currentTime.Format("2006-01-02") + `/userlogs/`)
	f, err := os.OpenFile(`./`+currentTime.Format("2006-01-02")+`/userlogs/`+nick+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(message + "\n"); err != nil {
		log.Println(err)
	}
}

func connect(host string, path string) *websocket.Conn {
	var wsURL = url.URL{Scheme: "wss", Host: host, Path: path}
	var dialer *websocket.Dialer
	header := http.Header{}
	ws, _, err := dialer.Dial(wsURL.String(), header)
	if err != nil {
		fmt.Println(err)
	}
	return ws
}
