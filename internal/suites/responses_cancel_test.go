package suites

import "testing"

func TestValidateCompletedResponseHasOutputRejectsEmpty(t *testing.T) {
	if err := validateCompletedResponseHasOutput("responses_cancel", nil); err == nil {
		t.Fatal("expected error for response without output")
	}
}

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

