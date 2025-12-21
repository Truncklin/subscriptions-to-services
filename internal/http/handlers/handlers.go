package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{DB: pool}
}

type Subscription struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var s Subscription
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	s.ID = id

	query := `INSERT INTO subscriptions (id, user_id, service_name, price, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := h.DB.Exec(ctx, query, s.ID, s.UserID, s.ServiceName, s.Price, s.StartDate, s.EndDate)
	if err != nil {
		slog.Error("Failed to insert subscription",
			slog.String("s.ID", s.ID),
			slog.String("s.ServiceName", s.ServiceName),
			slog.String("s.Price", strconv.Itoa(s.Price)),
			slog.String("s.UserID", s.UserID),
			slog.String("s.StartDate", s.StartDate),
			slog.Any("s.EndDate", &s.EndDate),
			slog.Any("error", err))

		http.Error(w, "Failed to insert subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": s.ID})
}

func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := chi.URLParam(r, "id")
	query := `SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE id = $1`
	row := h.DB.QueryRow(ctx, query, id)

	var s Subscription
	err := row.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate)
	if err != nil {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(s)
}

func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := chi.URLParam(r, "id")
	query := `DELETE FROM subscriptions WHERE id = $1`
	tag, err := h.DB.Exec(ctx, query, id)
	if err != nil || tag.RowsAffected() == 0 {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateSubscription — PUT /subscriptions/{id}
func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := chi.URLParam(r, "id")
	var s Subscription
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
		UPDATE subscriptions
		SET service_name=$1, price=$2, user_id=$3, start_date=$4, end_date=$5
		WHERE id=$6
	`
	tag, err := h.DB.Exec(ctx, query, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate, id)
	if err != nil || tag.RowsAffected() == 0 {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	s.ID = id
	json.NewEncoder(w).Encode(s)
}
