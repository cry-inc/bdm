package server

// Roles is a struct to describe permissions of users and tokens
type Roles struct {
	Reader bool
	Writer bool
	Admin  bool
}
