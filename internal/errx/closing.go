package errx

import (
	"context"
	"errors"
	"io"
)

func Close(closer io.Closer, err *error) {
	*err = errors.Join(*err, closer.Close())
}

func CloseCtx(ctx context.Context, closable interface {
	Close(ctx context.Context) error
}, err *error,
) {
	*err = errors.Join(*err, closable.Close(ctx))
}
