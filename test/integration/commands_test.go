package integration

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

func TestCommands(t *testing.T) {
	for _, tc := range []struct{ path, want string }{
		{"01_linkedlist", "reverse=[3 9 2]"}, {"02_stack", "valid=true crossed=false"}, {"03_queue", "dequeue=1/1 front=2 len=2"}, {"04_buffer", "unread=\"API\""}, {"05_ringbuffer", "overflow-accepted=false"}, {"06_hashtable", "len=4 cap=8 load=0.50"}, {"07_visualizations", ""},
	} {
		t.Run(tc.path, func(t *testing.T) {
			out, err := exec.Command("go", "run", "../../cmd/"+tc.path).CombinedOutput()
			if err != nil {
				t.Fatalf("%v: %s", err, out)
			}
			if tc.path == "07_visualizations" {
				var scenes []json.RawMessage
				if json.Unmarshal(out, &scenes) != nil || len(scenes) != 6 {
					t.Fatal("trace JSON")
				}
			} else if !strings.Contains(string(out), tc.want) {
				t.Fatal(string(out))
			}
		})
	}
}
