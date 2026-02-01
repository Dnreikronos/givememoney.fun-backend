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

// mockWalletService is a test double for WalletService.
type mockWalletService struct {
	createFn            func(ctx context.Context, streamerID uuid.UUID, req *model.WalletRequest) (*model.Wallet, error)
	getByIDFn           func(ctx context.Context, id uuid.UUID) (*model.Wallet, error)
	getByWalletAddressFn func(ctx context.Context, walletAddress string) (*model.Wallet, error)
	getByStreamerIDFn   func(ctx context.Context, streamerID uuid.UUID) ([]model.Wallet, error)
	updateFn            func(ctx context.Context, id uuid.UUID, req *model.WalletUpdateInput) (*model.Wallet, error)
	deleteFn            func(ctx context.Context, id uuid.UUID) error
}

func (m *mockWalletService) Create(ctx context.Context, streamerID uuid.UUID, req *model.WalletRequest) (*model.Wallet, error) {
	if m.createFn != nil {
		return m.createFn(ctx, streamerID, req)
	}
	return nil, nil
}

func (m *mockWalletService) GetByID(ctx context.Context, id uuid.UUID) (*model.Wallet, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockWalletService) GetByWalletAddress(ctx context.Context, walletAddress string) (*model.Wallet, error) {
	if m.getByWalletAddressFn != nil {
		return m.getByWalletAddressFn(ctx, walletAddress)
	}
	return nil, nil
}

func (m *mockWalletService) GetByStreamerID(ctx context.Context, streamerID uuid.UUID) ([]model.Wallet, error) {
	if m.getByStreamerIDFn != nil {
		return m.getByStreamerIDFn(ctx, streamerID)
	}
	return nil, nil
}

func (m *mockWalletService) Update(ctx context.Context, id uuid.UUID, req *model.WalletUpdateInput) (*model.Wallet, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, req)
	}
	return nil, nil
}

func (m *mockWalletService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func TestWalletController_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		idParam        string
		mockGetByID    func(ctx context.Context, id uuid.UUID) (*model.Wallet, error)
		wantStatus     int
		wantBodyContains string
	}{
		{
			name:           "invalid uuid",
			idParam:        "not-a-uuid",
			mockGetByID:    nil,
			wantStatus:     http.StatusBadRequest,
			wantBodyContains: "invalid id",
		},
		{
			name:    "wallet found",
			idParam: uuid.New().String(),
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*model.Wallet, error) {
				return &model.Wallet{ID: id, WalletAddress: "abc123"}, nil
			},
			wantStatus:     http.StatusOK,
			wantBodyContains: "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockWalletService{}
			if tt.mockGetByID != nil {
				mock.getByIDFn = tt.mockGetByID
			}
			ctrl := NewWalletController(mock)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Params = gin.Params{{Key: "id", Value: tt.idParam}}
			ctx.Request = httptest.NewRequest(http.MethodGet, "/wallets/"+tt.idParam, nil)

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

func TestWalletController_GetByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	mock := &mockWalletService{
		getByIDFn: func(ctx context.Context, gotID uuid.UUID) (*model.Wallet, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	ctrl := NewWalletController(mock)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Params = gin.Params{{Key: "id", Value: id.String()}}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/wallets/"+id.String(), nil)

	ctrl.GetByID(ctx)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetByID() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
