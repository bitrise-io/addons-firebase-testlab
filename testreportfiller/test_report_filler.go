package testreportfiller

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/bitrise-io/addons-firebase-testlab/bitrise"
	"github.com/bitrise-io/addons-firebase-testlab/junit"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/gobuffalo/uuid"
	junitmodels "github.com/joshdk/go-junit"
	"github.com/pkg/errors"
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
	TestSuites []junitmodels.Suite   `json:"test_suites"`
	StepInfo   models.StepInfo       `json:"step_info"`
	TestAssets []TestReportAssetInfo `json:"test_assets"`
}

// CheckStyleResult represents checkstyle XML result.
// <?xml version="1.0" encoding="utf-8"?><checkstyle version="4.3"><file ...></file>...</checkstyle>
//
// References:
//   - http://checkstyle.sourceforge.net/
//   - http://eslint.org/docs/user-guide/formatters/#checkstyle
type CheckStyleResult struct {
	XMLName xml.Name          `xml:"checkstyle"`
	Version string            `xml:"version,attr"`
	Files   []*CheckStyleFile `xml:"file,omitempty"`
}

// CheckStyleFile represents <file name="fname"><error ... />...</file>
type CheckStyleFile struct {
	Name   string             `xml:"name,attr"`
	Errors []*CheckStyleError `xml:"error"`
}

// CheckStyleError represents <error line="1" column="10" severity="error" message="msg" source="src" />
type CheckStyleError struct {
	Column   int    `xml:"column,attr,omitempty"`
	Line     int    `xml:"line,attr"`
	Message  string `xml:"message,attr"`
	Severity string `xml:"severity,attr,omitempty"`
	Source   string `xml:"source,attr,omitempty"`
}

// Filler ...
type Filler struct{}

// DownloadURLCreator ...
type DownloadURLCreator interface {
	DownloadURLforPath(string) (string, error)
}

// FillMore ...
func (f *Filler) FillMore(testReportRecords []models.TestReport, fAPI DownloadURLCreator, junitParser junit.Parser, httpClient *http.Client, status string) ([]TestReportWithTestSuites, error) {
	testReportsWithTestSuites := []TestReportWithTestSuites{}

	for _, trr := range testReportRecords {
		trwts, err := f.FillOne(trr, fAPI, junitParser, httpClient, status)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to fill test report")
		}

		testReportsWithTestSuites = append(testReportsWithTestSuites, trwts)
	}
	return testReportsWithTestSuites, nil
}

// FillOne ...
func (f *Filler) FillOne(trr models.TestReport, fAPI DownloadURLCreator, junitParser junit.Parser, httpClient *http.Client, status string) (TestReportWithTestSuites, error) {
	downloadURL, err := fAPI.DownloadURLforPath(trr.PathInBucket())
	xml, err := getContent(downloadURL, httpClient)
	if err != nil {
		return TestReportWithTestSuites{}, errors.Wrap(err, "Failed to get test report XML")
	}

	testSuites, err := junitParser.Parse(xml)
	if err != nil {
		return TestReportWithTestSuites{}, errors.Wrap(err, "Failed to parse test report XML")
	}

	if status != "" {
		testSuites = filterTestSuitesByStatus(testSuites, status)
	}

	stepInfo := models.StepInfo{}
	err = json.Unmarshal([]byte(trr.Step), &stepInfo)
	if err != nil {
		return TestReportWithTestSuites{}, errors.Wrap(err, "Failed to get step info for test report")
	}

	testReportAssetInfos := []TestReportAssetInfo{}
	for _, tra := range trr.TestReportAssets {
		trai := TestReportAssetInfo{
			Filename:  tra.Filename,
			Filesize:  tra.Filesize,
			Uploaded:  tra.Uploaded,
			CreatedAt: tra.CreatedAt,
		}
		tra.TestReport = trr
		downloadURL, err := fAPI.DownloadURLforPath(tra.PathInBucket())
		if err != nil {
			return TestReportWithTestSuites{}, errors.Wrap(err, "Failed to get test report asset download URL")
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
	return trwts, nil
}

// Annotate ...
func (f *Filler) Annotate(trr models.TestReport, fAPI DownloadURLCreator, httpClient *http.Client) ([]bitrise.Annotation, error) {
	downloadURL, err := fAPI.DownloadURLforPath(trr.PathInBucket())
	dataBytes, err := getContent(downloadURL, httpClient)
	if err != nil {
		return []bitrise.Annotation{}, errors.Wrap(err, "Failed to get test report XML")
	}

	r := bytes.NewReader(dataBytes)

	cs := new(CheckStyleResult)
	if err := xml.NewDecoder(r).Decode(cs); err != nil {
		return nil, errors.Wrap(err, "failed to parse XML")
	}
	var annotations []bitrise.Annotation
	for _, file := range cs.Files {
		for _, cerr := range file.Errors {
			annotations = append(annotations, bitrise.Annotation{
				Path:            file.Name,
				StartLine:       cerr.Line,
				EndLine:         cerr.Line,
				StartColumn:     cerr.Column,
				EndColumn:       cerr.Column,
				AnnotationLevel: severityToLevel(cerr.Severity),
				Message:         cerr.Message,
			})
		}
	}

	return annotations, nil
}

func severityToLevel(sev string) string {
	levelMap := map[string]string{
		"error":   "failure",
		"failure": "failure",
		"warning": "warning",
		"notice":  "notice",
	}

	if levelMap[sev] != "" {
		return levelMap[sev]
	}

	return "failure"
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

func filterTestSuitesByStatus(testSuites []junitmodels.Suite, status string) []junitmodels.Suite {
	filteredSuites := []junitmodels.Suite{}
	filteredTests := []junitmodels.Test{}

	for _, suite := range testSuites {
		filteredTests = []junitmodels.Test{}
		for _, test := range suite.Tests {
			if statusMatch(string(test.Status), status) {
				filteredTests = append(filteredTests, test)
			}
		}

		if len(filteredTests) > 0 {
			suite.Tests = filteredTests
			filteredSuites = append(filteredSuites, suite)
		}
	}

	return filteredSuites
}

func statusMatch(testStatus string, expected string) bool {
	if testStatus == expected {
		return true
	}

	if testStatus == "error" && expected == "failed" {
		return true
	}

	return false
}
