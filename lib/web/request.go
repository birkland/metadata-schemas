package web

import (
	"bufio"
	"encoding/json"
	"io"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// Request is a schema service request, logically containing a list of URLs
type Request struct {
	Resources []string
}

// Read parses a request body containing a single list of URIs,
// e.g.
//    ["http://example.org/one", "http://example.org/two"]
func (r *Request) ReadJSON(stream io.Reader) error {

	var given []string

	err := json.NewDecoder(stream).Decode(&given)
	if err != nil {
		return errors.Wrap(err, "could not parse json input")
	}

	for _, addr := range given {
		_, err := url.Parse(addr)
		if err != nil {
			return errors.Wrapf(err, `"%s" is not a URL`, addr)
		}

		r.Resources = append(r.Resources, addr)
	}

	return nil
}

// ReadText parses a body containing URIs separated by newlines
func (r *Request) ReadText(stream io.Reader) error {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		uriText := strings.Trim(scanner.Text(), " 	")
		if uriText == "" {
			continue
		}
		_, err := url.Parse(uriText)
		if err != nil {
			return errors.Wrapf(err, `"%s" is not a URL`, scanner.Text())
		}

		r.Resources = append(r.Resources, uriText)
	}

	if err := scanner.Err(); err != nil {
		return errors.Wrapf(err, "error parsing ")
	}

	return nil
}
