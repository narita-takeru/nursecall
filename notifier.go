package nursecall

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Notifier struct {
	debug bool

	intervalHeartBeat int

	httpClient  *http.Client
	endpointURL string

	jobID    string
	taskName string

	stdOut string
	stdErr string
}

const (
	defaultEndpointURL = "https://api.nursecall.run/jobs"
)

func getHeartBeatInterval() int {
	interval := os.Getenv("NURSECALL_HEARTBEAT_INTERVAL_SEC")
	if interval == "" {
		return 0
	}
	value, err := strconv.Atoi(interval)
	if err != nil {
		return 0
	}

	return value
}

func NewNotifier() Notifier {
	n := Notifier{
		debug:             os.Getenv("NURSECALL_DEBUG") == "TRUE",
		intervalHeartBeat: getHeartBeatInterval(),
		httpClient:        &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}},
		endpointURL:       defaultEndpointURL,
		taskName:          os.Getenv("NURSECALL_TASK_NAME"),
	}

	return n
}

func (n *Notifier) Start(cmdStr string) error {
	input := map[string]interface{}{
		"call_token": os.Getenv("NURSECALL_CALL_TOKEN"),
		"path":       cmdStr,
		"task_name":  n.taskName,
	}

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return err
	}

	res, err := n.httpClient.Post(n.endpointURL, "application/json", bytes.NewBuffer(inputBytes))
	if err != nil {
		return err
	}

	defer res.Body.Close()

	bs, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if n.debug {
		log.Println(string(bs))
	}

	var data createdResponse

	if err := json.Unmarshal(bs, &data); err != nil {
		if n.debug {
			log.Println(err)
		}

		// nursecall apiと連携が取れなかったら、普通に実行する
		return nil
	} else {
		n.jobID = data.ID
	}

	// TODO catch error for safety use nursecall
	if 0 < n.intervalHeartBeat {
		go n.heartbeat()
	}

	return nil
}

type createdResponse struct {
	ID string `json:"id"`
}

func (n *Notifier) update(path string, params map[string]interface{}) error {
	params["id"] = n.jobID
	inputBytes, err := json.Marshal(params)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPut,
		n.endpointURL+path,
		strings.NewReader(string(inputBytes)),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if n.debug {
		bs, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		log.Println(string(bs))
	}

	return nil
}

func (n *Notifier) setStdout(stdout string) {
	n.stdOut = stdout
}

func (n *Notifier) setStderr(stderr string) {
	n.stdErr = stderr
}

func (n *Notifier) heartbeat() {
	for {
		time.Sleep(time.Second * time.Duration(n.intervalHeartBeat))
		params := map[string]interface{}{}

		if err := n.update("/update_heartbeat", params); err != nil {
			log.Println(err)
		}
	}
}

func (n *Notifier) Done(exitCode int) error {
	if n.jobID == "" {
		return nil
	}

	params := map[string]interface{}{
		"execute_status": "success",
		"exit_code":      exitCode,
		"stdout":         n.stdOut,
		"stderr":         n.stdErr,
	}

	if err := n.update("/update_job", params); err != nil {
		return err
	}
	return nil
}

func (n *Notifier) Error(exitCode int) error {
	if n.jobID == "" {
		return nil
	}

	params := map[string]interface{}{
		"execute_status": "failed",
		"exit_code":      exitCode,
		"stdout":         n.stdOut,
		"stderr":         n.stdErr,
	}

	if err := n.update("/update_job", params); err != nil {
		return err
	}
	return nil
}
