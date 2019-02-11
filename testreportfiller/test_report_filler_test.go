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

	t.Run("when the test report files are found and valid", func(t *testing.T) {

		xml := []byte(`
	    <?xml version="1.0" encoding="UTF-8"?>
	    <testsuites>
	        <testsuite>
	        </testsuite>
	    </testsuites>
	`)

		filler := testreportfiller.Filler{}
		httpClient := NewTestClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBuffer(xml)),
			}
		})

		expSuites := []testreportfiller.TestReportWithTestSuites{
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
		}

		testReportsWithTestSuites, err := filler.Fill(trs, &TestFAPI{}, &junit.Client{}, httpClient)
		require.NoError(t, err)
		require.Equal(t, expSuites, testReportsWithTestSuites)

	})

	t.Run("when the test report file is not found", func(t *testing.T) {
		var resp []byte
		filler := testreportfiller.Filler{}
		httpClient := NewTestClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: 404,
				Body:       ioutil.NopCloser(bytes.NewBuffer(resp)),
			}
		})

		_, err := filler.Fill(trs, &TestFAPI{}, &junit.Client{}, httpClient)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Failed to get test report XML")
	})

	t.Run("when the test report file is not valid", func(t *testing.T) {
		invalidXML := []byte(`
	    <xml?>
	`)

		filler := testreportfiller.Filler{}
		httpClient := NewTestClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBuffer(invalidXML)),
			}
		})

		_, err := filler.Fill(trs, &TestFAPI{}, &junit.Client{}, httpClient)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Failed to parse test report XML")
	})
}
