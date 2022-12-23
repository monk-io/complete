package complete

func contains(arr []string, x string) bool {
	for _, s := range arr {
		if s == x {
			return true
		}
	}
	return false
}
