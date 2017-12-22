package debug

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"time"

	"github.com/ktt-ol/mqlux/internal/mqlux"
)

func MessagesFromCSV(filename string, fwd func(mqlux.Message)) error {
	var r io.Reader
	if filename == "-" {
		r = os.Stdin
	} else {
		var err error
		r, err = os.Open(filename)
		if err != nil {
			return err
		}
	}

	// messages are handled in parallel
	// limit this with a semaphore chan
	concurrent := 32
	sem := make(chan struct{}, concurrent)
	for i := 0; i < concurrent; i++ {
		sem <- struct{}{}
	}

	send := func(msg mqlux.Message) {
		fwd(msg)
		sem <- struct{}{}
	}

	reader := csv.NewReader(r)
	reader.FieldsPerRecord = 3
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		msg := mqlux.Message{
			Topic:   record[1],
			Payload: ([]byte)(record[2]),
		}
		msg.Time, err = time.Parse(time.RFC3339, record[0])
		if err != nil {
			log.Println("error: found invalid timestamp in CSV", err)
		}

		<-sem
		go send(msg)
	}

	for i := 0; i < concurrent; i++ {
		<-sem
	}
	return nil
}
