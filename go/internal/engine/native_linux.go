//go:build linux && cgo

package engine

/*
#cgo CFLAGS: -I${SRCDIR}/../../../core/include
#cgo LDFLAGS: -L${SRCDIR}/../../../build -lsanta_native
#include <stdlib.h>
#include "santa/abi.h"
*/
import "C"
import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func nativeDraw(ctx context.Context, ids []uuid.UUID, exclusions []Exclusion, budget Budget) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{Status: Cancelled}, err
	}
	index := make(map[uuid.UUID]uint32, len(ids))
	for i, id := range ids {
		index[id] = uint32(i)
	}
	pairs := make([]C.SantaPair, 0, len(exclusions)*2)
	for _, p := range exclusions {
		a, okA := index[p.First]
		b, okB := index[p.Second]
		if !okA || !okB {
			return Result{Status: InvalidInput}, fmt.Errorf("unknown exclusion")
		}
		pairs = append(pairs, C.SantaPair{giver: C.uint32_t(a), receiver: C.uint32_t(b)}, C.SantaPair{giver: C.uint32_t(b), receiver: C.uint32_t(a)})
	}
	out := make([]C.uint32_t, len(ids))
	status := C.santa_draw(C.uint32_t(len(ids)), ptrPair(pairs), C.uint64_t(len(pairs)), C.SantaBudget{max_attempts: C.uint64_t(budget.MaxAttempts), max_work: C.uint64_t(budget.MaxWork), max_memory_bytes: C.uint64_t(budget.MaxMemoryBytes)}, ptrUint(out), C.uint64_t(len(out)))
	result := Result{Status: Status(status)}
	if status != C.SANTA_SUCCESS {
		return result, nil
	}
	result.Assignments = make(map[uuid.UUID]uuid.UUID, len(ids))
	for i, receiver := range out {
		if int(receiver) >= len(ids) {
			return Result{Status: InternalError}, ErrInvalidResult
		}
		result.Assignments[ids[i]] = ids[receiver]
	}
	if err := ValidateAssignments(Input{Participants: ids, Exclusions: exclusions}, result.Assignments); err != nil {
		return Result{Status: InternalError}, err
	}
	return result, nil
}

func ptrPair(v []C.SantaPair) *C.SantaPair {
	if len(v) == 0 {
		return nil
	}
	return &v[0]
}
func ptrUint(v []C.uint32_t) *C.uint32_t {
	if len(v) == 0 {
		return nil
	}
	return &v[0]
}

func init() {
	if C.santa_abi_version() != 1 {
		panic("unsupported santa ABI")
	}
}
