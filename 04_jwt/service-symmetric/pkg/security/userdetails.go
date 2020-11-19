package security

type UserDetails struct {
	UserID int64    `json:"userId"`
	Login  string   `json:"login"`
	Roles  []string `json:"roles"`
	Issued int64    `json:"iat"`
	Expire int64    `json:"exp"`
}
