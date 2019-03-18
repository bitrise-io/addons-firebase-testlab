package testreportfiller

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/bitrise-io/addons-firebase-testlab/junit"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// TestReportAssetInfo ...
type TestReportAssetInfo struct {
	Filename    string    `json:"filename"`
	Filesize    int       `json:"filesize"`
	Uploaded    bool      `json:"uploaded"`
	DownloadURL string    `json:"download_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// TestReportWithTestSuites ...
type TestReportWithTestSuites struct {
	ID         uuid.UUID             `json:"id"`
	TestSuites []junit.Suite         `json:"test_suites"`
	StepInfo   models.StepInfo       `json:"step_info"`
	TestAssets []TestReportAssetInfo `json:"test_assets"`
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

		stepInfo := models.StepInfo{}
		err = json.Unmarshal([]byte(trr.Step), &stepInfo)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to get step info for test report")
		}

		testReportAssetInfos := []TestReportAssetInfo{}
		for _, tra := range trr.TestReportAssets {
			trai := TestReportAssetInfo{
				Filename:  tra.Filename,
				Filesize:  tra.Filesize,
				Uploaded:  tra.Uploaded,
				CreatedAt: tra.CreatedAt,
			}
			downloadURL, err := fAPI.DownloadURLforPath(tra.PathInBucket())
			if err != nil {
				return nil, errors.Wrap(err, "Failed to get test report asset download URL")
			}
			trai.DownloadURL = downloadURL
			testReportAssetInfos = append(testReportAssetInfos, trai)
		}

		trwts := TestReportWithTestSuites{
			ID:         trr.ID,
			TestSuites: testSuites,
			StepInfo:   stepInfo,
			TestAssets: testReportAssetInfos,
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
