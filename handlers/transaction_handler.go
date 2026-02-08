package handlers

import (
	"encoding/json"
	"net/http"

	"kasir-api/models"
	"kasir-api/services"
)

type TransactionHandler struct {
	service *services.TransactionService
}

func NewTransactionHandler(service *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

func (h *TransactionHandler) HandleCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate request
	if len(req.Items) == 0 {
		WriteError(w, http.StatusBadRequest, "Items cannot be empty")
		return
	}

	for i, item := range req.Items {
		if item.ProductID <= 0 {
			WriteError(w, http.StatusBadRequest, "Invalid product_id in item "+string(rune(i)))
			return
		}
		if item.Quantity <= 0 {
			WriteError(w, http.StatusBadRequest, "Quantity must be greater than 0")
			return
		}
	}

	// Process checkout
	transaction, err := h.service.Checkout(&req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, transaction)
}

func (h *TransactionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Handle GET, DELETE /api/transactions/{id}
	if r.URL.Path != "/api/transactions" && r.URL.Path != "/api/transactions/" {
		if r.Method != http.MethodGet {
			WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		
		id, err := ParseAndValidateIDFromPath(r.URL.Path, "/api/transactions/")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid transaction ID")
			return
		}

		// GET by ID
		transaction, err := h.service.GetByID(id)
		if err != nil {
			WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, transaction)
		return
	}

	// Handle GET all transactions
	if r.Method == http.MethodGet {
		transactions, err := h.service.GetAll()
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, transactions)
		return
	}

	WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

func (h *TransactionHandler) HandleDailyReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	report, err := h.service.GetDailyReport()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, report)
}
