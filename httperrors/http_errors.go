package httperrors

import (
	"net/http"
)

type HttpError struct {
	err    error
	status int // must be a valid HTTP status code
	msg    string
}

func (h *HttpError) Error() string {
	return h.err.Error()
}

func (h *HttpError) Unwrap() error {
	return h.err
}

func New(err error) *HttpError {
	return &HttpError{
		err: err,
	}
}

func (h *HttpError) WithStatus(code int) *HttpError {
	h.status = code
	return h
}

func (h *HttpError) WithMsg(msg string) *HttpError {
	h.msg = msg
	return h
}

func (h *HttpError) Write(w http.ResponseWriter) (int, error) {
	w.WriteHeader(h.status)
	written, err := w.Write([]byte(h.msg))
	if err != nil {
		return written, err
	}
	return written, nil
}

// func (h *HttpError) AsResponse() *http.Response {
// 	return &http.Response{
// 		Body:       io.NopCloser(bufio.NewReader(strings.NewReader(h.msg))),
// 		StatusCode: h.status,
// 		Status:     http.StatusText(h.status),
// 	}
// }
