package comparison

import "sort"

func CheckStringArraysEqual(array1 []string, array2 []string) bool {
	if len(array1) != len(array2) {
		return false
	}

	sort.Strings(array1)
	sort.Strings(array2)

	for index, value := range array1 {
		if value != array2[index] {
			return false
		}
	}

	return true
}
