package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"movies.samkha.net/internal/validator"
)

func (app *application) readIdParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)

	if err != nil {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

type envelope map[string]any

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// to format the raw json
	// js, err := json.MarshalIndent(data, "", "\t") // impact performance better Marshal, and better than Marshal json.Encoder

	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)

	if err != nil {

		var syntaxError *json.SyntaxError
		var unMarshallTypeError *json.UnmarshalTypeError
		var invalidUnMarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly formatted JSON")

		case errors.As(err, &unMarshallTypeError):
			if unMarshallTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unMarshallTypeError.Field)
			}

			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unMarshallTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json:unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnMarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})

	if !errors.Is(err, io.EOF) {
		return errors.New("body must contain a single json value")
	}

	return nil
}

// reads string from query string
func (app *application) readString(qs url.Values, key, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

// reads string and split them into slice on comma character
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

// reads int from query string
// validator to record error if value can't be converted into int
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(s)

	if err != nil {
		v.AddError(key, "must be an integer value.")
		return defaultValue
	}

	return i
}

// helper to run background functions
func (app *application) background(fn func()) {
	go func() {
		defer func() {

			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}

		}()

		fn()
	}()
}
