package testreportfiller

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bitrise-io/addons-firebase-testlab/junit"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

// TestReportWithTestSuites ...
type TestReportWithTestSuites struct {
	ID         uuid.UUID     `json:"id"`
	TestSuites []junit.Suite `json:"test_suites"`
}

// Filler ...
type Filler struct{}

// DownloadURLCreator ...
type DownloadURLCreator interface {
	DownloadURLforPath(string) (string, error)
}

// Fill ...
func (f *Filler) Fill(testReportRecords []models.TestReport, fAPI DownloadURLCreator, junitParser junit.Parser, httpClient *http.Client) ([]TestReportWithTestSuites, error) {
	testReportsWithTestSuites := []TestReportWithTestSuites{}

	for _, trr := range testReportRecords {
		downloadURL, err := fAPI.DownloadURLforPath(trr.PathInBucket())
		xml, err := getContent(downloadURL, httpClient)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to get test report XML")
		}

		testSuites, err := junitParser.Parse(xml)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse test report XML")
		}

		trwts := TestReportWithTestSuites{
			ID:         trr.ID,
			TestSuites: testSuites,
		}

		testReportsWithTestSuites = append(testReportsWithTestSuites, trwts)
	}
	return testReportsWithTestSuites, nil
}

func getContent(url string, httpClient *http.Client) ([]byte, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "GET request failed")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Resp body close failed: %+v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Non-200 status code was returned")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Reading body failed")
	}

	return data, nil
}
