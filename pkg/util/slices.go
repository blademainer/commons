package util

func RemoveElements(slice []interface{}, element ...interface{}) []interface{} {
	for _, f := range element {
		for i, e := range slice {
			if e == f {
				slice = RemoveSliceIndex(slice, i)
			}
		}
	}
	return slice
}

func RemoveSliceIndex(slice []interface{}, index ...int) []interface{} {
	for _, i := range index {
		slice = append(slice[:i], slice[i+1:]...)
	}
	return slice
}
