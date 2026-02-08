package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"kasir-api/models"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Checkout - create a new transaction with details
func (repo *TransactionRepository) Checkout(req *models.CheckoutRequest) (*models.Transaction, error) {
	// Start database transaction
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Calculate total and prepare transaction details
	var totalAmount int
	var details []models.TransactionDetail

	for _, item := range req.Items {
		// Get product details
		var productID int
		var productName string
		var price float64
		var stock int

		err := tx.QueryRow(
			"SELECT id, name, price, stock FROM products WHERE id = $1",
			item.ProductID,
		).Scan(&productID, &productName, &price, &stock)

		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product with id %d not found", item.ProductID)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get product: %w", err)
		}

		// Check stock availability
		if stock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product %s (available: %d, requested: %d)",
				productName, stock, item.Quantity)
		}

		// Calculate subtotal
		subtotal := int(price * float64(item.Quantity))
		totalAmount += subtotal

		// Update product stock
		_, err = tx.Exec(
			"UPDATE products SET stock = stock - $1 WHERE id = $2",
			item.Quantity, item.ProductID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update product stock: %w", err)
		}

		// Prepare detail (will be inserted after transaction creation)
		details = append(details, models.TransactionDetail{
			ProductID:   item.ProductID,
			ProductName: productName,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	// Create transaction record
	var transactionID int
	err = tx.QueryRow(
		"INSERT INTO transactions (total_amount) VALUES ($1) RETURNING id",
		totalAmount,
	).Scan(&transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Insert transaction details
	for i := range details {
		err = tx.QueryRow(
			"INSERT INTO transaction_details (transaction_id, product_id, quantity, subtotal) VALUES ($1, $2, $3, $4) RETURNING id",
			transactionID, details[i].ProductID, details[i].Quantity, details[i].Subtotal,
		).Scan(&details[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to create transaction detail: %w", err)
		}
		details[i].TransactionID = transactionID
	}

	// Get the created transaction with timestamp
	var transaction models.Transaction
	err = tx.QueryRow(
		"SELECT id, total_amount, created_at FROM transactions WHERE id = $1",
		transactionID,
	).Scan(&transaction.ID, &transaction.TotalAmount, &transaction.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get created transaction: %w", err)
	}

	transaction.Details = details

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &transaction, nil
}

// GetAll - get all transactions
func (repo *TransactionRepository) GetAll() ([]models.Transaction, error) {
	query := "SELECT id, total_amount, created_at FROM transactions ORDER BY created_at DESC"
	rows, err := repo.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := make([]models.Transaction, 0)
	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(&t.ID, &t.TotalAmount, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}

// GetByID - get transaction by ID with details
func (repo *TransactionRepository) GetByID(id int) (*models.Transaction, error) {
	var transaction models.Transaction
	err := repo.db.QueryRow(
		"SELECT id, total_amount, created_at FROM transactions WHERE id = $1",
		id,
	).Scan(&transaction.ID, &transaction.TotalAmount, &transaction.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("transaction not found")
	}
	if err != nil {
		return nil, err
	}

	// Get transaction details
	detailRows, err := repo.db.Query(`
		SELECT td.id, td.transaction_id, td.product_id, p.name, td.quantity, td.subtotal
		FROM transaction_details td
		LEFT JOIN products p ON td.product_id = p.id
		WHERE td.transaction_id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer detailRows.Close()

	details := make([]models.TransactionDetail, 0)
	for detailRows.Next() {
		var d models.TransactionDetail
		var productName sql.NullString
		err := detailRows.Scan(&d.ID, &d.TransactionID, &d.ProductID, &productName, &d.Quantity, &d.Subtotal)
		if err != nil {
			return nil, err
		}
		d.ProductName = productName.String
		details = append(details, d)
	}

	transaction.Details = details
	return &transaction, nil
}

// GetTodayReport - get today's report summary
func (repo *TransactionRepository) GetTodayReport() (*models.DailyReport, error) {
	var report models.DailyReport

	// Get total revenue and total transactions for today
	err := repo.db.QueryRow(`
		SELECT 
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COUNT(*) as total_transaksi
		FROM transactions
		WHERE DATE(created_at) = CURRENT_DATE
	`).Scan(&report.TotalRevenue, &report.TotalTransaksi)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get daily summary: %w", err)
	}

	// Get best selling product for today
	var productName sql.NullString
	var qtyTerjual sql.NullInt64
	
	err = repo.db.QueryRow(`
		SELECT 
			p.name,
			SUM(td.quantity) as qty_terjual
		FROM transaction_details td
		INNER JOIN transactions t ON td.transaction_id = t.id
		INNER JOIN products p ON td.product_id = p.id
		WHERE DATE(t.created_at) = CURRENT_DATE
		GROUP BY p.id, p.name
		ORDER BY qty_terjual DESC
		LIMIT 1
	`).Scan(&productName, &qtyTerjual)

	// If no transactions today, return report with empty best seller
	if err == sql.ErrNoRows {
		report.ProdukTerlaris = models.ProdukTerlaris{
			Nama:       "",
			QtyTerjual: 0,
		}
		return &report, nil
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get best selling product: %w", err)
	}

	report.ProdukTerlaris = models.ProdukTerlaris{
		Nama:       productName.String,
		QtyTerjual: int(qtyTerjual.Int64),
	}

	return &report, nil
}

// GetReportByDateRange - get report summary for a date range
func (repo *TransactionRepository) GetReportByDateRange(startDate, endDate string) (*models.DailyReport, error) {
	var report models.DailyReport

	// Get total revenue and total transactions for date range
	err := repo.db.QueryRow(`
		SELECT 
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COUNT(*) as total_transaksi
		FROM transactions
		WHERE DATE(created_at) >= $1 AND DATE(created_at) <= $2
	`, startDate, endDate).Scan(&report.TotalRevenue, &report.TotalTransaksi)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get report summary: %w", err)
	}

	// Get best selling product for date range
	var productName sql.NullString
	var qtyTerjual sql.NullInt64
	
	err = repo.db.QueryRow(`
		SELECT 
			p.name,
			SUM(td.quantity) as qty_terjual
		FROM transaction_details td
		INNER JOIN transactions t ON td.transaction_id = t.id
		INNER JOIN products p ON td.product_id = p.id
		WHERE DATE(t.created_at) >= $1 AND DATE(t.created_at) <= $2
		GROUP BY p.id, p.name
		ORDER BY qty_terjual DESC
		LIMIT 1
	`, startDate, endDate).Scan(&productName, &qtyTerjual)

	// If no transactions in date range, return report with empty best seller
	if err == sql.ErrNoRows {
		report.ProdukTerlaris = models.ProdukTerlaris{
			Nama:       "",
			QtyTerjual: 0,
		}
		return &report, nil
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get best selling product: %w", err)
	}

	report.ProdukTerlaris = models.ProdukTerlaris{
		Nama:       productName.String,
		QtyTerjual: int(qtyTerjual.Int64),
	}

	return &report, nil
}
