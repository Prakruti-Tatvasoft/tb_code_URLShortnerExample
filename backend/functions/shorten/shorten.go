package lib

import (
	"io"
	"strings"

	"github.com/taubyte/go-sdk/database"
	"github.com/taubyte/go-sdk/event"
	"github.com/taubyte/utils/multihash"
)

//go:generate go get github.com/mailru/easyjson
//go:generate go install github.com/mailru/easyjson/...@latest
//go:generate easyjson -omit_empty -all ${GOFILE}

const (
	hostDomain = "qzmgbe9m0.g.tau.link"
	minLen     = 5
)

type BodyRequest struct {
	HostDomain string `json:"base_url"`
	URL        string `json:"url"`
}

type BodyResponse struct {
	URL    string `json:"url"`
	Short  string `json:"short"`
	Exists bool   `json:"exists"`
	Error  string `json:"error"`
}

/* POST /shorten */
//export shorten
func shorten(e event.Event) uint32 {

	bodyResponse := BodyResponse{
		URL:    "",
		Short:  "",
		Exists: false,
		Error:  "",
	}

	h, err := e.HTTP()
	if err != nil {
		bodyResponse.Error = err.Error()
	} else {

		bodyData, err := io.ReadAll(h.Body())
		if err != nil {
			bodyResponse.Error = err.Error()
		} else {

			bodyRequest := &BodyRequest{}
			err = bodyRequest.UnmarshalJSON(bodyData)
			if err != nil {
				bodyResponse.Error = err.Error()
			} else {

				if bodyRequest.HostDomain != hostDomain {
					bodyResponse.Error = "Requested domain is not matched"
				} else {

					urlHashCode := multihash.Hash(bodyRequest.URL)

					urlHashCode = strings.ToLower(urlHashCode[len(urlHashCode)-minLen:])

					db, err := database.New("urls")
					if err != nil {
						bodyResponse.Error = err.Error()
					}
					defer db.Close()

					if bodyResponse.Error == "" {
						_, err = db.Get(urlHashCode)
						if err == nil {
							bodyResponse.Exists = true
						} else {
							err = db.Put(urlHashCode, []byte(bodyRequest.URL))
							if err != nil {
								bodyResponse.Error = err.Error()
							}
						}
					}

					if bodyResponse.Error == "" {
						bodyResponse.URL = "https://" + bodyRequest.HostDomain + "/r?s=" + urlHashCode
						bodyResponse.Short = urlHashCode
					}
				}
			}
		}
	}

	res, _ := bodyResponse.MarshalJSON()

	h.Write(res)
	h.Return(200)

	return 0
}
