package expenses

import (
	"database/sql"

	"github.com/IvanLouren/GoSplit/pkg/models"
	"github.com/google/uuid"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type SplitInput struct {
	UserID uuid.UUID
	Amount float64
}

func (s *Service) CreateExpense(groupID uuid.UUID, paidBy uuid.UUID, description string, amount float64, splits []SplitInput) (models.Expense, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return models.Expense{}, err
	}
	defer tx.Rollback()

	var expense models.Expense
	err = tx.QueryRow(`INSERT INTO expenses (group_id, paid_by, description, amount) VALUES ($1, $2, $3, $4) RETURNING id, group_id, paid_by, description, amount, created_at`, groupID, paidBy, description, amount).Scan(&expense.ID, &expense.GroupID, &expense.PaidBy, &expense.Description, &expense.Amount, &expense.CreatedAt)
	if err != nil {
		return models.Expense{}, err
	}

	for _, split := range splits {
		_, err = tx.Exec(`INSERT INTO expense_splits (expense_id, user_id, amount) VALUES ($1, $2, $3)`, expense.ID, split.UserID, split.Amount)
		if err != nil {
			return models.Expense{}, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return models.Expense{}, err
	}
	return expense, nil
}

func (s *Service) GetExpenses(groupID uuid.UUID) ([]models.Expense, error) {
	expenses, err := s.db.Query(`SELECT id, group_id, paid_by, description, amount, created_at FROM expenses WHERE group_id = $1`, groupID)
	if err != nil {
		return nil, err
	}
	defer expenses.Close()

	var result []models.Expense
	for expenses.Next() {
		var expense models.Expense
		err = expenses.Scan(&expense.ID, &expense.GroupID, &expense.PaidBy, &expense.Description, &expense.Amount, &expense.CreatedAt)
		if err != nil {
			return nil, err
		}
		result = append(result, expense)
	}

	return result, nil
}

func (s *Service) GetExpense(expenseID uuid.UUID) (models.Expense, error) {
	var expense models.Expense
	err := s.db.QueryRow(`SELECT id, group_id, paid_by, description, amount, created_at FROM expenses WHERE id = $1`, expenseID).
		Scan(&expense.ID, &expense.GroupID, &expense.PaidBy, &expense.Description, &expense.Amount, &expense.CreatedAt)
	if err != nil {
		return models.Expense{}, err
	}
	return expense, nil
}

func (s *Service) UpdateExpense(expenseID uuid.UUID, description string, amount float64, splits []SplitInput) (models.Expense, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return models.Expense{}, err
	}
	defer tx.Rollback()

	var expense models.Expense
	err = tx.QueryRow(
		`UPDATE expenses SET description = $1, amount = $2 WHERE id = $3 RETURNING id, group_id, paid_by, description, amount, created_at`,
		description, amount, expenseID,
	).Scan(&expense.ID, &expense.GroupID, &expense.PaidBy, &expense.Description, &expense.Amount, &expense.CreatedAt)
	if err != nil {
		return models.Expense{}, err
	}

	_, err = tx.Exec(`DELETE FROM expense_splits WHERE expense_id = $1`, expenseID)
	if err != nil {
		return models.Expense{}, err
	}

	for _, split := range splits {
		_, err = tx.Exec(`INSERT INTO expense_splits (expense_id, user_id, amount) VALUES ($1, $2, $3)`, expenseID, split.UserID, split.Amount)
		if err != nil {
			return models.Expense{}, err
		}
	}

	return expense, tx.Commit()
}

func (s *Service) DeleteExpense(expenseID uuid.UUID) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM expense_splits WHERE expense_id = $1`, expenseID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM expenses WHERE id = $1`, expenseID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
