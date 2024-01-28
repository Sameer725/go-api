package main

import "net/http"

type metricsResponseWrite struct {
	wrapped       http.ResponseWriter
	statusCode    int
	headerWritten bool
}

func newMetricsResponseWrite(w http.ResponseWriter) *metricsResponseWrite {
	return &metricsResponseWrite{
		wrapped:    w,
		statusCode: http.StatusOK,
	}
}

func (mw *metricsResponseWrite) Header() http.Header {
	return mw.wrapped.Header()
}

func (mw *metricsResponseWrite) WriteHeader(statusCode int) {
	mw.wrapped.WriteHeader(statusCode)

	if !mw.headerWritten {
		mw.statusCode = statusCode
		mw.headerWritten = true
	}
}

func (mw *metricsResponseWrite) Write(b []byte) (int, error) {
	mw.headerWritten = true

	return mw.wrapped.Write(b)
}

func (mw *metricsResponseWrite) Unwrap() http.ResponseWriter {
	return mw.wrapped
}
