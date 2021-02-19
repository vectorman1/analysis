package common

import (
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"net/http"
)

type HttpResponse struct {
	Code int
	Body interface{}
}

var (
	InvalidModelError     = errors.New("invalid model")
	EntityNotFoundError   = errors.New("entity not found")
	DuplicateEntityError  = errors.New("duplicate entity by unique column")
	PasswordTooShortError = errors.New("password is too short")
	DataTooLongError      = errors.New("a field exceeds it's max length")
	CannotBeNullError     = errors.New("a field is null when it cannot be")
	RPCTLSError           = errors.New("unable to load TLS configuration for RPC client")
	InternalServerError   = errors.New("generic error")
)

func GetErrorResponse(err error) (int, HttpResponse) {
	me, ok := err.(*mysql.MySQLError)
	if ok {
		return http.StatusInternalServerError, HttpResponse{
			Code: http.StatusInternalServerError,
			Body: fmt.Sprintf("%d: %s", me.Number, me.Message),
		}
	}

	switch err {
	case InvalidModelError:
		return http.StatusBadRequest, HttpResponse{
			Code: http.StatusBadRequest,
			Body: err.Error(),
		}
	case EntityNotFoundError:
		return http.StatusNotFound, HttpResponse{
			Code: http.StatusNotFound,
			Body: err.Error(),
		}
	case DuplicateEntityError:
		return http.StatusBadRequest, HttpResponse{
			Code: http.StatusBadRequest,
			Body: err.Error(),
		}
	case DataTooLongError:
		return http.StatusBadRequest, HttpResponse{
			Code: http.StatusBadRequest,
			Body: err.Error(),
		}
	case CannotBeNullError:
		return http.StatusBadRequest, HttpResponse{
			Code: http.StatusBadRequest,
			Body: err.Error(),
		}
	default:
		return http.StatusInternalServerError, HttpResponse{
			Code: http.StatusInternalServerError,
			Body: err.Error(),
		}
	}
}
