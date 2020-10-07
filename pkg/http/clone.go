package http

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
)

func Clone(ctx context.Context, r *http.Request) (err error, r2 *http.Request) {
	r2 = r.Clone(ctx)
	//*r2 = *r

	var b bytes.Buffer
	_, err = b.ReadFrom(r.Body)
	if err != nil {
		return err, r2
	}
	r.Body = ioutil.NopCloser(&b)
	r2.Body = ioutil.NopCloser(bytes.NewReader(b.Bytes()))
	return nil, r2
}
