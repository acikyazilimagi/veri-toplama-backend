package util

func getCity(index int) []float64 {
	switch index {
	case 1:
		return []float64{36.852702785393014, 36.87286376953126, 36.535570922786015, 35.88409423828126}
	case 2:
		return []float64{36.2104851748389, 36.81861877441407, 35.84286468375614, 35.82984924316407}
	case 3:
		return []float64{36.495937096205274, 36.649870522206335, 36.064120488812605, 35.4740187605459}
	case 4:
		return []float64{36.50903585150776, 36.402143998719424, 36.47976138594277, 36.31474829364722}
	case 5:
		return []float64{36.64234742932176, 36.3232450328562, 36.53629731173617, 36.029282092441115}
	case 6:
		return []float64{36.116001873480265, 36.06470054394251, 36.0627178139989, 35.91771907373497}
	case 7:
		return []float64{38.53348725642158, 38.78062516773912, 37.32756763881127, 35.45481415037825}
	case 8:
		return []float64{37.35461473302187, 38.0755896764663, 36.85431769725969, 36.67725839531126}
	case 9:
		return []float64{39.065058845523424, 40.013647871307754, 37.86798402826048, 36.687836853946884}
	case 10:
		return []float64{38.160827052916495, 39.33362355320935, 37.44250898099215, 37.35608449070936}
	default:
		return nil
	}
}

func GetCity(index int) []float64 {
	return getCity(index)
}