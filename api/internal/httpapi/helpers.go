package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
)

func jsonNewDecoder(r *http.Request) *json.Decoder {
	return json.NewDecoder(io.LimitReader(r.Body, 2<<20))
}
