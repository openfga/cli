package storetest

func GetEffectiveUsers(checkTest ModelTestCheck) []string {
	if len(checkTest.Users) > 0 {
		return checkTest.Users
	}

	return []string{checkTest.User}
}
