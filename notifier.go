package nursecall

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Notifier struct {
	Debug bool

	intervalHeartBeat int

	HTTPClient  *http.Client
	EndpointURL string

	jobID string
}

const (
	defaultEndpointURL = "https://api.nursecall.run/jobs"
)

func NewNotifier(tokens []string) Notifier {
	n := Notifier{
		Debug:             "TRUE" == os.Getenv("NURSECALL_DEBUG"),
		intervalHeartBeat: 0,
		HTTPClient:        &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}},
		EndpointURL:       defaultEndpointURL,
	}

	return n
}

func (n *Notifier) Start(cmdStr string) error {
	input := map[string]interface{}{
		"call_token": os.Getenv("NURSECALL_CALL_TOKEN"),
		"path":       cmdStr,
	}

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return err
	}

	res, err := n.HTTPClient.Post(n.EndpointURL, "application/json", bytes.NewBuffer(inputBytes))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	bs, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var data createdResponse

	if err := json.Unmarshal(bs, &data); err != nil {
		log.Fatal(err)
	}

	n.jobID = data.ID

	// TODO catch error for safety use nursecall
	if 0 < n.intervalHeartBeat {
		go n.heartbeat()
	}

	return nil
}

type createdResponse struct {
	ID string `json:"id"`
}

func (n *Notifier) update(params map[string]interface{}) error {
	params["id"] = n.jobID
	inputBytes, err := json.Marshal(params)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPut,
		n.EndpointURL+"/update_job",
		strings.NewReader(string(inputBytes)),
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

	bs, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	log.Println(string(bs))

	return nil
}

func (n *Notifier) heartbeat() {
	params := map[string]interface{}{}

	for {
		time.Sleep(time.Second * time.Duration(n.intervalHeartBeat))
		if err := n.update(params); err != nil {
			log.Println(err)
		}
	}
}

func (n *Notifier) Done(exitCode int) error {
	params := map[string]interface{}{
		"execute_status": "success",
		"exit_code":      exitCode,
	}

	if err := n.update(params); err != nil {
		return err
	}
	return nil
}

func (n *Notifier) Error(exitCode int) error {
	params := map[string]interface{}{
		"execute_status": "failed",
		"exit_code":      exitCode,
	}

	if err := n.update(params); err != nil {
		return err
	}
	return nil
}
