package testreportfiller_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/bitrise-io/addons-firebase-testlab/junit"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/bitrise-io/addons-firebase-testlab/testreportfiller"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

type TestFAPI struct{}

func (f *TestFAPI) DownloadURLforPath(string) (string, error) {
	return "http://dont.call.me.pls", nil
}

// RoundTripFunc ...
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip ...
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient ...
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func Test_TestReportFiller_Fill(t *testing.T) {
	id1, err := uuid.FromString("aaaaaaaa-18d6-11e9-ab14-d663bd873d93")
	if err != nil {
		t.Fatal(err)
	}

	id2, err := uuid.FromString("bbbbbbbb-18d6-11e9-ab14-d663bd873d93")
	if err != nil {
		t.Fatal(err)
	}

	trs := []models.TestReport{
		models.TestReport{
			ID:        id1,
			Filename:  "test1.xml",
			BuildSlug: "buildslug1",
		},
		models.TestReport{
			ID:        id2,
			Filename:  "test1.xml",
			BuildSlug: "buildslug1",
		},
	}

	testCases := []struct {
		name                  string
		xml                   string
		statusFromXMLDownload int
		expResp               []testreportfiller.TestReportWithTestSuites
		expErr                string
	}{
		{

			name: "when the test report files are found and valid",
			xml: `
	    <?xml version="1.0" encoding="UTF-8"?>
	    <testsuites>
			<testsuite>
			</testsuite>
	    </testsuites>
			`,
			statusFromXMLDownload: 200,
			expResp: []testreportfiller.TestReportWithTestSuites{
				testreportfiller.TestReportWithTestSuites{
					id1,
					[]junit.Suite{
						junit.Suite{},
					},
				},
				testreportfiller.TestReportWithTestSuites{
					id2,
					[]junit.Suite{
						junit.Suite{},
					},
				},
			},
			expErr: "",
		},
		{
			name: "when the test report file is not found",
			xml:  "",
			statusFromXMLDownload: 404,
			expErr:                "Failed to get test report XML",
		},
		{
			name: "when the test report file is not valid",
			xml:  "<xml?>",
			statusFromXMLDownload: 200,
			expErr:                "Failed to parse test report XML",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filler := testreportfiller.Filler{}
			httpClient := NewTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: tc.statusFromXMLDownload,
					Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(tc.xml))),
				}
			})
			got, err := filler.Fill(trs, &TestFAPI{}, &junit.Client{}, httpClient)

			if len(tc.expErr) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expResp, got)
			}
		})
	}
}
