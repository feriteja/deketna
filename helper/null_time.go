package helper

import (
	"database/sql"
	"encoding/json"
)

// NullTime wraps sql.NullTime to allow proper JSON encoding
type NullTime struct {
	sql.NullTime
}

// MarshalJSON handles JSON encoding for NullTime
func (nt NullTime) MarshalJSON() ([]byte, error) {
	if nt.Valid {
		return json.Marshal(nt.Time)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON handles JSON decoding for NullTime
func (nt *NullTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nt.Valid = false
		return nil
	}
	err := json.Unmarshal(data, &nt.Time)
	nt.Valid = (err == nil)
	return err
}
