package messaging

type User struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"` // optional
	UserName  string `json:"username"`  // optional
}
