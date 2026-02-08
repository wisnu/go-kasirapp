package services

import (
	"kasir-api/models"
	"kasir-api/repositories"
)

type TransactionService struct {
	repo *repositories.TransactionRepository
}

func NewTransactionService(repo *repositories.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) Checkout(req *models.CheckoutRequest) (*models.Transaction, error) {
	return s.repo.Checkout(req)
}

func (s *TransactionService) GetAll() ([]models.Transaction, error) {
	return s.repo.GetAll()
}

func (s *TransactionService) GetByID(id int) (*models.Transaction, error) {
	return s.repo.GetByID(id)
}

func (s *TransactionService) GetTodayReport() (*models.DailyReport, error) {
	return s.repo.GetTodayReport()
}

func (s *TransactionService) GetReportByDateRange(startDate, endDate string) (*models.DailyReport, error) {
	return s.repo.GetReportByDateRange(startDate, endDate)
}
