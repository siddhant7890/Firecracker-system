package whatsapp

import (
	"fmt"
	"net/url"
	"strings"

	"salestrack/internal/billing"
)

// BuildLink generates a WhatsApp "Click to Chat" deep link (wa.me) prefilled
// with a bill summary, so the sales agent can tap "Send to WhatsApp" and
// share the bill with the customer without any WhatsApp Business API
// integration. Swap this out for the WhatsApp Cloud API later if automatic
// server-side sending (rather than agent-initiated) is needed.
func BuildLink(bill billing.Bill) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Hi %s! Here is your bill from us.\n\n", bill.CustomerName)
	fmt.Fprintf(&sb, "Bill No: %s\n", bill.BillNo)
	for _, it := range bill.Items {
		fmt.Fprintf(&sb, "%s x%d - Rs.%.2f\n", it.ProductName, it.Qty, it.TotalAmount)
	}
	fmt.Fprintf(&sb, "\nTotal: Rs.%.2f\n", bill.TotalAmount)
	sb.WriteString("Thank you for your purchase!")

	target := bill.CustomerMobile
	base := "https://wa.me/"
	if target != "" {
		base += "91" + target
	}
	return base + "?text=" + url.QueryEscape(sb.String())
}
