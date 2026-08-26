package input

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ReadInput reads data from either a file path or standard input (if "-") and unmarshals JSON into target.
func ReadInput(inputSource string, target interface{}) error {
	var rawData []byte
	var err error

	if inputSource == "-" {
		rawData, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
	} else {
		rawData, err = os.ReadFile(inputSource)
		if err != nil {
			return fmt.Errorf("failed to read input file %s: %w", inputSource, err)
		}
	}

	if len(rawData) == 0 {
		return fmt.Errorf("input data is empty")
	}

	if err := json.Unmarshal(rawData, target); err != nil {
		return fmt.Errorf("failed to parse input JSON: %w", err)
	}

	return nil
}
