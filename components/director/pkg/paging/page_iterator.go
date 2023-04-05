package paging

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const queryPairFormat = "%s=%s"

// FetchPageFunc is responsible for executing an HTTP request to the url which the PageIterator provided.
// If the results need to be further processed, they should be saved in a closure.
// The function returns the number of the results fetched and an error if any occurred during the request execution.
type FetchPageFunc func(string) (uint64, error)

// PageIterator is responsible for executing multiple HTTP requests until all results
// of a pageable API are fetched
type PageIterator struct {
	baseURL               string
	pageSkipParam         string
	pageSizeParam         string
	additionalQueryParams map[string]string

	do       FetchPageFunc
	skip     uint64
	pageSize uint64

	nextURL     string
	hasNext     bool
	paramsCount uint64
}

// NewPageIterator constructs a new page iterator with the given args
func NewPageIterator(baseURL, skipParam, sizeParam string, additionalQueryParams map[string]string, pageSize uint64, do FetchPageFunc) PageIterator {
	return PageIterator{
		baseURL:               baseURL,
		pageSkipParam:         skipParam,
		pageSizeParam:         sizeParam,
		additionalQueryParams: additionalQueryParams,
		skip:                  0,
		pageSize:              pageSize,
		do:                    do,
		hasNext:               true,
	}
}

// Next fetches the next page of the PageIterator. In order to get all the results that the PageIterator can fetch
// Next should be called until it returns false and no error.
// Once Next returns false and no error, Next should not be called anymore because it will do nothing.
func (p *PageIterator) Next() (bool, error) {
	time.Sleep(time.Second * 2)
	if !p.hasNext {
		return false, nil
	}
	p.buildNextURL()
	count, err := p.do(p.nextURL)
	if err != nil {
		return p.hasNext, errors.Wrapf(err, "while fetching next page: ")
	}

	if count == p.pageSize {
		p.skip += p.pageSize
	} else {
		p.hasNext = false
	}
	return p.hasNext, nil
}

// FetchAll fetches all the pages (calls Next method) that the PageIterator can fetch.
func (p *PageIterator) FetchAll() (err error) {
	var hasNext bool
	for hasNext, err = p.Next(); hasNext; hasNext, err = p.Next() {
		if err != nil {
			return err
		}
	}
	return err
}

func (p *PageIterator) buildNextURL() {
	p.resetURL()
	p.setQueryParam(p.pageSkipParam, strconv.FormatUint(p.skip, 10))
	p.setQueryParam(p.pageSizeParam, strconv.FormatUint(p.pageSize, 10))
	for k, v := range p.additionalQueryParams {
		p.setQueryParam(k, v)
	}
}

func (p *PageIterator) resetURL() {
	p.nextURL = p.baseURL
	p.paramsCount = 0
}

// setQueryParam is needed because the builtin functions of net/url will encode everything
// that is in the query param section which is not what we desire. We only want the values to be
// encoded.
func (p *PageIterator) setQueryParam(key, value string) {
	if p.paramsCount == 0 {
		p.nextURL = p.nextURL + "?" + fmt.Sprintf(queryPairFormat, key, url.QueryEscape(value))
	} else {
		p.nextURL = p.nextURL + "&" + fmt.Sprintf(queryPairFormat, key, url.QueryEscape(value))
	}
	p.paramsCount++
}
