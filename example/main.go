package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/TBD54566975/ssi-sdk/did"
)

type Options struct {
	Address string `json:"api"`
	Debug   bool   `json:"debug"`
}

func parseOptions() *Options {
	var opts Options
	flag.StringVar(&opts.Address, "addr", "0.0.0.0:8080", "address")
	flag.BoolVar(&opts.Debug, "debug", false, "enable debug mode")
	flag.Parse()
	return &opts
}

type ServiceDefinition struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	ID       string   `json:"id"`
	Profiles []string `json:"profiles,omitempty"`
}

type TRServices struct {
	Services []ServiceDefinition `json:"services"`
}

type Signature struct {
}

type Container struct {
	Context        []string                     `json:"@context,omitempty"`
	Source         string                       `json:"source,omitempty"`
	Target         string                       `json:"target,omitempty"`
	AudienceTarget string                       `json:"audienceTarget,omitempty"`
	ThreadNumber   int                          `json:"thread,omitempty"`
	Parent         map[string]map[string]string `json:"parents,omitempty"`
	Children       map[string]map[string]string `json:"children,omitempty"`
	Received       string                       `json:"received, omitempty"`
	Expiry         string                       `json:"expiry, omitempty"`
	Path           string                       `json:"path, omitempty"`
	Scope          string                       `json:"scope, omitempty"`
	Data           string                       `json:"data"`
	Signature      Signature                    `json:"signature,omitempty"`
}

func NewContainer(source, target string) *Container {
	return &Container{
		Context: []string{
			"https://gist.githubusercontent.com/andorsk/2d8bea42e288a9145437a194f45edad0/raw/b74b6c8156518884f5b811486fbee435a54907c1/service-override.jsonld",
		},
		Source:   source,
		Target:   target,
		Received: time.Now().Format(time.RFC3339),
		Expiry:   time.Now().Add(time.Minute * 15).Format(time.RFC3339),
	}
}

type Service struct {
	ID              string   `json:"id" validate:"required"`
	Type            string   `json:"type" validate:"required"`
	ServiceEndpoint any      `json:"serviceEndpoint" validate:"required"`
	RoutingKeys     []string `json:"routingKeys,omitempty"`
	Profiles        []string `json:"profiles"`
	Accept          []string `json:"accept,omitempty"`
}

type Document struct {
	Context any `json:"@context,omitempty"`
	// As per https://www.w3.org/TR/did-core/#did-subject intermediate representations of DID Documents do not
	// require an ID property. The provided test vectors demonstrate IRs. As such, the property is optional.
	ID                   string                      `json:"id,omitempty"`
	Controller           string                      `json:"controller,omitempty"`
	AlsoKnownAs          string                      `json:"alsoKnownAs,omitempty"`
	VerificationMethod   []did.VerificationMethod    `json:"verificationMethod,omitempty" validate:"dive"`
	Authentication       []did.VerificationMethodSet `json:"authentication,omitempty" validate:"dive"`
	AssertionMethod      []did.VerificationMethodSet `json:"assertionMethod,omitempty" validate:"dive"`
	KeyAgreement         []did.VerificationMethodSet `json:"keyAgreement,omitempty" validate:"dive"`
	CapabilityInvocation []did.VerificationMethodSet `json:"capabilityInvocation,omitempty" validate:"dive"`
	CapabilityDelegation []did.VerificationMethodSet `json:"capabilityDelegation,omitempty" validate:"dive"`
	Services             []Service                   `json:"service,omitempty" validate:"dive"`
}

func prepareDIDDocument(opts Options, def ServiceDefinition) error {
	doc := Document{
		Context: []string{
			"https://gist.githubusercontent.com/andorsk/2d8bea42e288a9145437a194f45edad0/raw/b74b6c8156518884f5b811486fbee435a54907c1/service-override.jsonld",
		},
		ID: fmt.Sprintf("did:trustregistry:%s", def.ID),
		Services: []Service{
			{
				ID:              fmt.Sprintf("did:trustregistry:%s", def.ID),
				Type:            "TrustRegistryService",
				ServiceEndpoint: fmt.Sprintf("http://%s/%s", opts.Address, def.ID),
				Accept: []string{
					"application/json",
				},
				Profiles: def.Profiles,
			},
		},
	}

	dat, err := json.MarshalIndent(doc, " ", "   ")
	if err != nil {
		return err
	}
	fmt.Println(string(dat))
	return nil
}

func main() {

	opts := parseOptions()
	services := TRServices{
		Services: []ServiceDefinition{
			{
				Name:     "Credential Trust Establishment",
				ID:       "1",
				File:     "data/registries/credential-trust-establishment.json",
				Profiles: []string{"data/profiles/credential-trust-establishment.json"},
			},
			{
				Name:     "Simple DID Registry",
				ID:       "2",
				File:     "data/registries/tbd.json",
				Profiles: []string{"data/profiles/did-list.json"},
			},
		},
	}

	wg := sync.WaitGroup{}
	for _, service := range services.Services {
		wg.Add(1)
		go func(def ServiceDefinition, opts Options) {
			defer wg.Done()
			data, err := ioutil.ReadFile(def.File)
			if err != nil {
				log.Fatal(err)
			}
			if err := prepareDIDDocument(opts, def); err != nil {
				log.Fatal(err)
			}

			http.HandleFunc(fmt.Sprintf("/%s", def.ID), func(w http.ResponseWriter, r *http.Request) {
				container := NewContainer(service.ID, "did:targetid:1")
				container.Data = string(base64.RawStdEncoding.EncodeToString(data))
				dd, err := json.MarshalIndent(container, " ", "   ")
				if err != nil {
					fmt.Fprintf(w, "error: %s", err)
					return
				}
				fmt.Fprint(w, string(dd))
			})

			// FIXME: same profile against multiple services will collide.
			for _, p := range def.Profiles {
				pp := filepath.Base(p)
				http.HandleFunc(fmt.Sprintf("/profile/%s", pp), func(w http.ResponseWriter, r *http.Request) {
					data, err := ioutil.ReadFile(p)
					if err != nil {
						fmt.Fprintf(w, "error: %s", err)
						return
					}
					fmt.Fprint(w, string(data))
				})
			}

		}(service, *opts)
	}
	wg.Wait()

	// Prepare did document
	log.Fatal(http.ListenAndServe(opts.Address, nil))

}
