package sub

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func NewErr(message string, v ...interface{}) error {
	return fmt.Errorf(message, v...)
}

type HttpClient struct {
	Method    string
	Url       string
	Payload   io.Reader
	Auth      Auth
	TlsConfig *tls.Config
	Client    *http.Client
}

func (cl *HttpClient) Do() (*http.Response, error) {
	if cl.Method == "" {
		cl.Method = "GET"
	}
	if cl.TlsConfig == nil {
		cl.TlsConfig = &tls.Config{InsecureSkipVerify: true}
	}
	req, err := http.NewRequest(cl.Method, cl.Url, cl.Payload)
	if err != nil {
		return nil, err
	}
	if cl.Auth.Type == "basic" {
		req.Header.Set("Authorization", BasicAuth(cl.Auth.Username, cl.Auth.Password))
	}
	cl.Client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: cl.TlsConfig,
		},
	}
	return cl.Client.Do(req)
}

func (cl *HttpClient) Close() {
	if cl.Client != nil {
		cl.Client.CloseIdleConnections()
	}
}

type Auth struct {
	Type     string
	Username string
	Password string
}

func BasicAuth(username, password string) string {
	auth := username + ":"
	if password != "" {
		auth += password
	}
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

type Client struct {
	Auth Auth
	Host string
}

func (cl Client) NewRequest(url string) *HttpClient {
	client := &HttpClient{
		Auth: Auth{
			Type:     "basic",
			Username: cl.Auth.Username,
			Password: cl.Auth.Password,
		},
		Url: url,
	}
	return client
}

func (cl Client) GetBody(url string) ([]byte, error) {
	client := cl.NewRequest(url)
	r, err := client.Do()
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		return nil, NewErr(r.Status)
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (cl Client) JSON(client *HttpClient, i, o interface{}) error {
	data, err := json.Marshal(i)
	if err != nil {
		return err
	}
	if Verbose {
		log.Printf("Client.JSON -> %s %s", client.Method, client.Url)
		log.Printf("Client.JSON -> %s", string(data))
	}
	client.Payload = bytes.NewReader(data)
	if r, err := client.Do(); err != nil {
		return err
	} else {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		if Verbose {
			log.Printf("client.JSON <- %s", string(body))
		}
		if r.StatusCode != http.StatusOK {
			return NewErr("%s %s", r.Status, body)
		} else if o != nil {
			if err := json.Unmarshal(body, o); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cl Client) GetJSON(url string, v interface{}) error {
	client := cl.NewRequest(url)
	client.Method = "GET"
	return cl.JSON(client, nil, v)
}

func (cl Client) PostJSON(url string, i, o interface{}) error {
	client := cl.NewRequest(url)
	client.Method = "POST"
	return cl.JSON(client, i, o)
}

func (cl Client) PutJSON(url string, i, o interface{}) error {
	client := cl.NewRequest(url)
	client.Method = "PUT"
	return cl.JSON(client, i, o)
}

func (cl Client) DeleteJSON(url string, i, o interface{}) error {
	client := cl.NewRequest(url)
	client.Method = "DELETE"
	return cl.JSON(client, i, o)
}

type Cmd struct {
}

func (c Cmd) NewHttp(token string) Client {
	values := strings.SplitN(token, ":", 2)
	username := values[0]
	password := values[0]
	if len(values) == 2 {
		password = values[1]
	}
	if len(values) == 1 {
		username = "vrr"
	}
	client := Client{
		Auth: Auth{
			Username: username,
			Password: password,
		},
	}
	return client
}

func (c Cmd) Url(prefix, name string) string {
	return ""
}

func (c Cmd) Tmpl() string {
	return ""
}

func (c Cmd) Out(data interface{}, format string) error {
	return Out(data, format)
}
