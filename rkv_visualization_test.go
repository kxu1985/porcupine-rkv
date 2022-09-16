package porcupine

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"
)

func visualizeTempFile(t *testing.T, model Model, info linearizationInfo) {
	file, err := ioutil.TempFile("", "*.html")
	if err != nil {
		t.Fatalf("failed to create temp file")
	}
	err = Visualize(model, info, file)
	if err != nil {
		t.Fatalf("visualization failed")
	}
	t.Logf("wrote visualization to %s", file.Name())
}

func TestRKVVisual(t *testing.T) {
	files, err := ioutil.ReadDir(LOG_PATH)
	checkFatal(err)

	var records [][]string
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".log" {
			records = append(records, readCSV(LOG_PATH+f.Name())...)
		}
	}

	var ops []Operation
	for _, row := range records {
		start_epoch, err := strconv.ParseInt(row[2], 10, 64)
		checkFatal(err)

		end_epoch, err := strconv.ParseInt(row[3], 10, 64)
		checkFatal(err)

		value, err := strconv.Atoi(row[1])
		checkFatal(err)

		client_id, err := strconv.Atoi(row[4])
		checkFatal(err)

		if row[0] == "write" {
			ops = append(ops, Operation{
				ClientId: client_id,
				Input:    rkvInput{false, value},
				Call:     start_epoch,
				Output:   0,
				Return:   end_epoch})

		} else if row[0] == "read" {
			ops = append(ops, Operation{
				ClientId: client_id,
				Input:    rkvInput{true, 0},
				Call:     start_epoch,
				Output:   value,
				Return:   end_epoch})
		}
	}

	res, info := CheckOperationsVerbose(rkvModel, ops, 0)

	visualizeTempFile(t, rkvModel, info)

	if res != Ok {
		t.Fatal("expected operations to be linearizable")
	}
}
