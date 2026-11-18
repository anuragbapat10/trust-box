package repository

import (
	"context"
	"sort"
	"testing"

	"github.com/google/uuid"
)

// CREATES

func TestInMemoryRepository_Create(t *testing.T) {
	repo := NewInMemoryRepository()

	inst := Instance{ID: uuid.NewString(), State: StateActive}

	err := repo.Create(context.Background(), inst)
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}
	if len(repo.instances) != 1 {
		t.Errorf("Create: want 1 stored instance, got %d", len(repo.instances))
	}
	if stored, ok := repo.instances[inst.ID]; !ok || stored != inst {
		t.Errorf("Create: stored instance mismatch")
	}
}

func TestInMemoryRepository_Create_AtCapacity(t *testing.T) {
	// Test capacity 5
	maxCap := 5
	repo, err := NewInMemoryRepositoryWithCapacity(maxCap)
	if err != nil {
		t.Fatalf("Repository Creation: unexpected error: %v", err)
	}

	ct := 0
	for ct = 0; ct < maxCap; ct++ {
		inst := Instance{ID: uuid.NewString(), State: StateActive}
		if err := repo.Create(context.Background(), inst); err != nil {
			t.Fatalf("Create number %d: %v", ct, err)
		}
	}

	instAtCap := Instance{ID: uuid.NewString(), State: StateActive}
	err = repo.Create(context.Background(), instAtCap)
	if err != ErrAtCapacity {
		t.Fatalf("Create #%d: want ErrAtCapacity, got %v", ct+1, err)
	}
	if len(repo.instances) != ct {
		t.Errorf("Create #%d: want %d stored instance, got %d", ct+1, ct, len(repo.instances))
	}
	if _, ok := repo.instances[instAtCap.ID]; ok {
		t.Errorf("Instance #%d should not have been stored", ct+1)
	}
}

func TestInMemoryRepository_Create_DeletedState(t *testing.T) {
	repo := NewInMemoryRepository()
	inst := Instance{ID: "deleted-id", State: StateDeleted}
	err := repo.Create(context.Background(), inst)
	if err != ErrInvalidState {
		t.Fatalf("Create: want ErrInvalidState, got %v", err)
	}
}

// GETS

func TestInMemoryRepository_Get(t *testing.T) {
	repo := NewInMemoryRepository()

	inst := Instance{ID: "get-test-id", State: StateActive}
	if err := repo.Create(context.Background(), inst); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.Get(context.Background(), inst.ID)
	if err != nil {
		t.Fatalf("Get: unexpected error: %v", err)
	}
	if got != inst {
		t.Errorf("Get: mismatch\ngot:  %+v\nwant:  %+v", got, inst)
	}
}

func TestInMemoryRepository_Get_NotFound(t *testing.T) {
	repo := NewInMemoryRepository()

	got, err := repo.Get(context.Background(), "does-not-exist")
	if err != ErrNotFound {
		t.Fatalf("Get: want ErrNotFound, got %v", err)
	}
	if got != (Instance{}) {
		t.Errorf("Get: want zero value, got %+v", got)
	}
}

// LISTS

func TestInMemoryRepository_List_Empty(t *testing.T) {
	repo := NewInMemoryRepository()

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List: unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("List: want 0, got %d", len(got))
	}
}

func TestInMemoryRepository_List(t *testing.T) {
	repo, err := NewInMemoryRepositoryWithCapacity(2)
	if err != nil {
		t.Fatalf("Repository Creation: %v", err)
	}
	inst1 := Instance{ID: "a", State: StateActive}
	inst2 := Instance{ID: "b", State: StateActive}
	if err := repo.Create(context.Background(), inst1); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Create(context.Background(), inst2); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List: unexpected error: %v", err)
	}

	// Check count
	if len(got) != 2 {
		t.Fatalf("List: want 2, got %d", len(got))
	}

	// Match individual instances
	sort.Slice(got, func(i, j int) bool { return got[i].ID < got[j].ID })
	if got[0] != inst1 {
		t.Errorf("List: got[0] mismatch\ngot:  %+v\nwant: %+v", got[0], inst1)
	}
	if got[1] != inst2 {
		t.Errorf("List: got[1] mismatch\ngot:  %+v\nwant: %+v", got[1], inst2)
	}
}

func TestInMemoryRepository_List_ExcludesDeleted(t *testing.T) {
	repo, err := NewInMemoryRepositoryWithCapacity(2)
	if err != nil {
		t.Fatalf("Repository Creation: %v", err)
	}

	inst1 := Instance{ID: "a", State: StateActive}
	inst2 := Instance{ID: "b", State: StateActive}
	if err := repo.Create(context.Background(), inst1); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Create(context.Background(), inst2); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(context.Background(), inst1.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List: unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("List: want 1 active instance, got %d", len(got))
	}
	if got[0] != inst2 {
		t.Errorf("List: expected inst2, got %+v", got[0])
	}
}

// DELETES

func TestInMemoryRepository_Delete_NotFound(t *testing.T) {
	repo := NewInMemoryRepository()

	err := repo.Delete(context.Background(), "does-not-exist")
	if err != ErrNotFound {
		t.Fatalf("Delete: want ErrNotFound, got %v", err)
	}
}

func TestInMemoryRepository_Delete(t *testing.T) {
	repo, err := NewInMemoryRepositoryWithCapacity(2)
	if err != nil {
		t.Fatalf("Repository Creation: %v", err)
	}

	inst1 := Instance{ID: "a", State: StateActive}
	inst2 := Instance{ID: "b", State: StateActive}
	if err := repo.Create(context.Background(), inst1); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Create(context.Background(), inst2); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Delete inst1
	if err := repo.Delete(context.Background(), inst1.ID); err != nil {
		t.Fatalf("Delete: unexpected error: %v", err)
	}

	// Verify deleted instance is tombstoned
	deletedInst, err := repo.Get(context.Background(), inst1.ID)
	if err != nil {
		t.Fatalf("Get deleted instance: %v", err)
	}
	if deletedInst.State != StateDeleted {
		t.Errorf("Delete: want StateDeleted, got %v", deletedInst.State)
	}
	if deletedInst.DeleteTime.IsZero() {
		t.Errorf("Delete: expected non-zero DeleteTime")
	}

	// Verify un-deleted instance remains unchanged
	otherInst, err := repo.Get(context.Background(), inst2.ID)
	if err != nil {
		t.Fatalf("Get other instance: %v", err)
	}
	if otherInst != inst2 {
		t.Errorf("Delete: other instance changed\ngot:  %+v\nwant: %+v", otherInst, inst2)
	}
}

func TestInMemoryRepository_Delete_Idempotent(t *testing.T) {
	repo := NewInMemoryRepository()

	inst := Instance{ID: "a", State: StateActive}
	if err := repo.Create(context.Background(), inst); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(context.Background(), inst.ID); err != nil {
		t.Fatalf("Delete first call: %v", err)
	}

	deletedInst1, err := repo.Get(context.Background(), inst.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	// Second delete on same tombstone
	if err := repo.Delete(context.Background(), inst.ID); err != nil {
		t.Fatalf("Delete second call: %v", err)
	}

	deletedInst2, err := repo.Get(context.Background(), inst.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if deletedInst1 != deletedInst2 {
		t.Errorf("Delete idempotency mismatch\nfirst:  %+v\nsecond: %+v", deletedInst1, deletedInst2)
	}
}
