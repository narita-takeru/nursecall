package nursecall

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type APIParam struct {
	CallToken string `json:"call_token"`
	ProcessID int64  `json:"process_id"`
	Name      string `json:"name"`
}

func (a *APIParam) genName(tokens []string) {
	a.Name = strings.Join(tokens, "\t")
}

func (a *APIParam) genProcessID() {
	rand.Seed(time.Now().UnixNano())
	r := rand.Int63() / 100000 * 100000
	a.ProcessID = r + int64(os.Getpid())
}

func newAPIParam(tokens []string) APIParam {
	a := APIParam{CallToken: os.Getenv("NURSECALL_CALL_TOKEN")}
	a.genName(tokens)
	a.genProcessID()
	return a
}

type Notifier struct {
	Debug bool

	intervalHeartBeat int

	HTTPClient  *http.Client
	EndpointURL string
	APIParam    APIParam
}

const (
	defaultEndpointURL = "https://nursecall.io/api/v1/progresses"
)

func NewNotifier(tokens []string) Notifier {
	n := Notifier{
		Debug: "TRUE" == os.Getenv("NURSECALL_DEBUG"),

		intervalHeartBeat: 60,

		HTTPClient:  &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}},
		EndpointURL: os.Getenv("NURSECALL_ENDPOINT"),
		APIParam:    newAPIParam(tokens),
	}

	if len(n.EndpointURL) == 0 {
		n.EndpointURL = defaultEndpointURL
	}
	return n
}

func (n *Notifier) Validate() error {
	if len(n.APIParam.CallToken) == 0 {
		return errors.New("No CallToken")
	}
	return nil
}

func (n *Notifier) Start() error {
	input, err := json.Marshal(n.APIParam)
	if err != nil {
		return err
	}

	res, err := n.HTTPClient.Post(n.EndpointURL, "application/json", bytes.NewBuffer(input))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if n.Debug {
		bs, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		log.Println(string(bs))
	}

	// TODO catch error for safety use nursecall
	go n.heartbeat()

	return nil
}

func (n *Notifier) doPut(path string) error {

	input, err := json.Marshal(n.APIParam)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPut,
		n.EndpointURL+path,
		strings.NewReader(string(input)),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := n.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if n.Debug {
		bs, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		log.Println(string(bs))
	}

	return nil
}

func (n *Notifier) heartbeat() {
	for {
		time.Sleep(time.Second * time.Duration(n.intervalHeartBeat))
		if err := n.doPut(""); err != nil {
			log.Println(err)
		}
	}
}

func (n *Notifier) Done() error {
	if err := n.doPut("/done"); err != nil {
		return err
	}
	return nil
}

func (n *Notifier) Error() error {
	if err := n.doPut("/error"); err != nil {
		return err
	}
	return nil
}
