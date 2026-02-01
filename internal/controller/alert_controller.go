package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AlertController struct{}

func NewAlertController() *AlertController {
	return &AlertController{}
}

// ServeAlertPage returns the HTML page for the donation alert overlay (WebSocket client).
func (ac *AlertController) ServeAlertPage(ctx *gin.Context) {
	streamerID := ctx.Param("streamer_id")

	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { 
            margin: 0; 
            background: transparent; 
            font-family: Arial, sans-serif;
            overflow: hidden;
        }
        .alert { 
            position: fixed; 
            top: 50%; 
            left: 50%; 
            transform: translate(-50%, -50%);
            background: rgba(0, 0, 0, 0.9);
            color: #fff;
            padding: 40px 60px;
            border-radius: 15px;
            font-size: 28px;
            text-align: center;
            display: none;
            box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
            animation: slideIn 0.5s ease-out;
        }
        .amount {
            font-size: 48px;
            font-weight: bold;
            color: #4CAF50;
            margin-bottom: 15px;
        }
        .message {
            font-size: 20px;
            color: #ddd;
            margin-top: 10px;
        }
        @keyframes slideIn {
            from { 
                opacity: 0; 
                transform: translate(-50%, -70%);
            }
            to { 
                opacity: 1; 
                transform: translate(-50%, -50%);
            }
        }
    </style>
</head>
<body>
    <div class="alert" id="alert">
        <div class="amount" id="amount"></div>
        <div>New Donation!</div>
        <div class="message" id="message"></div>
    </div>
    <script>
        const streamerID = '` + streamerID + `';
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const ws = new WebSocket(protocol + '//' + window.location.host + '/api/ws/alerts/' + streamerID);
        
        ws.onopen = () => console.log('Connected to alerts');
        ws.onerror = (error) => console.error('WebSocket error:', error);
        ws.onclose = () => console.log('Disconnected from alerts');
        
        ws.onmessage = (event) => {
            const tx = JSON.parse(event.data);
            const alert = document.getElementById('alert');
            const amount = document.getElementById('amount');
            const message = document.getElementById('message');
            
            amount.textContent = '💰 $' + tx.amount;
            message.textContent = tx.message || '';
            
            alert.style.display = 'block';
            setTimeout(() => {
                alert.style.display = 'none';
            }, 5000);
        };
    </script>
</body>
</html>`

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
