package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type eventReader func(io.Reader, func(map[string]interface{}) error) error

var eventReaders = map[string]eventReader{
	"plain": readPlainEvents,
	"json":  readJSONEvents,
}

func findEventReader(format string) (eventReader, error) {
	r := eventReaders[format]
	if r == nil {
		return nil, fmt.Errorf("format '%v' not supported", format)
	}
	return r, nil
}

func readEvents(format, inFile string) ([]map[string]interface{}, error) {
	type reader func(io.Reader, func(map[string]interface{}) error) error

	r := eventReaders[format]
	if r == nil {
		return nil, fmt.Errorf("format '%v' not supported", format)
	}

	var eventSource io.Reader = os.Stdin
	if inFile != "" {
		f, err := os.Open(inFile)
		if err != nil {
			log.Fatal(err)
		}

		defer f.Close()
		eventSource = f
	}

	var events []map[string]interface{}
	err := r(eventSource, func(event map[string]interface{}) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return events, nil
}

func readPlainEvents(in io.Reader, out func(map[string]interface{}) error) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		msg := scanner.Text()
		out(map[string]interface{}{
			"message": msg,
		})
	}
	return scanner.Err()
}

func readJSONEvents(in io.Reader, out func(map[string]interface{}) error) error {
	dec := json.NewDecoder(in)
	for dec.More() {
		var tmp map[string]interface{}
		if err := dec.Decode(&tmp); err != nil {
			return err
		}
		if err := out(tmp); err != nil {
			return err
		}
	}
	return nil
}
