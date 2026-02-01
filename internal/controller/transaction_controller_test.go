package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// mockTransactionService is a test double for TransactionService.
type mockTransactionService struct {
	createFn          func(ctx context.Context, walletID uuid.UUID, req *model.TransactionRequest) (*model.Transaction, error)
	getByIDFn         func(ctx context.Context, id uuid.UUID) (*model.Transaction, error)
	getAllFn          func(ctx context.Context) (*[]model.Transaction, error)
	getByWalletIDFn   func(ctx context.Context, walletID uuid.UUID) (*[]model.Transaction, error)
	getByStreamerIDFn func(ctx context.Context, streamerID uuid.UUID) (*[]model.Transaction, error)
}

func (m *mockTransactionService) Create(ctx context.Context, walletID uuid.UUID, req *model.TransactionRequest) (*model.Transaction, error) {
	if m.createFn != nil {
		return m.createFn(ctx, walletID, req)
	}
	return nil, nil
}

func (m *mockTransactionService) GetByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockTransactionService) GetAllTransactions(ctx context.Context) (*[]model.Transaction, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx)
	}
	return nil, nil
}

func (m *mockTransactionService) GetByWalletID(ctx context.Context, walletID uuid.UUID) (*[]model.Transaction, error) {
	if m.getByWalletIDFn != nil {
		return m.getByWalletIDFn(ctx, walletID)
	}
	return nil, nil
}

func (m *mockTransactionService) GetByStreamerID(ctx context.Context, streamerID uuid.UUID) (*[]model.Transaction, error) {
	if m.getByStreamerIDFn != nil {
		return m.getByStreamerIDFn(ctx, streamerID)
	}
	return nil, nil
}

func TestTransactionController_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		idParam          string
		mockGetByID      func(ctx context.Context, id uuid.UUID) (*model.Transaction, error)
		wantStatus       int
		wantBodyContains string
	}{
		{
			name:             "invalid uuid",
			idParam:          "not-a-uuid",
			mockGetByID:      nil,
			wantStatus:       http.StatusBadRequest,
			wantBodyContains: "invalid id",
		},
		{
			name:    "transaction found",
			idParam: uuid.New().String(),
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
				return &model.Transaction{ID: id, Message: "thanks"}, nil
			},
			wantStatus:       http.StatusOK,
			wantBodyContains: "thanks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockTransactionService{}
			if tt.mockGetByID != nil {
				mock.getByIDFn = tt.mockGetByID
			}
			ctrl := NewTransactionController(mock)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			ctx.Request = httptest.NewRequest(http.MethodGet, "/transactions/"+tt.idParam, nil)

			ctrl.GetByID(ctx)

			if w.Code != tt.wantStatus {
				t.Errorf("GetByID() status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBodyContains != "" && !strings.Contains(w.Body.String(), tt.wantBodyContains) {
				t.Errorf("GetByID() body = %q, want to contain %q", w.Body.String(), tt.wantBodyContains)
			}
		})
	}
}

func TestTransactionController_GetByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	mock := &mockTransactionService{
		getByIDFn: func(ctx context.Context, gotID uuid.UUID) (*model.Transaction, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	ctrl := NewTransactionController(mock)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Params = gin.Params{{Key: "id", Value: id.String()}}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/transactions/"+id.String(), nil)

	ctrl.GetByID(ctx)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetByID() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
