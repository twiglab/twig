package twig

var MaxParam int = 5

var (
	NotFoundHandler = func(c C) error {
		return ErrNotFound
	}

	MethodNotAllowedHandler = func(c C) error {
		return ErrMethodNotAllowed
	}
)
