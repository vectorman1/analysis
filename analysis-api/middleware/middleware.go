package middleware

import (
	"context"

	"github.com/vectorman1/analysis/analysis-api/model"

	"github.com/dgrijalva/jwt-go"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// codeToLevel redirects OK to DEBUG level logging instead of INFO
// This is example how you can log several gRPC code results
func codeToLevel(code codes.Code) zapcore.Level {
	if code == codes.OK {
		// It is DEBUG
		return zap.DebugLevel
	}
	return grpc_zap.DefaultCodeToLevel(code)
}

// LoadMiddleware returns grpc.Server config option with loaded middlewares
func LoadMiddleware(logger *zap.Logger, opts []grpc.ServerOption) []grpc.ServerOption {
	// Shared options for the logger-grpc, with a custom gRPC code to log level function.
	o := []grpc_zap.Option{
		grpc_zap.WithLevels(codeToLevel),
	}

	// Make sure that log statements internal to gRPC library are logged using the zapLogger as well.
	grpc_zap.ReplaceGrpcLoggerV2(logger)

	// Add authentication
	// Add unary interceptor
	opts = append(opts, grpc_middleware.WithUnaryServerChain(
		//grpc_auth.UnaryServerInterceptor(authorizeToken),
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.UnaryServerInterceptor(logger, o...),
	))

	// Add stream interceptor
	opts = append(opts, grpc_middleware.WithStreamServerChain(
		grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.StreamServerInterceptor(logger, o...),
	))

	return opts
}

// authorizeToken parses the token and stores the user info in the request context
func authorizeToken(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
	if err != nil {
		return nil, err
	}

	claims := &model.Token{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	if err != nil {
		return nil, err
	}

	newCtx := context.WithValue(ctx, "user_info", claims)

	return newCtx, nil
}
