package server

type Permissions interface {
	CanRead(token string) bool
	CanWrite(token string) bool
}

type simplePermissions struct {
	readToken  string
	writeToken string
}

// SimplePermissions returns a simple permission implementation that allows reading
// and uploading based on two single shared secret tokens. An empty token means no
// permission required and everyone is allowed for the corresponding action.
// Please keep in mind that a writing token will also always grant read permission!
func SimplePermissions(readToken, writeToken string) Permissions {
	permissions := simplePermissions{readToken, writeToken}
	return permissions
}

func (s simplePermissions) CanRead(token string) bool {
	return token == s.readToken || token == s.writeToken
}

func (s simplePermissions) CanWrite(token string) bool {
	return token == s.writeToken
}
