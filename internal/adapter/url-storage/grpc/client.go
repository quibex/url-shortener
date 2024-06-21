package grpc

import (
	"context"
	"fmt"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	usv1 "github.com/quibex/url-storage-api/gen/go/url-storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"log/slog"
	"time"
	"url-shortener/internal/storage"
)

type Client struct {
	api usv1.UrlStorageClient
	log *slog.Logger
}

func New(ctx context.Context,
	log *slog.Logger,
	addr string,
	timeout time.Duration,
	retriesCount int,
) (*Client, error) {
	const op = "grpc.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	cc, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpcretry.UnaryClientInterceptor(retryOpts...),
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api: usv1.NewUrlStorageClient(cc),
	}, nil
}

func (c *Client) SaveURL(ctx context.Context, url, alias string) (int64, error) {
	const op = "grpc.Client.SetURL"

	res, err := c.api.SetUrl(ctx, &usv1.SetUrlRequest{
		Url:   url,
		Alias: alias,
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return -1, storage.ErrURLExists
		}
		return -1, fmt.Errorf("%s: %w", op, err)
	}

	return res.Id, nil
}

func (c *Client) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "grpc.Client.GetURL"

	res, err := c.api.GetUrl(ctx, &usv1.GetUrlRequest{
		Alias: alias,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return res.Url, nil

}

func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
