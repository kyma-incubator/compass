package paging

import (
	"fmt"

	"github.com/pkg/errors"
)

//FetchPageFunc is responsible for executing an HTTP request to the url which the PageIterator provided.
//If the results need to be further processed, they should be saved in a closure.
//The function returns the number of the results fetched and an error if any occurred during the request execution.
type FetchPageFunc func(string) (uint, error)

//PageIterator is responsible for executing multiple HTTP requests until all results
//of a pageable API are fetched
type PageIterator struct {
	baseURL               string
	pageSkipFormat        string
	pageSizeFormat        string
	additionalQueryParams []string

	do       FetchPageFunc
	skip     uint
	pageSize uint

	paramsCount uint
	nextURL     string
}

//NewPageIterator constructs a new page iterator with the given args
func NewPageIterator(baseURL, skipFormat, sizeFormat string, additionalQueryParams []string, pageSize uint, do FetchPageFunc) PageIterator {
	return PageIterator{
		baseURL:               baseURL,
		pageSkipFormat:        skipFormat,
		pageSizeFormat:        sizeFormat,
		additionalQueryParams: additionalQueryParams,
		skip:                  0,
		pageSize:              pageSize,
		do:                    do,
	}
}

//Next fetches the next page of the PageIterator. It returns true if there possibly are more pages that can eb fetched.
//In order to get all the results that the PageIterator can fetch Next should be called until it returns false and no error.
//Once next returns false with no error it should not be called anymore. If it does get called the behaviour is undefined.
func (p *PageIterator) Next() (bool, error) {
	p.buildNextURL()
	count, err := p.do(p.nextURL)
	if err != nil {
		return false, errors.Wrapf(err, "while fetching next page: ")
	}
	if count < p.pageSize {
		return false, nil
	}
	p.skip += p.pageSize
	return true, nil
}

//FetchAll fetches all the pages (calls Next method) that the PageIterator can fetch.
func (p *PageIterator) FetchAll() error {
	for again, err := p.Next(); again; again, err = p.Next() {
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PageIterator) buildNextURL() {
	p.resetURL()
	p.setQueryParam(fmt.Sprintf(p.pageSkipFormat, p.skip))
	p.setQueryParam(fmt.Sprintf(p.pageSizeFormat, p.pageSize))
	for _, param := range p.additionalQueryParams {
		p.setQueryParam(param)
	}
}

func (p *PageIterator) resetURL() {
	p.nextURL = p.baseURL
	p.paramsCount = 0
}

func (p *PageIterator) setQueryParam(q string) {
	if p.paramsCount == 0 {
		p.nextURL = p.nextURL + "?" + q
	} else {
		p.nextURL = p.nextURL + "&" + q
	}
	p.paramsCount++
}
