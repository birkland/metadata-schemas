package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/OA-PASS/metadata-schemas/lib/jsonschema"
	"github.com/OA-PASS/metadata-schemas/lib/web"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var serveOpts = struct {
	publicBaseURI  string
	privateBaseURI string
	username       string
	passwd         string
	port           int
}{}

var serve cli.Command = cli.Command{
	Name:  "serve",
	Usage: "Sereve the PASS schema service over http",
	Description: `

		An optional list of files or directories may be provided, which will be
		examined for the presence of schema files which will be used for static lookups
		`,
	ArgsUsage: "[ file | dir ] ...",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "external, e",
			Usage:       "External (public) PASS baseuri",
			EnvVar:      "PASS_EXTERNAL_FEDORA_BASEURL",
			Destination: &serveOpts.publicBaseURI,
		},
		cli.StringFlag{
			Name:        "internal, i",
			Usage:       "Internal (private) PASS baseuri",
			EnvVar:      "PASS_FEDORA_BASEURL",
			Destination: &serveOpts.privateBaseURI,
		},
		cli.StringFlag{
			Name:        "username, u",
			Usage:       "Username for basic auth to Fedora",
			EnvVar:      "PASS_FEDORA_USER",
			Destination: &serveOpts.username,
		},
		cli.StringFlag{
			Name:        "password, p",
			Usage:       "Password for basic auth to Fedora",
			EnvVar:      "PASS_FEDORA_PASSWORD",
			Destination: &serveOpts.passwd,
		},
		cli.IntFlag{
			Name:        "port",
			Usage:       "Port for the schema service http endpoint",
			EnvVar:      "SCHEMA_SERVICE_PORT",
			Destination: &serveOpts.port,
		},
	},

	Action: func(c *cli.Context) error {
		return serveAction(c.Args())
	},
}

func serveAction(paths []string) error {
	staticSchemas, err := jsonschema.Load(paths)
	if err != nil {
		return errors.Wrapf(err, "Error loading schemas")
	}

	var credentials *web.Credentials
	if serveOpts.username != "" {
		credentials = &web.Credentials{
			Username: serveOpts.username,
			Password: serveOpts.passwd,
		}
	}

	schemaService := &web.SchemaService{
		PassClient: &web.InternalPassClient{
			Requester:       &http.Client{},
			ExternalBaseURI: serveOpts.publicBaseURI,
			InternalBaseURI: serveOpts.privateBaseURI,
			Credentials:     credentials,
		},
		SchemaFetcher: staticSchemas,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlePost(schemaService, w, r)
		case http.MethodHead:
			commonHeaders(w)
		case http.MethodGet:
			handleGet(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", serveOpts.port))
	if err != nil {
		return err
	}

	serveOpts.port = listener.Addr().(*net.TCPAddr).Port
	log.Printf("Listening on port %d", serveOpts.port)

	return http.Serve(listener, nil)
}

func handlePost(schemaService *web.SchemaService, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	commonHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	var req web.Request
	var err error

	switch r.Header.Get("Content-Type") {
	case "text/plain":
		err = req.ReadText(r.Body)
	default:
		err = req.ReadJSON(r.Body)
	}
	if err != nil {
		log.Printf("Error reading request body: %s", err)
		http.Error(w, fmt.Sprintf("Malformed request: %s", err), http.StatusBadRequest)
		return
	}

	schemas, err := schemaService.Schemas(&req)
	if err != nil {
		log.Printf("Error processing schemas: %s", err)
		http.Error(w, "server error!", http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(schemas)
	if err != nil {
		log.Printf("error encoding JSON response: %s", err)
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	commonHeaders(w)
	w.Header().Set("Content-Type", "text/html")
	_, err := w.Write([]byte(`
	<html>
	<body>
	<p>
	See the PASS schema service 
	<a href="https://docs.google.com/document/d/1sLWGZR4kCvQVGv-TA5x8ny-AxL3ChBYNeFYW1eACsDw/edit">documentation</a>
	</p>
	</body>
	</html>
	`))
	if err != nil {
		log.Printf("Error writing GET response: %s\n", err)
	}
}

func commonHeaders(w http.ResponseWriter) {
	w.Header().Set("Accept-Post", "application/json, text/plain")
	w.Header().Set("Server", "PASS schema service")
}
