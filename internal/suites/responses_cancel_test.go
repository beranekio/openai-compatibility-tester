package suites

import "testing"

func TestResponsesCancelSkipsWhenAlreadyCompleted(t *testing.T) {
	if !responsesCancelSkipsCancel("completed") {
		t.Fatal("expected completed status to skip cancel")
	}
	for _, status := range []string{"queued", "in_progress"} {
		if responsesCancelSkipsCancel(status) {
			t.Fatalf("status %q should not skip cancel", status)
		}
	}
}

