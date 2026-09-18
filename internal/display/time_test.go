package display

import (
	"pico_co2/internal/clockedit"
	"testing"
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
		field  clockedit.EditField
		hour   int
		minute int
		want   string
	}{
		{"hour field", clockedit.FieldHour, 14, 23, "[14]:23"},
		{"minute field", clockedit.FieldMinute, 14, 23, "14:[23]"},
		{"save field", clockedit.FieldSave, 14, 23, "[SAVE] EXIT"},
		{"cancel field", clockedit.FieldCancel, 14, 23, "SAVE [EXIT]"},
		{"single digit hour", clockedit.FieldHour, 9, 5, "[9]:05"},
		{"single digit minute", clockedit.FieldMinute, 9, 5, "9:[05]"},
		{"midnight hour", clockedit.FieldHour, 0, 0, "[0]:00"},
		{"midnight minute", clockedit.FieldMinute, 0, 0, "0:[00]"},
		{"max values hour", clockedit.FieldHour, 23, 59, "[23]:59"},
		{"max values minute", clockedit.FieldMinute, 23, 59, "23:[59]"},
		{"save ignores time", clockedit.FieldSave, 23, 59, "[SAVE] EXIT"},
		{"cancel ignores time", clockedit.FieldCancel, 0, 0, "SAVE [EXIT]"},
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
