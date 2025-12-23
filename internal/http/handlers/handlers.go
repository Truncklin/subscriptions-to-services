package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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

type SubscriptionCreateRequest struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

type SubscriptionUpdateRequest struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

type Subscription struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// CreateSubscription godoc
// @Summary Создать подписку
// @Description Создаёт новую подписку. start_date и end_date передаются в формате MM-YYYY
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body SubscriptionCreateRequest true "Данные подписки"
// @Success 201 {object} map[string]string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /subscriptions [post]
func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req SubscriptionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	s, err := req.ToModel()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `INSERT INTO subscriptions (id, user_id, service_name, price, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = h.DB.Exec(ctx, query, s.ID, s.UserID, s.ServiceName, s.Price, s.StartDate, s.EndDate)
	if err != nil {
		slog.Error("Failed to insert subscription",
			slog.String("s.ID", s.ID),
			slog.String("s.ServiceName", s.ServiceName),
			slog.String("s.Price", strconv.Itoa(s.Price)),
			slog.String("s.UserID", s.UserID),

			slog.Any("s.EndDate", &s.EndDate),
			slog.Any("error", err))

		http.Error(w, "Failed to insert subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": s.ID})
}

// GetSubscription godoc
// @Summary Получить подписку
// @Description Возвращает подписку по ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки"
// @Success 200 {object} Subscription
// @Failure 404 {string} string
// @Router /subscriptions/{id} [get]
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

// DeleteSubscription godoc
// @Summary Удалить подписку
// @Description Удаляет подписку по ID
// @Tags subscriptions
// @Param id path string true "ID подписки"
// @Success 204
// @Failure 404 {string} string
// @Router /subscriptions/{id} [delete]
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

// UpdateSubscription godoc
// @Summary Обновить подписку
// @Description Полностью обновляет подписку по ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Param subscription body SubscriptionUpdateRequest true "Данные подписки"
// @Success 200 {object} Subscription
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Router /subscriptions/{id} [put]
func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := chi.URLParam(r, "id")

	var req SubscriptionUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	s, err := req.ToModel(id)
	if err != nil {
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
		slog.Error("Failed to update subscription", slog.Any("error", err))
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	s.ID = id
	json.NewEncoder(w).Encode(s)
}

// ListSubscriptions godoc
// @Summary Получить подписки за период
// @Description Фильтрация по периоду
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param from query string false "Начало периода (MM-YYYY)"
// @Param to query string false "Конец периода (MM-YYYY)"
// @Success 200 {array} Subscription
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /subscriptions [get]
func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	q := r.URL.Query()

	var (
		from *time.Time
		to   *time.Time
	)

	if v := q.Get("from"); v != "" {
		t, err := parseMonth(v)
		if err != nil {
			http.Error(w, "invalid from", http.StatusBadRequest)
			return
		}
		from = &t
	}

	if v := q.Get("to"); v != "" {
		t, err := parseMonth(v)
		if err != nil {
			http.Error(w, "invalid to", http.StatusBadRequest)
			return
		}
		to = &t
	}

	query := `
		SELECT id, service_name, price, user_id, start_date, end_date
		FROM subscriptions
		WHERE
		    ($1::timestamp IS NULL OR start_date >= $1)
		AND ($2::timestamp IS NULL OR end_date <= $2)
		`

	rows, err := h.DB.Query(ctx, query, from, to)
	if err != nil {
		slog.Error("query failed", slog.Any("error", err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var result []Subscription

	for rows.Next() {
		var s Subscription
		err := rows.Scan(
			&s.ID,
			&s.ServiceName,
			&s.Price,
			&s.UserID,
			&s.StartDate,
			&s.EndDate,
		)
		if err != nil {
			slog.Error("scan failed", slog.Any("error", err))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		result = append(result, s)
	}

	if err := rows.Err(); err != nil {
		slog.Error("rows error", slog.Any("error", err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (r SubscriptionCreateRequest) ToModel() (*Subscription, error) {

	start, err := parseMonth(r.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date")
	}

	var end *time.Time
	if r.EndDate != nil {
		parsedEnd, err := parseMonth(*r.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date")
		}
		end = &parsedEnd
	}

	return &Subscription{
		ID:          uuid.New().String(),
		UserID:      r.UserID,
		ServiceName: r.ServiceName,
		Price:       r.Price,
		StartDate:   start,
		EndDate:     end,
	}, nil
}

func (r SubscriptionUpdateRequest) ToModel(id string) (*Subscription, error) {

	start, err := parseMonth(r.StartDate)
	if err != nil {
		return nil, err
	}

	var end *time.Time
	if r.EndDate != nil {
		parsedEnd, err := parseMonth(*r.EndDate)
		if err != nil {
			return nil, err
		}
		end = &parsedEnd
	}

	return &Subscription{
		ID:          id,
		UserID:      r.UserID,
		ServiceName: r.ServiceName,
		Price:       r.Price,
		StartDate:   start,
		EndDate:     end,
	}, nil
}
func parseMonth(v string) (time.Time, error) {

	return time.Parse("01-2006", v)
}
