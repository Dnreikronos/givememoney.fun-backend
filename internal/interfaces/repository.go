package interfaces

import (
	"context"
	"time"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
	"github.com/google/uuid"
)

// StreamerRepositoryInterface defines the contract for streamer data operations
type StreamerRepositoryInterface interface {
	FindByProvider(ctx context.Context, provider utils.StreamerProvider, providerID string) (*model.Streamer, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Streamer, error)
	Create(ctx context.Context, streamer *model.Streamer) error
	Update(ctx context.Context, streamer *model.Streamer) error
	CreateWithWallet(ctx context.Context, streamer *model.Streamer, wallet *model.Wallet) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SessionRepositoryInterface defines the contract for session and refresh token operations
type SessionRepositoryInterface interface {
	// Session operations
	CreateSession(session *model.Session) error
	GetSessionByID(id uuid.UUID) (*model.Session, error)
	GetSessionByToken(token string) (*model.Session, error)
	GetActiveSessionsByStreamerID(streamerID uuid.UUID) ([]model.Session, error)
	UpdateSession(session *model.Session) error
	DeactivateSession(id uuid.UUID) error
	DeactivateAllSessionsExcept(streamerID uuid.UUID, exceptSessionID *uuid.UUID) error
	CleanupExpiredSessions() error
	DeleteExpiredSessions(olderThan time.Time) error

	// Refresh token operations
	CreateRefreshToken(token *model.RefreshToken) error
	GetRefreshTokenByHash(hash string) (*model.RefreshToken, error)
	GetRefreshTokensByStreamerID(streamerID uuid.UUID) ([]model.RefreshToken, error)
	UpdateRefreshToken(token *model.RefreshToken) error
	RevokeRefreshToken(id uuid.UUID) error
	RevokeAllRefreshTokensForStreamer(streamerID uuid.UUID) error
	CleanupExpiredRefreshTokens() error
	DeleteExpiredRefreshTokens(olderThan time.Time) error

	// Statistics and monitoring
	GetActiveSessionCount() (int64, error)
	GetActiveSessionCountByStreamer(streamerID uuid.UUID) (int64, error)
	GetRecentLoginsByStreamer(streamerID uuid.UUID, since time.Time) ([]model.Session, error)

	// Security monitoring
	GetSessionsByIP(ipAddress string, since time.Time) ([]model.Session, error)
	GetSuspiciousActivities(streamerID uuid.UUID, since time.Time) ([]model.Session, error)
}