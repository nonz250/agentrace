package repository

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

// CursorInfo represents the pagination cursor data
type CursorInfo struct {
	SortValue string `json:"sv"` // Sort key value (RFC3339Nano format)
	ID        string `json:"id"` // ID to distinguish items with same sort value
}

// EncodeCursor creates a base64-encoded cursor from sort value and ID
func EncodeCursor(sortValue time.Time, id string) string {
	info := CursorInfo{
		SortValue: sortValue.Format(time.RFC3339Nano),
		ID:        id,
	}
	data, _ := json.Marshal(info)
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeCursor parses a base64-encoded cursor
// Returns nil if cursor is empty or invalid
func DecodeCursor(cursor string) *CursorInfo {
	if cursor == "" {
		return nil
	}

	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil
	}

	var info CursorInfo
	if err := json.Unmarshal(decoded, &info); err != nil {
		return nil
	}

	return &info
}

// ParseCursorTime parses the sort value from cursor info as time.Time
func (c *CursorInfo) ParseSortTime() (time.Time, error) {
	return time.Parse(time.RFC3339Nano, c.SortValue)
}
