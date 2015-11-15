package slackbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/masonforest/slackbot/Godeps/_workspace/src/github.com/gorilla/schema"
	"io"
	"net/http"
	"os"
)

type RequestData struct {
	Token       string `schema:"token"`
	TeamId      string `schema:"team_id"`
	TeamDomain  string `schema:"team_domain"`
	ChannelId   string `schema:"channel_id"`
	ChannelName string `schema:"channel_name"`
	UserId      string `schema:"user_id"`
	UserName    string `schema:"user_name"`
	Command     string `schema:"command"`
	Text        string `schema:"text"`
	ResponseUrl string `schema:"response_url"`
}

type Request struct {
	w    http.ResponseWriter
	r    *http.Request

	Data *RequestData
}

type Response struct {
	Text string
}

var EmptyResponse = Response{}

func (r Response) toString() string {
  var s string

  if(r.Text == ""){
    s = ""
  } else {
	  data := map[string]string{"text": r.Text}
	  bs, _ := json.Marshal(data)
    s = string(bs)
  }
	return s
}

func (req Request) Respond(resp Response) {
	var byteString = []byte(resp.toString())
	post, err := http.NewRequest("POST", req.Data.ResponseUrl, bytes.NewBuffer(byteString))

	client := &http.Client{}
	pr, err := client.Do(post)
	if err != nil {
		panic(err)
	}
	defer pr.Body.Close()
}

type Command interface {
	Respond(Request) string
}

type Server struct {
	commands map[string]func(Request) Response
}

func NewServer() *Server {
	return &Server{commands: make(map[string]func(Request) Response)}
}
func (s *Server) AddCommand(name string, command func(Request) Response) {
	s.commands[name] = command
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "", 400)
		return
	}

	data := &RequestData{}
	decoder := schema.NewDecoder()

	err = decoder.Decode(data, r.PostForm)
	if err != nil {
		http.Error(w, fmt.Sprintf("%#v", err), 500)
		return
	}

	request := Request{w: w, r: r, Data: data}

	io.WriteString(w, s.commands[request.Data.Command](request).toString())
}

func (s Server) Boot() {
	http.HandleFunc("/", s.ServeHTTP)
	http.ListenAndServe(fmt.Sprint(":", os.Getenv("PORT")), nil)
}
