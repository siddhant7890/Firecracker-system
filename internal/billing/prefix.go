package billing

import "fmt"

// BillPrefixForShop maps a sales_staff.shop_number to the bill-number prefix
// used for that shop's bills (e.g. "SF/A-0001/26-27" for SHOP-AKR).
func BillPrefixForShop(shopNumber string) (string, error) {
	switch shopNumber {
	case "SHOP-AKR":
		return "SF/A", nil
	case "SHOP-14-15":
		return "SF/R", nil
	default:
		return "", fmt.Errorf("unknown shop number %q", shopNumber)
	}
}
