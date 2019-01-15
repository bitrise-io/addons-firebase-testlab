package models

import (
	"testing"

	"github.com/satori/go.uuid"
)

func Test_TestReport_PathInBucket(t *testing.T) {
	id, err := uuid.FromString("addb692e-18d6-11e9-ab14-d663bd873d93")
	if err != nil {
		t.Fatal(err)
	}

	tr := TestReport{
		ID:        id,
		Filename:  "test.xml",
		BuildSlug: "buildslug1",
	}

	exp := "builds/buildslug1/test_reports/addb692e-18d6-11e9-ab14-d663bd873d93/test.xml"
	got := tr.PathInBucket()
	if got != exp {
		t.Error(
			"expected:", exp,
			"got:", got)
	}
}
