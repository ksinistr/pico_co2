package display

import (
	"testing"

	"pico_co2/internal/types"
)

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name   string
		hour   int
		minute int
		want   string
	}{
		{"midnight", 0, 0, "0:00"},
		{"single digit hour", 9, 5, "9:05"},
		{"double digit hour", 14, 23, "14:23"},
		{"max values", 23, 59, "23:59"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTime(tt.hour, tt.minute)
			if got != tt.want {
				t.Errorf("FormatTime(%d, %d) = %q, want %q", tt.hour, tt.minute, got, tt.want)
			}
		})
	}
}

func TestFormatEditTime(t *testing.T) {
	tests := []struct {
		name   string
		field  int
		hour   int
		minute int
		want   string
	}{
		{"hour field", types.EditFieldHour, 14, 23, "[14]:23"},
		{"minute field", types.EditFieldMinute, 14, 23, "14:[23]"},
		{"save field", types.EditFieldSave, 14, 23, "[SAVE] EXIT"},
		{"cancel field", types.EditFieldCancel, 14, 23, "SAVE [EXIT]"},
		{"single digit hour", types.EditFieldHour, 9, 5, "[9]:05"},
		{"single digit minute", types.EditFieldMinute, 9, 5, "9:[05]"},
		{"midnight hour", types.EditFieldHour, 0, 0, "[0]:00"},
		{"midnight minute", types.EditFieldMinute, 0, 0, "0:[00]"},
		{"max values hour", types.EditFieldHour, 23, 59, "[23]:59"},
		{"max values minute", types.EditFieldMinute, 23, 59, "23:[59]"},
		{"save ignores time", types.EditFieldSave, 23, 59, "[SAVE] EXIT"},
		{"cancel ignores time", types.EditFieldCancel, 0, 0, "SAVE [EXIT]"},
		{"unknown field falls back to normal", 99, 14, 23, "14:23"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatEditTime(tt.field, tt.hour, tt.minute)
			if got != tt.want {
				t.Errorf("FormatEditTime(%d, %d, %d) = %q, want %q",
					tt.field, tt.hour, tt.minute, got, tt.want)
			}
		})
	}
}
