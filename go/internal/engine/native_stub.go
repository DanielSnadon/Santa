//go:build !linux || !cgo

package engine

import (
	"context"
	"fmt"
	"github.com/google/uuid"
)

func nativeDraw(context.Context, []uuid.UUID, []Exclusion, Budget) (Result, error) {
	return Result{Status: InternalError}, fmt.Errorf("native engine requires linux cgo")
}
