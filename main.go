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

	"github.com/gdamore/tcell"
	"github.com/gorilla/websocket"
)

type config struct {
	Hosts []host `json:"hosts"`
}

type host struct {
	WebsocketHost string `json:"websocket_host"`
	WebsocketPath string `json:"websocket_path"`
	minutes       []float64
	currentMinute int
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

type monitorState struct {
	h           *host
	count       float64
	hasError    bool
	Error       string
	lastMessage string
}

var currentDate string

func main() {
	currentDate = time.Now().Format("2006-01-02")

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

	//set up tcell
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	if e = s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	defer s.Fini()
	s.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite))
	s.Clear()

	quit := make(chan struct{})

	//set up hotkeys
	go func() {
		for {
			ev := s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					close(quit)
					return
				case tcell.KeyCtrlL:
					s.Sync()
				}
			case *tcell.EventResize:
				s.Sync()
			}
		}
	}()

	//launch monitor for host
	stateChan := make(chan monitorState)
	var meme monitorState
	dataChan := make(chan []float64, 3)
	for _, host := range con.Hosts {
		go monitor(&host, stateChan)
		go monitorMpm(15, &meme, dataChan)
	}
	text := "waiting for messages"
	content := []float64{1}
	//manage ws output
	for {
		select {
		case <-quit:
			return
		case meme = <-stateChan:
			text = ""
			text += meme.h.WebsocketHost + "\n"
			text += "===========================\n\n"
			text += "messages received:  " + floatToString(meme.count) + "\n"
			text += "last message:  " + "\n"
			text += "  " + meme.lastMessage + "\n"
			if meme.hasError {
				text += "ERROR:\n"
				text += "	" + meme.Error + "\n"
			}
			text += "\n"
			text += sliceToString(content) + "\n"
		case content = <-dataChan:
		case <-time.After(time.Millisecond * 50):
		}
		w, h := s.Size()
		textMatrix := textToMatrix(text)
		squished := squash(content, h-1)
		graph := printAsGraphSetX(squished, w/3*2)
		newmatrix := combineMatrix(graph, 0, h-len(graph), textMatrix, w/3*2+1, 0)
		writeToScreen(s, newmatrix)
	}
}

//monitors a websocket and sends back an update for every message and error
func monitor(h *host, c chan<- monitorState) {
	ws, err := connect(h.WebsocketHost, h.WebsocketPath)
	if err != nil {
		return
	}

	//set variables
	h.minutes = make([]float64, 24*60)
	var state monitorState
	state.count = 0
	state.hasError = false
	state.h = h
	timeC := time.Now()
	_, min, _ := timeC.Clock()
	state.h.currentMinute = min

	path := `./` + h.WebsocketHost + `/` + currentDate + `/`
	createfolder(path + `userlogs/`)
	f, _ := os.OpenFile(path+"logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	for {
		var receivedMsg []byte
		conSucc := true
		_, receivedMsg, err := ws.ReadMessage()

		if err != nil {
			conSucc = false
		}
		errcounter := 0
		//try to reconnect after error
		for !conSucc {
			state.hasError = true
			state.Error = "error trying to read from websocket, try: " + strconv.Itoa(errcounter)
			c <- state
			time.Sleep(time.Millisecond * time.Duration(100*errcounter*3))
			ws, err = connect(h.WebsocketHost, h.WebsocketPath)
			if err == nil {
				_, receivedMsg, err = ws.ReadMessage()
				if err != nil {
					conSucc = true
					state.Error = "successfully reconnected"
					c <- state
					state.hasError = false
				}
			}
			if !conSucc {
				errcounter++
				if errcounter >= 50 {
					state.Error = "aborting connection"
					c <- state
					return
				}
			}
		}

		//process message
		messageString := string(receivedMsg)
		if len(messageString) > 4 && messageString[:4] == "MSG " {
			//string to struct
			var ms message
			json.Unmarshal([]byte(messageString[4:]), &ms)

			//update variables
			state.count = state.count + 1
			state.lastMessage = ms.Nick + ": " + ms.Data
			timeC := time.Now()
			hr, min, _ := timeC.Clock()
			state.h.minutes[hr*min]++

			c <- state

			//if a new day starts we make a new folder and update the date
			if timeC.Format("2006-01-02") != currentDate {
				state.count = 0
				currentDate = timeC.Format("2006-01-02")
				path = `./` + h.WebsocketHost + `/` + currentDate + `/`
				createfolder(path + `userlogs/`)
				f, _ = os.OpenFile(path+"logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			}
			//format message
			tm := time.Unix(ms.Timestamp/1000, 0).UTC()
			s := "[" + tm.String() + "] " + ms.Nick + ": " + ms.Data
			//write to global log
			f.WriteString(s + "\n")
			//write userlog
			writeToFile(s, ms.Nick, path+`userlogs/`)
		}
	}
}

//creates folder
func createfolder(s string) {
	if _, err := os.Stat(s); err != nil {
		os.MkdirAll(s, os.ModePerm)
	}
}

//writes to user log file
func writeToFile(message string, nick string, folderpath string) {
	f, err := os.OpenFile(folderpath+nick+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(message + "\n"); err != nil {
		log.Println(err)
	}
}

//connects to ws and returns the connection
func connect(host string, path string) (*websocket.Conn, error) {
	var wsURL = url.URL{Scheme: "wss", Host: host, Path: path}
	var dialer *websocket.Dialer
	header := http.Header{}
	ws, _, err := dialer.Dial(wsURL.String(), header)
	return ws, err
}

//writes the content of the array to the screen
func writeToScreen(s tcell.Screen, data [][]rune) {
	s.Clear()

	w, h := s.Size()
	if w == 0 || h == 0 {
		return
	}

	for y, row := range data {
		for x, column := range row {
			if x < w && y < h {
				s.SetCell(x, y, tcell.StyleDefault, column)
			}
		}
	}
	s.Show()
}

//sends message count for interval every n seconds
func monitorMpm(interval int, h *monitorState, c chan<- []float64) {
	output := make([]float64, 1)
	time.Sleep(time.Second * time.Duration(interval))
	msgcount := h.count
	output[0] = msgcount
	c <- output
	for {
		time.Sleep(time.Second * time.Duration(interval))
		new := h.count - msgcount
		msgcount += new
		output = append(output, float64(new))
		c <- output
	}
}
