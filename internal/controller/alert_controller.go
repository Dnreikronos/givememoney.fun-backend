package controller

import (
	"fmt"
	"net/http"

	apperrors "github.com/Dnreikronos/givememoney.fun-backend/internal/errors"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/middleware"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AlertController struct {
	alertSettingsService *service.AlertSettingsService
	hub                  *websocket.Hub
}

func NewAlertController(alertSettingsService *service.AlertSettingsService, hub *websocket.Hub) *AlertController {
	return &AlertController{
		alertSettingsService: alertSettingsService,
		hub:                  hub,
	}
}

type AlertTestRequest struct {
	DonorName string `json:"donor_name"`
	Amount    string `json:"amount"`
	Message   string `json:"message"`
}

type AlertTestPayload struct {
	DonorName string `json:"donor_name"`
	Amount    string `json:"amount"`
	Message   string `json:"message"`
}

// SendTestAlert broadcasts a test alert to the authenticated streamer.
func (ac *AlertController) SendTestAlert(ctx *gin.Context) {
	streamerID, exists := ctx.Get("streamer_id")
	if !exists {
		middleware.AbortWithError(ctx, apperrors.NewUnauthorizedError("streamer_id not found in context - authentication required"))
		return
	}

	streamerUUID, ok := streamerID.(uuid.UUID)
	if !ok {
		middleware.AbortWithError(ctx, apperrors.NewInternalError("invalid streamer_id type", nil))
		return
	}

	var req AlertTestRequest
	if ctx.Request.ContentLength > 0 {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(ctx, apperrors.NewValidationError("invalid request body", err))
			return
		}
	}

	payload := AlertTestPayload{
		DonorName: req.DonorName,
		Amount:    req.Amount,
		Message:   req.Message,
	}

	if payload.DonorName == "" {
		payload.DonorName = "CryptoFan42"
	}
	if payload.Amount == "" {
		payload.Amount = "0.5 SOL"
	}
	if payload.Message == "" {
		payload.Message = "Boa live! Continue assim"
	}

	ac.hub.BroadcastToStreamer(streamerUUID, payload)
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// ServeAlertPage returns the HTML page for the donation alert overlay (WebSocket client).
// Settings are loaded from the database if available, otherwise defaults are used.
func (ac *AlertController) ServeAlertPage(ctx *gin.Context) {
	streamerID := ctx.Param("streamer_id")

	// Default values
	bgColor := "#121621"
	textColor := "#F2F4F8"
	messageColor := "#A9B1BF"
	accentColor := "#3B82F6"
	headerText := "Nova Doacao!"
	showDonorName := true
	showAmount := true
	showMessage := true
	alertDuration := 5000
	position := "top"

	// Try to load settings from DB
	if uid, err := uuid.Parse(streamerID); err == nil {
		if settings, err := ac.alertSettingsService.GetByStreamerID(ctx, uid); err == nil && settings != nil {
			bgColor = settings.BackgroundColor
			textColor = settings.TextColor
			messageColor = settings.MessageColor
			accentColor = settings.AccentColor
			headerText = settings.HeaderText
			showDonorName = settings.ShowDonorName
			showAmount = settings.ShowAmount
			showMessage = settings.ShowMessage
			alertDuration = settings.MaxDuration
			if settings.Position != "" {
				position = settings.Position
			}
		}
	}

	showDonorNameJS := "true"
	if !showDonorName {
		showDonorNameJS = "false"
	}
	showAmountJS := "true"
	if !showAmount {
		showAmountJS = "false"
	}
	showMessageJS := "true"
	if !showMessage {
		showMessageJS = "false"
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700;800&display=swap" rel="stylesheet">
    <style>
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

        body {
            background: transparent;
            font-family: 'Inter', -apple-system, sans-serif;
            overflow: hidden;
            width: 100vw;
            height: 100vh;
        }

        .alert-container {
            position: fixed;
            left: 50%%;
            transform: translateX(-50%%);
            width: 420px;
            max-width: calc(100vw - 32px);
            pointer-events: none;
            z-index: 100;
        }

        /* Position variants */
        .alert-container.pos-top    { top: 40px; }
        .alert-container.pos-center { top: 50%%; transform: translate(-50%%, -50%%); }
        .alert-container.pos-bottom { bottom: 40px; }

        .alert-card {
            position: relative;
            background: %s;
            border-radius: 14px;
            border: 1px solid rgba(255, 255, 255, 0.06);
            padding: 16px 20px;
            opacity: 0;
            transform: translateY(-12px) scale(0.98);
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.35);
            overflow: hidden;
        }

        .alert-card.show {
            opacity: 1;
            transform: translateY(0) scale(1);
            animation: alertIn 0.4s cubic-bezier(0.22, 1, 0.36, 1) forwards;
        }

        .alert-card.hide {
            opacity: 0;
            transform: translateY(8px) scale(0.98);
            animation: alertOut 0.3s cubic-bezier(0.55, 0, 1, 0.45) forwards;
        }

        @keyframes alertIn {
            0%% {
                opacity: 0;
                transform: translateY(-12px) scale(0.98);
            }
            100%% {
                opacity: 1;
                transform: translateY(0) scale(1);
            }
        }

        @keyframes alertOut {
            0%% {
                opacity: 1;
                transform: translateY(0) scale(1);
            }
            100%% {
                opacity: 0;
                transform: translateY(8px) scale(0.98);
            }
        }

        .alert-header {
            font-size: 10px;
            font-weight: 700;
            letter-spacing: 0.12em;
            text-transform: uppercase;
            color: %s;
            margin-bottom: 6px;
            opacity: 0.7;
        }

        .alert-header:empty {
            display: none;
        }

        .alert-body {
            display: flex;
            align-items: baseline;
            justify-content: space-between;
            gap: 10px;
        }

        .alert-donor {
            font-size: 16px;
            font-weight: 600;
            color: %s;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            min-width: 0;
            flex: 1 1 auto;
        }

        .alert-amount {
            font-size: 20px;
            font-weight: 700;
            color: %s;
            white-space: nowrap;
            font-variant-numeric: tabular-nums;
            flex: 0 0 auto;
        }

        .alert-message {
            margin-top: 6px;
            font-size: 13px;
            line-height: 1.45;
            color: %s;
            display: -webkit-box;
            -webkit-line-clamp: 2;
            -webkit-box-orient: vertical;
            overflow: hidden;
            overflow-wrap: anywhere;
        }

        .alert-progress {
            position: absolute;
            bottom: 0;
            left: 0;
            right: 0;
            height: 2px;
            background: rgba(255, 255, 255, 0.08);
        }

        .alert-progress-bar {
            height: 100%%;
            background: %s;
            transform-origin: left;
            transform: scaleX(1);
        }

        @keyframes progressShrink {
            0%% { transform: scaleX(1); }
            100%% { transform: scaleX(0); }
        }

        @media (prefers-reduced-motion: reduce) {
            .alert-card,
            .alert-card.show,
            .alert-card.hide {
                animation: none !important;
            }
            .alert-card.show {
                opacity: 1;
                transform: none;
            }
            .alert-card.hide {
                opacity: 0;
                transform: none;
            }
            .alert-progress-bar {
                animation: none !important;
            }
        }
    </style>
</head>
<body>
    <div class="alert-container pos-%s" id="alertContainer">
        <div class="alert-card" id="alertCard">
            <div class="alert-header" id="alertHeader">%s</div>
            <div class="alert-body">
                <div class="alert-donor" id="alertDonor"></div>
                <div class="alert-amount" id="alertAmount"></div>
            </div>
            <div class="alert-message" id="alertMessage"></div>
            <div class="alert-progress">
                <div class="alert-progress-bar" id="alertProgress"></div>
            </div>
        </div>
    </div>

    <script>
        const streamerID = '%s';
        const ALERT_DURATION = %d;
        const SHOW_DONOR_NAME = %s;
        const SHOW_AMOUNT = %s;
        const SHOW_MESSAGE = %s;

        const card = document.getElementById('alertCard');
        const donorEl = document.getElementById('alertDonor');
        const amountEl = document.getElementById('alertAmount');
        const messageEl = document.getElementById('alertMessage');
        const progressEl = document.getElementById('alertProgress');

        // Donation queue
        const queue = [];
        let isShowing = false;

        function showAlert(tx) {
            // Donor name
            if (SHOW_DONOR_NAME && tx.donor_name) {
                donorEl.textContent = tx.donor_name;
                donorEl.style.display = 'block';
            } else {
                donorEl.style.display = 'none';
            }

            // Amount
            if (SHOW_AMOUNT) {
                amountEl.textContent = tx.amount;
                amountEl.style.display = 'block';
            } else {
                amountEl.style.display = 'none';
            }

            // Message
            if (SHOW_MESSAGE && tx.message) {
                messageEl.textContent = tx.message;
                messageEl.style.display = 'block';
            } else {
                messageEl.style.display = 'none';
            }

            // Reset animations
            card.classList.remove('show', 'hide');
            progressEl.style.animation = 'none';
            void card.offsetWidth; // force reflow

            // Show card
            card.classList.add('show');
            progressEl.style.animation = 'progressShrink ' + ALERT_DURATION + 'ms linear forwards';

            isShowing = true;

            // Auto-hide
            setTimeout(function() {
                card.classList.remove('show');
                card.classList.add('hide');

                setTimeout(function() {
                    card.classList.remove('hide');
                    card.style.opacity = '0';
                    isShowing = false;
                    processQueue();
                }, 450);
            }, ALERT_DURATION);
        }

        function processQueue() {
            if (queue.length > 0 && !isShowing) {
                showAlert(queue.shift());
            }
        }

        // WebSocket with auto-reconnect
        function connect() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const ws = new WebSocket(protocol + '//' + window.location.host + '/api/ws/alerts/' + streamerID);

            ws.onopen = function() {
                console.log('[GiveMeMoney] Connected to alert stream');
            };

            ws.onmessage = function(event) {
                try {
                    const tx = JSON.parse(event.data);
                    if (isShowing) {
                        queue.push(tx);
                    } else {
                        showAlert(tx);
                    }
                } catch (e) {
                    console.error('[GiveMeMoney] Parse error:', e);
                }
            };

            ws.onerror = function(error) {
                console.error('[GiveMeMoney] WebSocket error:', error);
            };

            ws.onclose = function() {
                console.log('[GiveMeMoney] Disconnected. Reconnecting in 3s...');
                setTimeout(connect, 3000);
            };
        }

        connect();
    </script>
</body>
</html>`,
		// CSS: alert-card background, header color, donor color, amount color, message color, progress color
		bgColor, accentColor, textColor, accentColor, messageColor, accentColor,
		// HTML: position class, header text
		position, headerText,
		// JS: streamerID, duration, toggles
		streamerID, alertDuration, showDonorNameJS, showAmountJS, showMessageJS)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
