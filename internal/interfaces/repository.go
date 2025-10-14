package interfaces

import (
	"context"
	"time"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
	"github.com/google/uuid"
)

type StreamerRepositoryInterface interface {
	FindByProvider(ctx context.Context, provider utils.StreamerProvider, providerID string) (*model.Streamer, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Streamer, error)
	FindByEmail(ctx context.Context, email string) (*model.Streamer, error)
	Create(ctx context.Context, streamer *model.Streamer) error
	Update(ctx context.Context, streamer *model.Streamer) error
	CreateStreamer(ctx context.Context, streamer *model.Streamer) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type SessionRepositoryInterface interface {
	CreateSession(session *model.Session) error
	GetSessionByID(id uuid.UUID) (*model.Session, error)
	GetSessionByToken(token string) (*model.Session, error)
	GetActiveSessionsByStreamerID(streamerID uuid.UUID) ([]model.Session, error)
	UpdateSession(session *model.Session) error
	DeactivateSession(id uuid.UUID) error
	DeactivateAllSessionsExcept(streamerID uuid.UUID, exceptSessionID *uuid.UUID) error
	CleanupExpiredSessions() error
	DeleteExpiredSessions(olderThan time.Time) error

	CreateRefreshToken(token *model.RefreshToken) error
	GetRefreshTokenByHash(hash string) (*model.RefreshToken, error)
	GetRefreshTokensByStreamerID(streamerID uuid.UUID) ([]model.RefreshToken, error)
	UpdateRefreshToken(token *model.RefreshToken) error
	RevokeRefreshToken(id uuid.UUID) error
	RevokeAllRefreshTokensForStreamer(streamerID uuid.UUID) error
	CleanupExpiredRefreshTokens() error
	DeleteExpiredRefreshTokens(olderThan time.Time) error

	GetActiveSessionCount() (int64, error)
	GetActiveSessionCountByStreamer(streamerID uuid.UUID) (int64, error)
	GetRecentLoginsByStreamer(streamerID uuid.UUID, since time.Time) ([]model.Session, error)

	GetSessionsByIP(ipAddress string, since time.Time) ([]model.Session, error)
	GetSuspiciousActivities(streamerID uuid.UUID, since time.Time) ([]model.Session, error)
}
