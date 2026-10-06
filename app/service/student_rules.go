package service

func BatasSKS(ipk *float64) int {
	if ipk == nil {

		return 18
	}
	switch {
	case *ipk >= 3.00:
		return 24
	case *ipk >= 2.50:
		return 21
	default:
		return 18
	}
}

func SisaSKS(totalSKSDiambil, batasSKS int) int {
	sisa := batasSKS - totalSKSDiambil
	if sisa < 0 {
		return 0
	}
	return sisa
}
