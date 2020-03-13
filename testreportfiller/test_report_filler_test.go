package testreportfiller_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/bitrise-io/addons-firebase-testlab/junit"
	"github.com/bitrise-io/addons-firebase-testlab/models"
	"github.com/bitrise-io/addons-firebase-testlab/testreportfiller"
	"github.com/gobuffalo/uuid"
	junitmodels "github.com/joshdk/go-junit"
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

func Test_TestReportFiller_FillMore(t *testing.T) {
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
			Step:      json.RawMessage(`{"id":"an-awesome-step"}`),
			TestReportAssets: []models.TestReportAsset{
				models.TestReportAsset{
					Filename: "my-important-asset",
					Filesize: 121,
				},
				models.TestReportAsset{
					Filename: "another-important-asset",
					Filesize: 534,
				},
			},
		},
		models.TestReport{
			ID:        id2,
			Filename:  "test1.xml",
			BuildSlug: "buildslug1",
			Step:      json.RawMessage(`{"version":"1.0"}`),
		},
	}

	testCases := []struct {
		name                  string
		xml                   string
		statusFilter          string
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
					ID: id1,
					TestSuites: []junitmodels.Suite{
						junitmodels.Suite{},
					},
					StepInfo: models.StepInfo{ID: "an-awesome-step"},
					TestAssets: []testreportfiller.TestReportAssetInfo{
						testreportfiller.TestReportAssetInfo{
							Filename:    "my-important-asset",
							Filesize:    121,
							DownloadURL: "http://dont.call.me.pls",
						},
						testreportfiller.TestReportAssetInfo{
							Filename:    "another-important-asset",
							Filesize:    534,
							DownloadURL: "http://dont.call.me.pls",
						},
					},
				},
				testreportfiller.TestReportWithTestSuites{
					ID: id2,
					TestSuites: []junitmodels.Suite{
						junitmodels.Suite{},
					},
					StepInfo:   models.StepInfo{Version: "1.0"},
					TestAssets: []testreportfiller.TestReportAssetInfo{},
				},
			},
			expErr: "",
		},
		{
			name: "when filtering is on",
			xml: `
	    <?xml version="1.0" encoding="UTF-8"?>
			  <testsuites>
				  <testsuite>
			      <testcase name="successful test"></testcase>
            <testcase name="failing test">
						  <failure/>
						</testcase>
			      <testcase name="skipped test">
						  <skipped />
							</testcase>
						<testcase name="erroneous test">
						  <error />
            </testcase>
			    </testsuite>
	    </testsuites>
			`,
			statusFilter:          "failed",
			statusFromXMLDownload: 200,
			expResp: []testreportfiller.TestReportWithTestSuites{
				testreportfiller.TestReportWithTestSuites{
					ID: id1,
					TestSuites: []junitmodels.Suite{
						junitmodels.Suite{
							Totals: junitmodels.Totals{
								Tests:   4,
								Passed:  1,
								Skipped: 1,
								Failed:  1,
								Error:   1,
							},
							Tests: []junitmodels.Test{
								junitmodels.Test{
									Name:   "failing test",
									Status: "failed",
									Error:  junitmodels.Error{},
									Properties: map[string]string{
										"name": "failing test",
									},
								},
								junitmodels.Test{
									Name:   "erroneous test",
									Status: "error",
									Error:  junitmodels.Error{},
									Properties: map[string]string{
										"name": "erroneous test",
									},
								},
							},
						},
					},
					StepInfo: models.StepInfo{ID: "an-awesome-step"},
					TestAssets: []testreportfiller.TestReportAssetInfo{
						testreportfiller.TestReportAssetInfo{
							Filename:    "my-important-asset",
							Filesize:    121,
							DownloadURL: "http://dont.call.me.pls",
						},
						testreportfiller.TestReportAssetInfo{
							Filename:    "another-important-asset",
							Filesize:    534,
							DownloadURL: "http://dont.call.me.pls",
						},
					},
				},
				testreportfiller.TestReportWithTestSuites{
					ID: id2,
					TestSuites: []junitmodels.Suite{
						junitmodels.Suite{
							Totals: junitmodels.Totals{
								Tests:   4,
								Passed:  1,
								Skipped: 1,
								Failed:  1,
								Error:   1,
							},
							Tests: []junitmodels.Test{
								junitmodels.Test{
									Name:   "failing test",
									Status: "failed",
									Error:  junitmodels.Error{},
									Properties: map[string]string{
										"name": "failing test",
									},
								},
								junitmodels.Test{
									Name:   "erroneous test",
									Status: "error",
									Error:  junitmodels.Error{},
									Properties: map[string]string{
										"name": "erroneous test",
									},
								},
							},
						},
					},
					StepInfo:   models.StepInfo{Version: "1.0"},
					TestAssets: []testreportfiller.TestReportAssetInfo{},
				},
			},
			expErr: "",
		},
		{
			name:                  "when the test report file is not found",
			xml:                   "",
			statusFromXMLDownload: 404,
			expErr:                "Failed to get test report XML",
		},
		{
			name:                  "when the test report file is not valid",
			xml:                   "<xml?>",
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
			got, err := filler.FillMore(trs, &TestFAPI{}, &junit.Client{}, httpClient, tc.statusFilter)

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
