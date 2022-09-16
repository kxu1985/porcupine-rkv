package porcupine

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

const LOG_PATH = "../"

type rkvInput struct {
	op    bool // false = put, true = get
	value int
}

// a sequential specification of a register
var rkvModel = Model{
	Init: func() interface{} {
		return 0
	},
	// step function: takes a state, input, and output, and returns whether it
	// was a legal operation, along with a new state
	Step: func(state, input, output interface{}) (bool, interface{}) {
		regInput := input.(rkvInput)
		if regInput.op == false {
			return true, regInput.value // always ok to execute a put
		} else {
			readCorrectValue := output == state
			return readCorrectValue, state // state is unchanged
		}
	},
	DescribeOperation: func(input, output interface{}) string {
		inp := input.(rkvInput)
		switch inp.op {
		case true:
			return fmt.Sprintf("get() -> '%d'", output.(int))
		case false:
			return fmt.Sprintf("put('%d')", inp.value)
		}
		return "<invalid>" // unreachable
	},
}

func checkFatal(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func readCSV(filePath string) [][]string {
	f, err := os.Open(filePath)
	checkFatal(err)
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	checkFatal(err)

	return records
}

func TestRKVLog(t *testing.T) {
	files, err := ioutil.ReadDir(LOG_PATH)
	checkFatal(err)

	var records [][]string
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".log" {
			records = append(records, readCSV(LOG_PATH+f.Name())...)
		}
	}

	//fmt.Println(records)
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

	res := CheckOperations(rkvModel, ops)

	if res != true {
		t.Fatal("expected operations to be linearizable")
	}
}
