package middleware

import (
	"io"

	"github.com/gin-gonic/gin"
)

type LimitsConfig struct {
	MaxFileSize    int64
	MaxRequestSize int64
}

type limitedReadCloser struct {
	rc     io.ReadCloser
	n      int64
	limit  int64
	exceed bool
}

func (lrc *limitedReadCloser) Read(p []byte) (int, error) {
	if lrc.exceed {
		return 0, io.ErrUnexpectedEOF
	}

	if lrc.n >= lrc.limit {
		lrc.exceed = true
		return 0, io.ErrUnexpectedEOF
	}

	max := len(p)
	if int(lrc.limit-lrc.n) < max {
		max = int(lrc.limit - lrc.n)
	}

	n, err := lrc.rc.Read(p[:max])
	lrc.n += int64(n)

	if lrc.n >= lrc.limit {
		lrc.exceed = true
		return n, io.ErrUnexpectedEOF
	}

	return n, err
}

func (lrc *limitedReadCloser) Close() error {
	return lrc.rc.Close()
}

func NewSizeLimits(cfg LimitsConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > cfg.MaxRequestSize {
			c.JSON(413, gin.H{"error": "Request too large"})
			c.Abort()
			return
		}

		if c.Request.Body != nil {
			c.Request.Body = &limitedReadCloser{
				rc:    c.Request.Body,
				limit: cfg.MaxFileSize,
			}
		}

		c.Next()
	}
}

