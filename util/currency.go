package util

const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
	AUD = "AUD"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, CAD, AUD:
		return true
	}
	return false
}
