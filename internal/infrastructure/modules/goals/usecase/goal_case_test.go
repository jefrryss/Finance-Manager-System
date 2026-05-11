package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	goalDomain "Finance-Manager-System/internal/infrastructure/modules/goals/domain"
	transactionDomain "Finance-Manager-System/internal/infrastructure/modules/transactions/domain"
)

type fakeGoalRepo struct {
	goals         map[uuid.UUID]*goalDomain.Goal
	contributions map[uuid.UUID][]goalDomain.GoalContribution
}

func newFakeGoalRepo() *fakeGoalRepo {
	return &fakeGoalRepo{
		goals:         make(map[uuid.UUID]*goalDomain.Goal),
		contributions: make(map[uuid.UUID][]goalDomain.GoalContribution),
	}
}

func (r *fakeGoalRepo) AddGoal(ctx context.Context, goal *goalDomain.Goal) (uuid.UUID, error) {
	if goal.GoalID == uuid.Nil {
		goal.GoalID = uuid.New()
	}
	r.goals[goal.GoalID] = goal
	return goal.GoalID, nil
}
func (r *fakeGoalRepo) GetGoalsByUser(ctx context.Context, userID uuid.UUID) ([]goalDomain.Goal, error) {
	out := make([]goalDomain.Goal, 0)
	for _, g := range r.goals {
		if g.UserID == userID {
			out = append(out, *g)
		}
	}
	return out, nil
}
func (r *fakeGoalRepo) GetGoalByID(ctx context.Context, userID uuid.UUID, goalID uuid.UUID) (*goalDomain.Goal, error) {
	g, ok := r.goals[goalID]
	if !ok || g.UserID != userID {
		return nil, goalDomain.ErrGoalNotFound
	}
	return g, nil
}
func (r *fakeGoalRepo) UpdateGoal(ctx context.Context, goal *goalDomain.Goal) error {
	r.goals[goal.GoalID] = goal
	return nil
}
func (r *fakeGoalRepo) DeleteGoal(ctx context.Context, userID uuid.UUID, goalID uuid.UUID) error {
	delete(r.goals, goalID)
	return nil
}
func (r *fakeGoalRepo) AddContribution(ctx context.Context, contribution *goalDomain.GoalContribution) (uuid.UUID, error) {
	if contribution.ContributionID == uuid.Nil {
		contribution.ContributionID = uuid.New()
	}
	r.contributions[contribution.GoalID] = append(r.contributions[contribution.GoalID], *contribution)
	return contribution.ContributionID, nil
}
func (r *fakeGoalRepo) IncreaseCurrentAmount(ctx context.Context, userID uuid.UUID, goalID uuid.UUID, amount int64) error {
	r.goals[goalID].CurrentAmount += amount
	return nil
}
func (r *fakeGoalRepo) GetGoalContributions(ctx context.Context, userID uuid.UUID, goalID uuid.UUID) ([]goalDomain.GoalContribution, error) {
	return r.contributions[goalID], nil
}
func (r *fakeGoalRepo) GetMonthlyContributions(ctx context.Context, userID uuid.UUID, goalID uuid.UUID, start time.Time, end time.Time) ([]goalDomain.MonthlyContribution, error) {
	return nil, nil
}

type fakeGoalTransRepo struct {
	tx *transactionDomain.Transaction
}

func (r *fakeGoalTransRepo) GetTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) (*transactionDomain.Transaction, error) {
	if r.tx == nil || r.tx.TransactionID != transactionID {
		return nil, transactionDomain.ErrTransNotFound
	}
	return r.tx, nil
}

type fakeGoalTxManager struct{}

func (m *fakeGoalTxManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestGetGoalDetailsExcessSuggestions(t *testing.T) {
	repo := newFakeGoalRepo()
	userID := uuid.New()
	mainGoalID := uuid.New()
	repo.goals[mainGoalID] = &goalDomain.Goal{
		GoalID:        mainGoalID,
		UserID:        userID,
		NameGoal:      "Main",
		TargetAmount:  100,
		CurrentAmount: 150,
	}
	otherID := uuid.New()
	repo.goals[otherID] = &goalDomain.Goal{
		GoalID:        otherID,
		UserID:        userID,
		NameGoal:      "Other",
		TargetAmount:  1000,
		CurrentAmount: 300,
	}
	uc := NewGoalUseCase(repo, &fakeGoalTransRepo{}, &fakeGoalTxManager{})
	details, err := uc.GetGoalDetails(context.Background(), userID, mainGoalID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if details.ExcessAmount != 50 {
		t.Fatalf("unexpected excess amount: %d", details.ExcessAmount)
	}
	if len(details.RedirectSuggestions) == 0 {
		t.Fatalf("expected redirect suggestion")
	}
}

func TestAddContributionFromTransaction(t *testing.T) {
	repo := newFakeGoalRepo()
	userID := uuid.New()
	goalID := uuid.New()
	repo.goals[goalID] = &goalDomain.Goal{
		GoalID:        goalID,
		UserID:        userID,
		NameGoal:      "Goal",
		TargetAmount:  1000,
		CurrentAmount: 0,
	}
	tx := &transactionDomain.Transaction{
		TransactionID: uuid.New(),
		UserID:        userID,
		Amount:        400,
		CompletedAt:   time.Now().UTC(),
	}
	uc := NewGoalUseCase(repo, &fakeGoalTransRepo{tx: tx}, &fakeGoalTxManager{})
	_, err := uc.AddContribution(context.Background(), userID, goalID, 0, nil, &tx.TransactionID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if repo.goals[goalID].CurrentAmount != 400 {
		t.Fatalf("unexpected amount: %d", repo.goals[goalID].CurrentAmount)
	}
}

