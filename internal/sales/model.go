package sales

// WhatsAppLinkResponse is returned after "Send to WhatsApp" so the mobile
// app can open the link (wa.me) in the WhatsApp app / browser.
type WhatsAppLinkResponse struct {
	Link string `json:"link"`
}
