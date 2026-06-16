package mockserver

import "testing"

func TestCancelFileBatchDoesNotResurrectDeletedFiles(t *testing.T) {
	store := newVectorStoreStore()
	vectorStore := store.create("test", nil)
	batch, ok := store.createFileBatch(vectorStore.id, []string{"file_a", "file_b"})
	if !ok {
		t.Fatal("createFileBatch() ok = false")
	}
	if !store.deleteFile(vectorStore.id, "file_a") {
		t.Fatal("deleteFile() ok = false")
	}

	cancelled, ok := store.cancelFileBatch(vectorStore.id, batch.id)
	if !ok {
		t.Fatal("cancelFileBatch() ok = false")
	}
	if cancelled.status != "cancelled" {
		t.Fatalf("cancelled.status = %q, want cancelled", cancelled.status)
	}
	if _, ok := store.getFile(vectorStore.id, "file_a"); ok {
		t.Fatal("deleted file was resurrected by cancelFileBatch")
	}
	file, ok := store.getFile(vectorStore.id, "file_b")
	if !ok {
		t.Fatal("remaining file missing after cancelFileBatch")
	}
	if file.status != "cancelled" {
		t.Fatalf("file.status = %q, want cancelled", file.status)
	}
}
