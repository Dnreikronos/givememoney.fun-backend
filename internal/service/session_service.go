package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionService struct {
	db         *gorm.DB
	jwtService *JWTService
}

type SessionCreateRequest struct {
	StreamerID  uuid.UUID
	UserAgent   string
	IPAddress   string
	LoginMethod string
	DeviceType  string
	Country     string
}

type SessionValidationResult struct {
	Session *model.Session
	Claims  *JWTClaims
	IsValid bool
}

func NewSessionService(db *gorm.DB, jwtService *JWTService) *SessionService {
	return &SessionService{
		db:         db,
		jwtService: jwtService,
	}
}

func (s *SessionService) CreateSession(req SessionCreateRequest) (*model.Session, *TokenPair, error) {
	var streamer model.Streamer
	if err := s.db.First(&streamer, req.StreamerID).Error; err != nil {
		return nil, nil, fmt.Errorf("streamer not found: %w", err)
	}

	tokenPair, err := s.jwtService.GenerateTokenPair(
		streamer.ID,
		streamer.Email,
		streamer.Name,
		string(streamer.Provider),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	refreshTokenHash := s.hashToken(tokenPair.RefreshToken)

	refreshToken := &model.RefreshToken{
		StreamerID:   req.StreamerID,
		TokenHash:    refreshTokenHash,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		UserAgent:    req.UserAgent,
		IPAddress:    req.IPAddress,
	}

	if err := s.db.Create(refreshToken).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	sessionToken, err := s.generateSessionToken()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	session := &model.Session{
		StreamerID:     req.StreamerID,
		SessionToken:   sessionToken,
		RefreshTokenID: &refreshToken.ID,
		ExpiresAt:      time.Now().Add(24 * time.Hour),
		UserAgent:      req.UserAgent,
		IPAddress:      req.IPAddress,
		Country:        req.Country,
		DeviceType:     req.DeviceType,
		LoginMethod:    req.LoginMethod,
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, tokenPair, nil
}

func (s *SessionService) ValidateSession(c *gin.Context) (*SessionValidationResult, error) {
	claims, err := s.jwtService.ValidateAccessTokenFromCookie(c)
	if err == nil {
		session, err := s.getActiveSessionByStreamerID(claims.UserID)
		if err != nil {
			return &SessionValidationResult{IsValid: false}, err
		}

		session.UpdateLastAccessed()
		s.db.Save(session)

		return &SessionValidationResult{
			Session: session,
			Claims:  claims,
			IsValid: true,
		}, nil
	}

	newTokenPair, err := s.jwtService.RefreshTokenFromCookie(c)
	if err != nil {
		return &SessionValidationResult{IsValid: false}, errors.New("session expired")
	}

	newClaims, err := s.jwtService.ValidateAccessToken(newTokenPair.AccessToken)
	if err != nil {
		return &SessionValidationResult{IsValid: false}, err
	}

	session, err := s.getActiveSessionByStreamerID(newClaims.UserID)
	if err != nil {
		return &SessionValidationResult{IsValid: false}, err
	}

	session.UpdateLastAccessed()
	session.ExtendExpiry(24 * time.Hour)
	s.db.Save(session)

	return &SessionValidationResult{
		Session: session,
		Claims:  newClaims,
		IsValid: true,
	}, nil
}

func (s *SessionService) RefreshSession(c *gin.Context) (*TokenPair, error) {
	refreshToken, err := s.jwtService.GetTokenFromCookie(c, "refresh")
	if err != nil {
		return nil, errors.New("refresh token not found")
	}

	tokenHash := s.hashToken(refreshToken)

	var dbRefreshToken model.RefreshToken
	if err := s.db.Where("token_hash = ? AND is_revoked = false", tokenHash).First(&dbRefreshToken).Error; err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if !dbRefreshToken.IsValid() {
		return nil, errors.New("refresh token expired or revoked")
	}

	dbRefreshToken.MarkAsUsed()
	dbRefreshToken.Revoke()
	s.db.Save(&dbRefreshToken)

	var streamer model.Streamer
	if err := s.db.First(&streamer, dbRefreshToken.StreamerID).Error; err != nil {
		return nil, fmt.Errorf("streamer not found: %w", err)
	}

	newTokenPair, err := s.jwtService.GenerateTokenPair(
		streamer.ID,
		streamer.Email,
		streamer.Name,
		string(streamer.Provider),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	newRefreshTokenHash := s.hashToken(newTokenPair.RefreshToken)
	newRefreshToken := &model.RefreshToken{
		StreamerID: dbRefreshToken.StreamerID,
		TokenHash:  newRefreshTokenHash,
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		UserAgent:  dbRefreshToken.UserAgent,
		IPAddress:  dbRefreshToken.IPAddress,
	}

	if err := s.db.Create(newRefreshToken).Error; err != nil {
		return nil, fmt.Errorf("failed to create new refresh token: %w", err)
	}

	var session model.Session
	if err := s.db.Where("refresh_token_id = ?", dbRefreshToken.ID).First(&session).Error; err == nil {
		session.RefreshTokenID = &newRefreshToken.ID
		session.ExtendExpiry(24 * time.Hour)
		s.db.Save(&session)
	}

	s.jwtService.SetTokenCookies(c, newTokenPair)

	return newTokenPair, nil
}

func (s *SessionService) InvalidateSession(c *gin.Context) error {
	result, err := s.ValidateSession(c)
	if err != nil || !result.IsValid {
		s.jwtService.ClearTokenCookies(c)
		return nil
	}

	result.Session.Deactivate()
	s.db.Save(result.Session)

	if result.Session.RefreshTokenID != nil {
		var refreshToken model.RefreshToken
		if err := s.db.First(&refreshToken, *result.Session.RefreshTokenID).Error; err == nil {
			refreshToken.Revoke()
			s.db.Save(&refreshToken)
		}
	}

	s.jwtService.ClearTokenCookies(c)

	return nil
}

func (s *SessionService) GetActiveSessions(streamerID uuid.UUID, currentSessionID *uuid.UUID) ([]model.SessionInfo, error) {
	var sessions []model.Session
	if err := s.db.Where("streamer_id = ? AND is_active = true AND expires_at > ?",
		streamerID, time.Now()).Find(&sessions).Error; err != nil {
		return nil, err
	}

	sessionInfos := make([]model.SessionInfo, len(sessions))
	for i, session := range sessions {
		isCurrent := currentSessionID != nil && session.ID == *currentSessionID
		sessionInfos[i] = session.ToSessionInfo(isCurrent)
	}

	return sessionInfos, nil
}

func (s *SessionService) InvalidateAllSessions(streamerID uuid.UUID, exceptSessionID *uuid.UUID) error {
	query := s.db.Model(&model.Session{}).Where("streamer_id = ? AND is_active = true", streamerID)

	if exceptSessionID != nil {
		query = query.Where("id != ?", *exceptSessionID)
	}

	return query.Update("is_active", false).Error
}

func (s *SessionService) CleanupExpiredSessions() error {
	now := time.Now()

	if err := s.db.Model(&model.Session{}).Where("expires_at < ? AND is_active = true", now).
		Update("is_active", false).Error; err != nil {
		return err
	}

	if err := s.db.Model(&model.RefreshToken{}).Where("expires_at < ? AND is_revoked = false", now).
		Update("is_revoked", true).Error; err != nil {
		return err
	}

	return nil
}


func (s *SessionService) generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *SessionService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (s *SessionService) getActiveSessionByStreamerID(streamerID uuid.UUID) (*model.Session, error) {
	var session model.Session
	err := s.db.Where("streamer_id = ? AND is_active = true AND expires_at > ?",
		streamerID, time.Now()).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}