package http_rest

import (
	"context"
	"net/http"

	"github.com/vectorman1/analysis/analysis-api/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func HandleMuxError(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	st, ok := status.FromError(err)
	w.Header().Set("Content-Type", "application/json")
	if ok {
		switch st.Code() {
		case codes.Unimplemented:
			w.WriteHeader(http.StatusMethodNotAllowed)
			res, _ := marshaler.Marshal(&model.HttpResponse{Code: http.StatusMethodNotAllowed, Message: st.Message(), Details: st.Details()})
			_, _ = w.Write(res)
			return
		case codes.InvalidArgument:
			w.WriteHeader(http.StatusBadRequest)
			res, _ := marshaler.Marshal(&model.HttpResponse{Code: http.StatusBadRequest, Message: st.Message(), Details: st.Details()})
			_, _ = w.Write(res)
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			res, _ := marshaler.Marshal(&model.HttpResponse{Code: http.StatusInternalServerError, Message: st.Message()})
			_, _ = w.Write(res)
			return
		}
	}
}
