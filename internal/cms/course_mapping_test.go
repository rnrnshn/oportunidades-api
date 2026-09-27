package cms

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

func TestMapCourseIncludesDetails(t *testing.T) {
	course := queries.Course{
		ID:                pgtype.UUID{Bytes: uuid.New(), Valid: true},
		UniversityID:      pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Name:              "Engenharia",
		DurationYears:     pgtype.Int4{Int32: 4, Valid: true},
		AnnualFee:         numericToPg("45000.00"),
		EntryRequirements: pgtype.Text{String: "12a classe", Valid: true},
	}

	item, err := mapCourse(course)
	if err != nil {
		t.Fatalf("mapCourse returned error: %v", err)
	}
	if item.DurationYears != 4 || item.AnnualFee == "" || item.EntryRequirements != "12a classe" {
		t.Fatalf("course details missing: %+v", item)
	}
}
