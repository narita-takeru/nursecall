package nursecall

import (
	"fmt"
	"encoding/json"
	"time"
	"os"
	"strings"
	"net/http"
	"math/rand"
	"log"
	"io/ioutil"
	"bytes"
	"crypto/tls"
)

type apiParam struct {
	CallToken string `json:"call_token"`
	ProcessID int64 `json:"process_id"`
	Name string `json:"name"`
}

type notifier struct {
	URL string
	CallToken string
	Debug bool

	started bool
	interval int
	isDone bool
	httpClient *http.Client

	apiParam apiParam
}

func buildName() string {
	tokens := os.Args[1:]
	return strings.Join(tokens, "\t")
}

func buildProcessID() int64 {
	rand.Seed(time.Now().UnixNano())
	r := rand.Int63() / 100000 * 100000
	return r + int64(os.Getpid())
}

func (n* notifier) Start() {

	if len(n.CallToken) <= 0 {
		return
	}

	n.interval = 60

	n.apiParam.CallToken = n.CallToken
	n.apiParam.ProcessID = buildProcessID()
	n.apiParam.Name = buildName()

	input, err := json.Marshal(n.apiParam)
	if err != nil {
		log.Println("Can't build params.")
		return
	}

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	n.httpClient = &http.Client{Transport: tr}

	res, err := n.httpClient.Post(n.URL, "application/json", bytes.NewBuffer(input))
	if err != nil {
		if n.Debug {
			log.Println(err)
		}

		return
	}

	if n.Debug {
		bs, _ := ioutil.ReadAll(res.Body)
		log.Println(string(bs))
	}

	n.started = true
	go n.heartbeat(n.interval)
}

func (n* notifier) doPut(path string) {

	input, err := json.Marshal(n.apiParam)
	if err != nil {
		log.Println("Can't build params.")
		return
	}

	req, _ := http.NewRequest(
		http.MethodPut,
		n.URL + path,
		strings.NewReader(string(input)),
	)

	req.Header.Set("Content-Type", "application/json")

	res, err := n.httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	bs, _ := ioutil.ReadAll(res.Body)
	_ = bs
}

func (n* notifier) heartbeat(interval int) {
	for !n.isDone {
		time.Sleep(time.Second * time.Duration(interval))
		if !n.isDone {
			n.doPut("")
		}
	}
}

func (n* notifier) Done(exitStatus int) {
	if n.started {
		n.doPut("/done")
	}
}

func (n* notifier) Error(e string) {
	if n.started {
		n.doPut("/error")
	}
}
