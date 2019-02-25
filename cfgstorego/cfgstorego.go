package cfgstorego

import (
	"encoding/base64"
	"github.com/pkg/errors"
	"os"
	"strings"
)

const EnvKey = "cfgstorekey"
const ClientName = "cfgstorego"
const Version = "0"
const ApiVersion = "v1"

var cli *apiClient

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func init() {
	Load()
}

func Load() {
	key := os.Getenv(EnvKey)

	if key == "" {
		panic(errors.New("missing " + EnvKey))
	}

	cfg, err := LoadConfig(key)

	if err != nil {
		panic(err)
	}

	for _, kv := range cfg {
		if os.Getenv(kv.Key) == "" {
			if err := os.Setenv(kv.Key, kv.Value); err != nil {
				panic(errors.Wrap(err, "error setting environment variable"))
			}
		}
	}
}

type Options struct {
	ClientName     string
	ClientVersion  string
	TimeoutSeconds int
	Retries        uint
	ApiHost        string
	ApiVersion     string
}

func LoadConfig(key string, options ...func(*Options)) (Config, error) {

	host, err := parseKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing "+EnvKey)
	}

	o := &Options{
		ClientName:     ClientName,
		ClientVersion:  "",
		TimeoutSeconds: 10,
		Retries:        3,
		ApiHost:        host,
		ApiVersion:     ApiVersion,
	}

	for _, option := range options {
		option(o)
	}

	if cli == nil {
		cli, err = newApiClient(o.ApiHost, o.ApiVersion, o.ClientName, o.ClientVersion, o.TimeoutSeconds, o.Retries)
		if err != nil {
			panic(err)
		}
	}

	return cli.GetConfig(key)
}

func parseKey(key string) (string, error) {
	split := strings.Split(key, "-")
	hostBytes, err := base64.StdEncoding.DecodeString(split[1])
	if err != nil {
		return "", errors.Wrap(err, "error decoding host")
	}
	return string(hostBytes), nil
}