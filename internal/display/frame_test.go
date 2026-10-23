package display

import (
	"errors"
	"image/color"
	"testing"
)

type countedDisplay struct {
	*VirtualDisplay
	sends int
	err   error
}

func (d *countedDisplay) Display() error { d.sends++; return d.err }

func TestFrameDisplay(t *testing.T) {
	for _, tt := range []struct {
		name string
		fail bool
	}{{"unchanged frame", false}, {"retry failed transmission", true}} {
		t.Run(tt.name, func(t *testing.T) {
			target := &countedDisplay{VirtualDisplay: NewVirtualDisplay(2, 2)}
			frame := NewFrameDisplay(target)
			if tt.fail {
				target.err = errors.New("send failed")
			}
			if err := frame.Display(); (err != nil) != tt.fail {
				t.Fatalf("error=%v", err)
			}
			target.err = nil
			if err := frame.Display(); err != nil {
				t.Fatal(err)
			}
			want := 1
			if tt.fail {
				want = 2
			}
			if target.sends != want {
				t.Fatalf("sends=%d want=%d", target.sends, want)
			}
			frame.SetPixel(0, 0, color.RGBA{R: 255, A: 255})
			if err := frame.Display(); err != nil {
				t.Fatal(err)
			}
			if target.sends != want+1 {
				t.Fatal("changed frame was skipped")
			}
		})
	}
}
