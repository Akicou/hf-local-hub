package middleware

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrRequestTooLarge is returned when the request body exceeds the configured limit
var ErrRequestTooLarge = errors.New("http: request body too large")

// ErrFileTooLarge is returned when a file exceeds the configured limit
var ErrFileTooLarge = errors.New("http: file too large")

type LimitsConfig struct {
	MaxFileSize    int64
	MaxRequestSize int64
}

type limitedReadCloser struct {
	rc     io.ReadCloser
	n      int64
	limit  int64
	exceed bool
	err    error
}

func (lrc *limitedReadCloser) Read(p []byte) (int, error) {
	if lrc.exceed {
		return 0, lrc.err
	}

	if lrc.n >= lrc.limit {
		lrc.exceed = true
		lrc.err = ErrRequestTooLarge
		return 0, lrc.err
	}

	max := len(p)
	if int64(max) > lrc.limit-lrc.n {
		max = int(lrc.limit - lrc.n)
	}

	n, err := lrc.rc.Read(p[:max])
	lrc.n += int64(n)

	if lrc.n >= lrc.limit {
		lrc.exceed = true
		lrc.err = ErrRequestTooLarge
		return n, lrc.err
	}

	return n, err
}

func (lrc *limitedReadCloser) Close() error {
	return lrc.rc.Close()
}

func NewSizeLimits(cfg LimitsConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > cfg.MaxRequestSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request too large"})
			c.Abort()
			return
		}

		if c.Request.Body != nil {
			c.Request.Body = &limitedReadCloser{
				rc:    c.Request.Body,
				limit: cfg.MaxFileSize,
				err:   ErrRequestTooLarge,
			}
		}

		c.Next()

		// Check if we hit the limit during request processing
		if errors.Is(c.Request.Context().Err(), ErrRequestTooLarge) {
			if !c.Writer.Written() {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "File too large"})
			}
		}
	}
}
