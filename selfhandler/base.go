package selfhandler

import "net/http"

// Handler is a interface which contains handler
type Handler interface {
	Handler(r *http.Request, ec chan error)
}
