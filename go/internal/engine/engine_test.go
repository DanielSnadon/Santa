package engine

import (
	"github.com/google/uuid"
	"testing"
)

func TestValidateAssignmentsRejectsMissingAndSelf(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	in := Input{Participants: []uuid.UUID{a, b}}
	if ValidateAssignments(in, map[uuid.UUID]uuid.UUID{a: a, b: b}) == nil {
		t.Fatal("self assignments accepted")
	}
	if ValidateAssignments(in, map[uuid.UUID]uuid.UUID{a: b}) == nil {
		t.Fatal("incomplete assignment accepted")
	}
}
