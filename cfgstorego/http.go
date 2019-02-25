package cfgstorego

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

type HttpStatusError struct {
	StatusCode int
	Body       string
}

func (e *HttpStatusError) Error() string { return "server error" }

func (e *HttpStatusError) Format() string {
	return fmt.Sprintf("status: %v\nbody: %v", e.StatusCode, e.Body)
}

func NewHttpStatusError(resp *http.Response) *HttpStatusError {
	body := ""
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		body = errors.Wrap(err, "error reading response body").Error()
	} else {
		body = string(bodyBytes)
	}

	e := &HttpStatusError{
		StatusCode: resp.StatusCode,
		Body:       body,
	}

	return e
}
