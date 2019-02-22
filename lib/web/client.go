package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// InternalPassClient uses "private" backend URIs for interacting with the PASS repository
// It is intended for use on private networks.  Public URIs will be
// converted to private URIs when accessing the repository.
type InternalPassClient struct {
	Requester
	ExternalBaseURI string
	InternalBaseURI string
	Credentials     *Credentials
}

type Credentials struct {
	Username string
	Password string
}

// Requester performs http requests
type Requester interface {
	Do(req *http.Request) (*http.Response, error)
}

// FetchEntity fetches and parses the PASS entity at the given URL to the struct or map
// pointed to by entityPointer
func (c *InternalPassClient) FetchEntity(url string, entityPointer interface{}) error {
	url, err := c.translate(url)
	if err != nil {
		return errors.Wrapf(err, "error translating url")
	}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return errors.Wrapf(err, "could not build http request to %s", url)
	}

	if c.Credentials != nil {
		request.SetBasicAuth(c.Credentials.Username, c.Credentials.Password)
	}
	request.Header.Set(headerUserAgent, "pass-schema-service")
	request.Header.Set(headerAccept, mediaJSONTypes)

	resp, err := c.Do(request)
	if err != nil {
		return errors.Wrapf(err, "error connecting to %s", url)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(entityPointer)
	if err != nil {
		return errors.Wrapf(err, "could not decode resource JSON")
	}

	return nil
}

func (c *InternalPassClient) translate(uri string) (string, error) {
	if !strings.HasPrefix(uri, c.ExternalBaseURI) &&
		!strings.HasPrefix(uri, c.InternalBaseURI) {
		return uri, fmt.Errorf(`uri "%s" must start with internal or external baseuri"`, uri)
	}
	return strings.Replace(uri, c.ExternalBaseURI, c.InternalBaseURI, 1), nil
}
