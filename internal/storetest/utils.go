package storetest

func GetEffectiveUsers(checkTest ModelTestCheck) []string {
	if len(checkTest.Users) > 0 {
		return checkTest.Users
	}

	return []string{checkTest.User}
}

func GetEffectiveObjects(checkTest ModelTestCheck) []string {
	if len(checkTest.Objects) > 0 {
		return checkTest.Objects
	}

	return []string{checkTest.Object}
}
