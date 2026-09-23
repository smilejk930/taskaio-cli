package apiclient

import (
	"bytes"
	"encoding/json"
)

// Record explicit nulls separately from omitted fields in JSON update input.
func readNullFields(data []byte) (map[string]bool, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	nulls := make(map[string]bool)
	for name, raw := range fields {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			nulls[name] = true
		}
	}
	return nulls, nil
}

func marshalNullableUpdate(value any, nulls map[string]bool) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil || len(nulls) == 0 {
		return encoded, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		return nil, err
	}
	for name := range nulls {
		fields[name] = json.RawMessage("null")
	}
	return json.Marshal(fields)
}

func (input *UpdateProjectInput) UnmarshalJSON(data []byte) error {
	type plain UpdateProjectInput
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	nulls, err := readNullFields(data)
	if err != nil {
		return err
	}
	*input = UpdateProjectInput(value)
	input.nullFields = nulls
	return nil
}

func (input UpdateProjectInput) MarshalJSON() ([]byte, error) {
	type plain UpdateProjectInput
	return marshalNullableUpdate(plain(input), input.nullFields)
}

func (input *UpdateTaskInput) UnmarshalJSON(data []byte) error {
	type plain UpdateTaskInput
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	nulls, err := readNullFields(data)
	if err != nil {
		return err
	}
	*input = UpdateTaskInput(value)
	input.nullFields = nulls
	return nil
}

func (input UpdateTaskInput) MarshalJSON() ([]byte, error) {
	type plain UpdateTaskInput
	return marshalNullableUpdate(plain(input), input.nullFields)
}

func (input *UpdateScheduleInput) UnmarshalJSON(data []byte) error {
	type plain UpdateScheduleInput
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	nulls, err := readNullFields(data)
	if err != nil {
		return err
	}
	*input = UpdateScheduleInput(value)
	input.nullFields = nulls
	return nil
}

func (input UpdateScheduleInput) MarshalJSON() ([]byte, error) {
	type plain UpdateScheduleInput
	return marshalNullableUpdate(plain(input), input.nullFields)
}
