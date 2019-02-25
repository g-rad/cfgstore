package cfgstorego

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/certifi/gocertifi"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Config []*KeyValue

type apiClient struct {
	Host          string
	ApiVersion    string
	ClientName    string
	ClientVersion string
	Retries       uint

	HttpCli *http.Client
}

func newApiClient(
	host string,
	apiVersion string,
	clientName string,
	clientVersion string,
	timeoutSeconds int,
	retries uint,
) (*apiClient, error) {
	httpCli, err := newHttpClient(timeoutSeconds)
	if err != nil {
		return nil, err
	}

	cli := &apiClient{
		Host:          sanitizeHost(host),
		ApiVersion:    apiVersion,
		ClientName:    clientName,
		ClientVersion: clientVersion,
		Retries:       retries,
		HttpCli:       httpCli,
	}
	return cli, nil
}

func (cli *apiClient) GetConfig(key string) (Config, error) {
	url := cli.buildRequestUrl(key)

	var response Config

	err := retry.Do(func()error {
		resp, err := cli.getConfig(url, key)
		if err != nil {
			return err
		}
		response = resp
		return nil
	}, retry.Attempts(cli.Retries))

	return response, err
}

func (cli *apiClient) getConfig(url, key string) (Config, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, errors.Wrap(err, "error creating request")
	}

	cli.setHeaders(req)

	resp, err := cli.HttpCli.Do(req)

	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, errors.Wrap(err, "error sending request")
	}

	if resp != nil && resp.StatusCode >= 500 {
		return nil, NewHttpStatusError(resp)
	}

	if resp != nil && resp.StatusCode == 404 {
		return nil, errors.New("cfg store key invalid")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response")
	}

	response := Config{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing response")
	}

	return response, nil
}

func (cli *apiClient) buildRequestUrl(key string) string {
	return fmt.Sprintf("%vapi/%v/config/%v", cli.Host, cli.ApiVersion, key)
}

func (cli *apiClient) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", fmt.Sprintf("%v %v", cli.ClientName, cli.ClientVersion))
}

func newHttpClient(timeoutSeconds int) (*http.Client, error) {

	certPool, err := gocertifi.CACerts()
	if err != nil {
		return nil, errors.Wrap(err, "error building cert pool")
	}

	cli := &http.Client{
		Timeout: time.Second * time.Duration(timeoutSeconds),
		Transport: &http.Transport{
			TLSHandshakeTimeout: time.Duration(timeoutSeconds/2) * time.Second,
			TLSClientConfig:     &tls.Config{RootCAs: certPool},
		},
	}

	return cli, nil
}

func sanitizeHost(host string) string {
	if !strings.HasSuffix(host, "/") {
		return host + "/"
	}
	return host
}