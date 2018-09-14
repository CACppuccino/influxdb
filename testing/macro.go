package testing

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/platform"
	kerrors "github.com/influxdata/platform/kit/errors"
	"github.com/influxdata/platform/mock"
)

const (
	idA = "020f755c3c082000"
	idB = "020f755c3c082001"
	idC = "020f755c3c082002"
)

var macroCmpOptions = cmp.Options{
	cmp.Comparer(func(x, y []byte) bool {
		return bytes.Equal(x, y)
	}),
	cmp.Transformer("Sort", func(in []*platform.Macro) []*platform.Macro {
		out := append([]*platform.Macro(nil), in...)
		sort.Slice(out, func(i, j int) bool {
			return out[i].ID.String() > out[j].ID.String()
		})
		return out
	}),
}

// MacroFields defines fields for a macro test
type MacroFields struct {
	Macros      []*platform.Macro
	IDGenerator platform.IDGenerator
}

// CreateMacro tests platform.MacroService CreateMacro interface method
func CreateMacro(init func(MacroFields, *testing.T) (platform.MacroService, func()), t *testing.T) {
	type args struct {
		macro *platform.Macro
	}
	type wants struct {
		err    error
		macros []*platform.Macro
	}

	tests := []struct {
		name   string
		fields MacroFields
		args   args
		wants  wants
	}{
		{
			name: "creating a macro assigns the macro an id and adds it to the store",
			fields: MacroFields{
				IDGenerator: &mock.IDGenerator{
					IDFn: func() platform.ID {
						return idFromString(t, idA)
					},
				},
				Macros: []*platform.Macro{
					&platform.Macro{
						ID:   idFromString(t, idB),
						Name: "existing-macro",
						Arguments: platform.MacroArguments{
							Type:   "constant",
							Values: platform.MacroConstantValues{},
						},
					},
				},
			},
			args: args{
				macro: &platform.Macro{
					Name: "my-macro",
					Arguments: platform.MacroArguments{
						Type:   "constant",
						Values: platform.MacroConstantValues{},
					},
				},
			},
			wants: wants{
				err: nil,
				macros: []*platform.Macro{
					&platform.Macro{
						ID:   idFromString(t, idB),
						Name: "existing-macro",
						Arguments: platform.MacroArguments{
							Type:   "constant",
							Values: platform.MacroConstantValues{},
						},
					},
					&platform.Macro{
						ID:   idFromString(t, idA),
						Name: "my-macro",
						Arguments: platform.MacroArguments{
							Type:   "constant",
							Values: platform.MacroConstantValues{},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		s, done := init(tt.fields, t)
		defer done()
		ctx := context.TODO()

		err := s.CreateMacro(ctx, tt.args.macro)
		diffErrors(err, tt.wants.err, t)

		macros, err := s.FindMacros(ctx)
		if err != nil {
			t.Fatalf("failed to retrieve macros: %v", err)
		}
		if diff := cmp.Diff(macros, tt.wants.macros, macroCmpOptions...); diff != "" {
			t.Fatalf("found unexpected macros -got/+want\ndiff %s", diff)
		}
	}
}

// FindMacroByID tests platform.MacroService FindMacroByID interface method
func FindMacroByID(init func(MacroFields, *testing.T) (platform.MacroService, func()), t *testing.T) {
	type args struct {
		id platform.ID
	}
	type wants struct {
		err   error
		macro *platform.Macro
	}

	tests := []struct {
		name   string
		fields MacroFields
		args   args
		wants  wants
	}{
		{
			name: "finding a macro that exists by id",
			fields: MacroFields{
				Macros: []*platform.Macro{
					&platform.Macro{
						ID:   idFromString(t, idA),
						Name: "existing-macro-a",
						Arguments: platform.MacroArguments{
							Type:   "constant",
							Values: platform.MacroConstantValues{},
						},
					},
					&platform.Macro{
						ID:   idFromString(t, idB),
						Name: "existing-macro-b",
						Arguments: platform.MacroArguments{
							Type:   "constant",
							Values: platform.MacroConstantValues{},
						},
					},
				},
			},
			args: args{
				id: idFromString(t, idB),
			},
			wants: wants{
				err: nil,
				macro: &platform.Macro{
					ID:   idFromString(t, idB),
					Name: "existing-macro-b",
					Arguments: platform.MacroArguments{
						Type:   "constant",
						Values: platform.MacroConstantValues{},
					},
				},
			},
		},
		{
			name: "finding a non-existant macro",
			fields: MacroFields{
				Macros: []*platform.Macro{},
			},
			args: args{
				id: idFromString(t, idA),
			},
			wants: wants{
				err:   kerrors.Errorf(kerrors.NotFound, "macro with ID %s not found", idA),
				macro: nil,
			},
		},
	}

	for _, tt := range tests {
		s, done := init(tt.fields, t)
		defer done()
		ctx := context.TODO()

		macro, err := s.FindMacroByID(ctx, tt.args.id)
		diffErrors(err, tt.wants.err, t)

		if diff := cmp.Diff(macro, tt.wants.macro); diff != "" {
			t.Fatalf("found unexpected macro -got/+want\ndiff %s", diff)
		}
	}
}

// UpdateMacro tests platform.MacroService UpdateMacro interface method
func UpdateMacro(init func(MacroFields, *testing.T) (platform.MacroService, func()), t *testing.T) {
	type args struct {
		id     platform.ID
		update *platform.MacroUpdate
	}
	type wants struct {
		err    error
		macros []*platform.Macro
	}

	tests := []struct {
		name   string
		fields MacroFields
		args   args
		wants  wants
	}{
		{
			name: "updating a macro's name",
			fields: MacroFields{
				Macros: []*platform.Macro{
					&platform.Macro{
						ID:   idFromString(t, idA),
						Name: "existing-macro-a",
						Arguments: platform.MacroArguments{
							Type:   "constant",
							Values: platform.MacroConstantValues{},
						},
					},
					&platform.Macro{
						ID:   idFromString(t, idB),
						Name: "existing-macro-b",
						Arguments: platform.MacroArguments{
							Type:   "constant",
							Values: platform.MacroConstantValues{},
						},
					},
				},
			},
			args: args{
				id: idFromString(t, idB),
				update: &platform.MacroUpdate{
					Name: "new-macro-b-name",
				},
			},
			wants: wants{
				err: nil,
				macros: []*platform.Macro{
					&platform.Macro{
						ID:   idFromString(t, idA),
						Name: "existing-macro-a",
						Arguments: platform.MacroArguments{
							Type:   "constant",
							Values: platform.MacroConstantValues{},
						},
					},
					&platform.Macro{
						ID:   idFromString(t, idB),
						Name: "new-macro-b-name",
						Arguments: platform.MacroArguments{
							Type:   "constant",
							Values: platform.MacroConstantValues{},
						},
					},
				},
			},
		},
		{
			name: "updating a non-existant macro fails",
			fields: MacroFields{
				Macros: []*platform.Macro{},
			},
			args: args{
				id:     idFromString(t, idA),
				update: &platform.MacroUpdate{},
			},
			wants: wants{
				err:    fmt.Errorf("macro with ID %s not found (error reference code: 5)", idA),
				macros: []*platform.Macro{},
			},
		},
	}

	for _, tt := range tests {
		s, done := init(tt.fields, t)
		defer done()
		ctx := context.TODO()

		macro, err := s.UpdateMacro(ctx, tt.args.id, tt.args.update)
		diffErrors(err, tt.wants.err, t)

		if tt.args.update.Name != "" && macro.Name != tt.args.update.Name {
			t.Fatalf("macro name not updated")
		}

		macros, err := s.FindMacros(ctx)
		if err != nil {
			t.Fatalf("failed to retrieve macros: %v", err)
		}
		if diff := cmp.Diff(macros, tt.wants.macros, macroCmpOptions...); diff != "" {
			t.Fatalf("found unexpected macros -got/+want\ndiff %s", diff)
		}
	}
}

// DeleteMacro tests platform.MacroService DeleteMacro interface method
func DeleteMacro(init func(MacroFields, *testing.T) (platform.MacroService, func()), t *testing.T) {
	type args struct {
		id platform.ID
	}
	type wants struct {
		err    error
		macros []*platform.Macro
	}

	tests := []struct {
		name   string
		fields MacroFields
		args   args
		wants  wants
	}{
		{
			name: "deleting a macro",
			fields: MacroFields{
				Macros: []*platform.Macro{
					&platform.Macro{
						ID:   idFromString(t, idA),
						Name: "existing-macro",
						Arguments: platform.MacroArguments{
							Type:   "constant",
							Values: platform.MacroConstantValues{},
						},
					},
				},
			},
			args: args{
				id: idFromString(t, idA),
			},
			wants: wants{
				err:    nil,
				macros: []*platform.Macro{},
			},
		},
		{
			name: "deleting a macro that doesn't exist",
			fields: MacroFields{
				Macros: []*platform.Macro{
					&platform.Macro{
						ID:   idFromString(t, idA),
						Name: "existing-macro",
						Arguments: platform.MacroArguments{
							Type:   "constant",
							Values: platform.MacroConstantValues{},
						},
					},
				},
			},
			args: args{
				id: idFromString(t, idB),
			},
			wants: wants{
				err: kerrors.Errorf(kerrors.NotFound, "macro with ID %s not found", idB),
				macros: []*platform.Macro{
					&platform.Macro{
						ID:   idFromString(t, idA),
						Name: "existing-macro",
						Arguments: platform.MacroArguments{
							Type:   "constant",
							Values: platform.MacroConstantValues{},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		s, done := init(tt.fields, t)
		defer done()
		ctx := context.TODO()

		err := s.DeleteMacro(ctx, tt.args.id)
		defer s.PutMacro(ctx, &platform.Macro{
			ID: tt.args.id,
		})
		diffErrors(err, tt.wants.err, t)

		macros, err := s.FindMacros(ctx)
		if err != nil {
			t.Fatalf("failed to retrieve macros: %v", err)
		}
		if diff := cmp.Diff(macros, tt.wants.macros, macroCmpOptions...); diff != "" {
			t.Fatalf("found unexpected macros -got/+want\ndiff %s", diff)
		}
	}
}

func diffErrors(actual, expected error, t *testing.T) {
	if expected == nil && actual != nil {
		t.Fatalf("unexpected error %q", actual.Error())
	}

	if expected != nil && actual == nil {
		t.Fatalf("expected error %q but received nil", expected.Error())
	}

	if expected != nil && actual != nil && expected.Error() != actual.Error() {
		t.Fatalf("expected error %q but received error %q", expected.Error(), actual.Error())
	}
}
