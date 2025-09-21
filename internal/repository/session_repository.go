package repository

import (
	"time"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/interfaces"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

var _ interfaces.SessionRepositoryInterface = (*SessionRepository)(nil)

func NewSessionRepository(db *gorm.DB) interfaces.SessionRepositoryInterface {
	return &SessionRepository{db: db}
}


func (r *SessionRepository) CreateSession(session *model.Session) error {
	return r.db.Create(session).Error
}

func (r *SessionRepository) GetSessionByID(id uuid.UUID) (*model.Session, error) {
	var session model.Session
	err := r.db.Preload("Streamer").Preload("RefreshToken").First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) GetSessionByToken(token string) (*model.Session, error) {
	var session model.Session
	err := r.db.Preload("Streamer").Preload("RefreshToken").
		Where("session_token = ?", token).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) GetActiveSessionsByStreamerID(streamerID uuid.UUID) ([]model.Session, error) {
	var sessions []model.Session
	err := r.db.Where("streamer_id = ? AND is_active = true AND expires_at > ?",
		streamerID, time.Now()).Find(&sessions).Error
	return sessions, err
}

func (r *SessionRepository) UpdateSession(session *model.Session) error {
	return r.db.Save(session).Error
}

func (r *SessionRepository) DeactivateSession(id uuid.UUID) error {
	return r.db.Model(&model.Session{}).Where("id = ?", id).
		Update("is_active", false).Error
}

func (r *SessionRepository) DeactivateAllSessionsExcept(streamerID uuid.UUID, exceptSessionID *uuid.UUID) error {
	query := r.db.Model(&model.Session{}).Where("streamer_id = ? AND is_active = true", streamerID)

	if exceptSessionID != nil {
		query = query.Where("id != ?", *exceptSessionID)
	}

	return query.Update("is_active", false).Error
}

func (r *SessionRepository) CleanupExpiredSessions() error {
	return r.db.Model(&model.Session{}).Where("expires_at < ? AND is_active = true", time.Now()).
		Update("is_active", false).Error
}

func (r *SessionRepository) DeleteExpiredSessions(olderThan time.Time) error {
	return r.db.Where("expires_at < ? AND is_active = false", olderThan).
		Delete(&model.Session{}).Error
}


func (r *SessionRepository) CreateRefreshToken(token *model.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *SessionRepository) GetRefreshTokenByHash(hash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	err := r.db.Preload("Streamer").Where("token_hash = ?", hash).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *SessionRepository) GetRefreshTokensByStreamerID(streamerID uuid.UUID) ([]model.RefreshToken, error) {
	var tokens []model.RefreshToken
	err := r.db.Where("streamer_id = ? AND is_revoked = false AND expires_at > ?",
		streamerID, time.Now()).Find(&tokens).Error
	return tokens, err
}

func (r *SessionRepository) UpdateRefreshToken(token *model.RefreshToken) error {
	return r.db.Save(token).Error
}

func (r *SessionRepository) RevokeRefreshToken(id uuid.UUID) error {
	return r.db.Model(&model.RefreshToken{}).Where("id = ?", id).
		Update("is_revoked", true).Error
}

func (r *SessionRepository) RevokeAllRefreshTokensForStreamer(streamerID uuid.UUID) error {
	return r.db.Model(&model.RefreshToken{}).Where("streamer_id = ?", streamerID).
		Update("is_revoked", true).Error
}

func (r *SessionRepository) CleanupExpiredRefreshTokens() error {
	return r.db.Model(&model.RefreshToken{}).Where("expires_at < ? AND is_revoked = false", time.Now()).
		Update("is_revoked", true).Error
}

func (r *SessionRepository) DeleteExpiredRefreshTokens(olderThan time.Time) error {
	return r.db.Where("expires_at < ? AND is_revoked = true", olderThan).
		Delete(&model.RefreshToken{}).Error
}


func (r *SessionRepository) GetActiveSessionCount() (int64, error) {
	var count int64
	err := r.db.Model(&model.Session{}).Where("is_active = true AND expires_at > ?", time.Now()).Count(&count).Error
	return count, err
}

func (r *SessionRepository) GetActiveSessionCountByStreamer(streamerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.Session{}).Where("streamer_id = ? AND is_active = true AND expires_at > ?",
		streamerID, time.Now()).Count(&count).Error
	return count, err
}

func (r *SessionRepository) GetRecentLoginsByStreamer(streamerID uuid.UUID, since time.Time) ([]model.Session, error) {
	var sessions []model.Session
	err := r.db.Where("streamer_id = ? AND created_at > ?", streamerID, since).
		Order("created_at DESC").Limit(10).Find(&sessions).Error
	return sessions, err
}


func (r *SessionRepository) GetSessionsByIP(ipAddress string, since time.Time) ([]model.Session, error) {
	var sessions []model.Session
	err := r.db.Where("ip_address = ? AND created_at > ?", ipAddress, since).
		Order("created_at DESC").Find(&sessions).Error
	return sessions, err
}

func (r *SessionRepository) GetSuspiciousActivities(streamerID uuid.UUID, since time.Time) ([]model.Session, error) {
	var sessions []model.Session
	err := r.db.Where("streamer_id = ? AND created_at > ?", streamerID, since).
		Order("created_at DESC").Find(&sessions).Error
	return sessions, err
}