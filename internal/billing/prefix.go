package billing

import "fmt"

// Bill-number prefixes, one per shop (see BillPrefixForShop).
const (
	PrefixShopAKR  = "SF/A"
	PrefixShop1415 = "SF/R"
)

// BillPrefixForShop maps a sales_staff.shop_number to the bill-number prefix
// used for that shop's bills (e.g. "SF/A-0001/26-27" for SHOP-AKR).
func BillPrefixForShop(shopNumber string) (string, error) {
	switch shopNumber {
	case "SHOP-AKR":
		return PrefixShopAKR, nil
	case "SHOP-14-15":
		return PrefixShop1415, nil
	default:
		return "", fmt.Errorf("unknown shop number %q", shopNumber)
	}
}
