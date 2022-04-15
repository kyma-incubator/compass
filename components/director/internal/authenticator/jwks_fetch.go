package authenticator

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/pkg/errors"
)

/**
Copied from github.com/lestrrat-go/jwx/jwk/jwk.go & modified loading files

The MIT License (MIT)

Copyright (c) 2015 lestrrat

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

// FetchJWK fetches a JWK resource specified by a URL
func FetchJWK(ctx context.Context, urlstring string, options ...jwk.FetchOption) (jwk.Set, error) {
	u, err := url.Parse(urlstring)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse url")
	}

	switch u.Scheme {
	case "http", "https":
		return jwk.Fetch(ctx, urlstring, options...)
	case "file":
		filePath := strings.TrimPrefix(urlstring, "file://")
		f, err := os.Open(filePath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to open jwk file")
		}
		defer func() {
			err := f.Close()
			if err != nil {
				log.C(ctx).WithError(err).Errorf("An error has occurred while closing file: %v", err)
			}
		}()

		buf, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, errors.Wrap(err, "failed read content from jwk file")
		}
		return jwk.Parse(buf)
	}
	return nil, errors.Errorf("invalid url scheme %s", u.Scheme)
}
