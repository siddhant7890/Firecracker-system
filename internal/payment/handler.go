package payment

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterPublicRoutes wires the Razorpay webhook. This endpoint is NOT
// behind our JWT auth (Razorpay's servers call it directly) — it is instead
// authenticated by verifying the X-Razorpay-Signature header against the
// raw body using the webhook secret.
func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.POST("/razorpay/webhook", h.webhook)
}

type webhookPayload struct {
	Event   string `json:"event"`
	Payload struct {
		Payment struct {
			Entity struct {
				ID      string `json:"id"`
				OrderID string `json:"order_id"`
				Status  string `json:"status"`
			} `json:"entity"`
		} `json:"payment"`
	} `json:"payload"`
}

func (h *Handler) webhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	signature := c.GetHeader("X-Razorpay-Signature")
	if !VerifyWebhookSignature(body, signature) {
		log.Println("payment: rejected webhook call with invalid signature")
		c.Status(http.StatusUnauthorized)
		return
	}

	var payload webhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if payload.Event == "payment.captured" {
		entity := payload.Payload.Payment.Entity
		if _, err := h.service.HandleWebhookPayment(c.Request.Context(), entity.OrderID, entity.ID); err != nil {
			log.Printf("payment: could not apply webhook for order %s: %v", entity.OrderID, err)
			// Still 200: Razorpay retries on non-2xx, and a bill that's
			// already approved (e.g. via VerifyPayment) is not an error.
		}
	}

	c.Status(http.StatusOK)
}
