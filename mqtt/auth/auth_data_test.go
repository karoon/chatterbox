package auth

var testAuthUserList = []struct {
	username string
	password string
}{
	{"unit-test-vahid", "1234567"},
	{"unit-test-saeed", "7654321"},
}

var testACLUserList = []struct {
	username string
	topic    string
	aclType  string
}{
	{"unit-test-vahid1", "groups/1", AclPub},
	{"unit-test-vahid", "groups/2", AclPubSub},
	{"unit-test-vahid", "groups/3", AclSub},
}
