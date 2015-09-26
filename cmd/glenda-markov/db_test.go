package main

import (
	"testing"
)

func TestInsertSuffix(t *testing.T) {
	db, err := NewMarkovDB("test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	idp, err := db.InsertPrefix("cat on a")
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("prefix id %d", idp)

	ids, err := db.InsertSuffix(idp, "hot pink roof")
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("suffix id %d", ids)
}
