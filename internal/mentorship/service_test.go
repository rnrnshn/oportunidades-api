package mentorship

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type mockRepository struct {
	listMentorsFn                    func(context.Context, queries.ListMentorsParams) ([]queries.ListMentorsRow, error)
	countMentorsFn                   func(context.Context) (int64, error)
	getMentorByIDFn                  func(context.Context, pgtype.UUID) (queries.GetMentorByIDRow, error)
	createMentorshipSessionFn        func(context.Context, queries.CreateMentorshipSessionParams) (queries.MentorshipSession, error)
	listMentorshipSessionsForUserFn  func(context.Context, queries.ListMentorshipSessionsForUserParams) ([]queries.MentorshipSession, error)
	countMentorshipSessionsForUserFn func(context.Context, pgtype.UUID) (int64, error)
	getMentorshipSessionByIDFn       func(context.Context, pgtype.UUID) (queries.MentorshipSession, error)
	updateMentorshipSessionStatusFn  func(context.Context, queries.UpdateMentorshipSessionStatusParams) (queries.MentorshipSession, error)
}

func (m *mockRepository) ListMentors(ctx context.Context, params queries.ListMentorsParams) ([]queries.ListMentorsRow, error) {
	return m.listMentorsFn(ctx, params)
}
func (m *mockRepository) CountMentors(ctx context.Context) (int64, error) {
	return m.countMentorsFn(ctx)
}
func (m *mockRepository) GetMentorByID(ctx context.Context, userID pgtype.UUID) (queries.GetMentorByIDRow, error) {
	return m.getMentorByIDFn(ctx, userID)
}
func (m *mockRepository) CreateMentorshipSession(ctx context.Context, params queries.CreateMentorshipSessionParams) (queries.MentorshipSession, error) {
	return m.createMentorshipSessionFn(ctx, params)
}
func (m *mockRepository) ListMentorshipSessionsForUser(ctx context.Context, params queries.ListMentorshipSessionsForUserParams) ([]queries.MentorshipSession, error) {
	return m.listMentorshipSessionsForUserFn(ctx, params)
}
func (m *mockRepository) CountMentorshipSessionsForUser(ctx context.Context, userID pgtype.UUID) (int64, error) {
	return m.countMentorshipSessionsForUserFn(ctx, userID)
}
func (m *mockRepository) GetMentorshipSessionByID(ctx context.Context, id pgtype.UUID) (queries.MentorshipSession, error) {
	return m.getMentorshipSessionByIDFn(ctx, id)
}
func (m *mockRepository) UpdateMentorshipSessionStatus(ctx context.Context, params queries.UpdateMentorshipSessionStatusParams) (queries.MentorshipSession, error) {
	return m.updateMentorshipSessionStatusFn(ctx, params)
}

func TestListMentors(t *testing.T) {
	mentorID := uuid.New()
	service := NewService(&mockRepository{
		listMentorsFn: func(context.Context, queries.ListMentorsParams) ([]queries.ListMentorsRow, error) {
			return []queries.ListMentorsRow{{
				UserID:    uuidToPg(mentorID),
				Name:      "Mentor Demo",
				Email:     "mentor@example.com",
				Headline:  "Mentor",
				Bio:       "Bio",
				Expertise: "Go",
			}}, nil
		},
		countMentorsFn: func(context.Context) (int64, error) { return 1, nil },
	})

	result, err := service.ListMentors(context.Background(), PaginationParams{})
	if err != nil {
		t.Fatalf("list mentors returned error: %v", err)
	}
	if len(result.Data) != 1 || result.Meta.Total != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestGetMentorByIDNotFound(t *testing.T) {
	service := NewService(&mockRepository{
		getMentorByIDFn: func(context.Context, pgtype.UUID) (queries.GetMentorByIDRow, error) {
			return queries.GetMentorByIDRow{}, ErrNotFound
		},
	})
	_, err := service.GetMentorByID(context.Background(), uuid.NewString())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestCreateSessionRequest(t *testing.T) {
	mentorID := uuid.New()
	requesterID := uuid.New()
	sessionID := uuid.New()
	service := NewService(&mockRepository{
		getMentorByIDFn: func(context.Context, pgtype.UUID) (queries.GetMentorByIDRow, error) {
			return queries.GetMentorByIDRow{UserID: uuidToPg(mentorID), Name: "Mentor Demo", Email: "mentor@example.com", Headline: "Mentor", Bio: "Bio", Expertise: "Go"}, nil
		},
		createMentorshipSessionFn: func(context.Context, queries.CreateMentorshipSessionParams) (queries.MentorshipSession, error) {
			return queries.MentorshipSession{ID: uuidToPg(sessionID), MentorID: uuidToPg(mentorID), RequesterID: uuidToPg(requesterID), Message: "Gostaria de pedir uma sessao.", Status: "pending", ScheduledAt: pgtype.Timestamptz{Time: time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC), Valid: true}}, nil
		},
	})

	result, err := service.CreateSessionRequest(context.Background(), SessionRequestInput{
		MentorID:    mentorID.String(),
		RequesterID: requesterID.String(),
		Message:     "Gostaria de pedir uma sessao.",
		ScheduledAt: "2026-04-20T09:00:00Z",
	})
	if err != nil {
		t.Fatalf("create session returned error: %v", err)
	}
	if result.Data.Status != "pending" {
		t.Fatalf("unexpected status: %s", result.Data.Status)
	}
}

func TestListSessions(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	service := NewService(&mockRepository{
		listMentorshipSessionsForUserFn: func(context.Context, queries.ListMentorshipSessionsForUserParams) ([]queries.MentorshipSession, error) {
			return []queries.MentorshipSession{{ID: uuidToPg(sessionID), MentorID: uuidToPg(uuid.New()), RequesterID: uuidToPg(userID), Message: "Oi", Status: "pending"}}, nil
		},
		countMentorshipSessionsForUserFn: func(context.Context, pgtype.UUID) (int64, error) { return 1, nil },
	})
	result, err := service.ListSessions(context.Background(), userID.String(), PaginationParams{})
	if err != nil || len(result.Data) != 1 {
		t.Fatalf("unexpected result err=%v result=%+v", err, result)
	}
}

func TestUpdateSessionStatusByMentor(t *testing.T) {
	mentorID := uuid.New()
	requesterID := uuid.New()
	sessionID := uuid.New()
	service := NewService(&mockRepository{
		getMentorshipSessionByIDFn: func(context.Context, pgtype.UUID) (queries.MentorshipSession, error) {
			return queries.MentorshipSession{ID: uuidToPg(sessionID), MentorID: uuidToPg(mentorID), RequesterID: uuidToPg(requesterID), Message: "Oi", Status: "pending"}, nil
		},
		updateMentorshipSessionStatusFn: func(context.Context, queries.UpdateMentorshipSessionStatusParams) (queries.MentorshipSession, error) {
			return queries.MentorshipSession{ID: uuidToPg(sessionID), MentorID: uuidToPg(mentorID), RequesterID: uuidToPg(requesterID), Message: "Oi", Status: "accepted"}, nil
		},
	})
	result, err := service.UpdateSessionStatus(context.Background(), SessionStatusUpdateInput{SessionID: sessionID.String(), ActorID: mentorID.String(), Status: "accepted"})
	if err != nil || result.Data.Status != "accepted" {
		t.Fatalf("unexpected result err=%v result=%+v", err, result)
	}
}

func TestUpdateSessionStatusRejectsRequesterAccept(t *testing.T) {
	mentorID := uuid.New()
	requesterID := uuid.New()
	sessionID := uuid.New()
	service := NewService(&mockRepository{
		getMentorshipSessionByIDFn: func(context.Context, pgtype.UUID) (queries.MentorshipSession, error) {
			return queries.MentorshipSession{ID: uuidToPg(sessionID), MentorID: uuidToPg(mentorID), RequesterID: uuidToPg(requesterID), Message: "Oi", Status: "pending"}, nil
		},
	})
	_, err := service.UpdateSessionStatus(context.Background(), SessionStatusUpdateInput{SessionID: sessionID.String(), ActorID: requesterID.String(), Status: "accepted"})
	if err == nil {
		t.Fatal("expected error")
	}
}
