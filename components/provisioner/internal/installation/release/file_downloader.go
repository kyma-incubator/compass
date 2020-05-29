package release

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

type httpGetter interface {
	Get(url string) (resp *http.Response, err error)
}

// FileDownloader downloads text files
type FileDownloader struct {
	httpGetter httpGetter
}

func NewFileDownloader(getter httpGetter) *FileDownloader {
	return &FileDownloader{
		httpGetter: getter,
	}
}

// Download downloads text file
func (fd *FileDownloader) Download(url string) (string, error) {
	resp, err := fd.httpGetter.Get(url)
	if err != nil {
		return "", errors.Wrapf(err, "while executing get request on url: %q", url)
	}
	defer util.Close(resp.Body)

	return fd.readResponse(resp)
}

// DownloadOrEmpty downloads text file
// If the response status is 404 return empty string with no error
func (fd *FileDownloader) DownloadOrEmpty(url string) (string, error) {
	resp, err := fd.httpGetter.Get(url)
	if err != nil {
		return "", errors.Wrapf(err, "while executing get request on url: %q", url)
	}
	defer util.Close(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}

	return fd.readResponse(resp)
}

func (fd *FileDownloader) readResponse(resp *http.Response) (string, error) {
	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("received unexpected http status %d", resp.StatusCode)
	}

	reqBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "while reading body")
	}

	return string(reqBody), nil
}
