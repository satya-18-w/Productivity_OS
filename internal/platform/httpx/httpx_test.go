package httpx

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func decodeError(t *testing.T, body io.Reader) errorPayload {
	t.Helper()
	var b errorBody
	require.NoError(t, json.NewDecoder(body).Decode(&b))
	return b.Error
}

func TestRecoverer_PanicBecomesGeneric500(t *testing.T) {
	const secret = "postgres://user:supersecret@db/internal"
	h := Chain(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic(secret)
		}),
		RequestIDMiddleware, Recoverer,
	)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	payload := decodeError(t, rec.Body)
	require.Equal(t, CodeInternal, payload.Code)
	require.NotContains(t, rec.Body.String(), "supersecret")
	require.NotContains(t, rec.Body.String(), "panic")
	require.NotEmpty(t, rec.Header().Get("X-Request-Id"))
}

func TestWriteError_APIErrorPassesThrough(t *testing.T) {
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/x", nil)
	WriteError(rec, r, ValidationError(map[string]string{"email": "email is required"}))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	payload := decodeError(t, rec.Body)
	require.Equal(t, CodeValidation, payload.Code)
	require.Equal(t, "email is required", payload.Fields["email"])
}

func TestWriteError_UnknownErrorIsScrubbed(t *testing.T) {
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	WriteError(rec, r, io.ErrUnexpectedEOF)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	payload := decodeError(t, rec.Body)
	require.Equal(t, CodeInternal, payload.Code)
	require.NotContains(t, rec.Body.String(), "EOF")
}

func TestDecodeJSON(t *testing.T) {
	type in struct {
		Name string `json:"name"`
	}

	cases := []struct {
		name string
		body string
		ok   bool
	}{
		{"valid", `{"name":"a"}`, true},
		{"malformed", `{"name":`, false},
		{"unknown field", `{"nope":1}`, false},
		{"trailing data", `{"name":"a"}{"name":"b"}`, false},
		{"too large", `{"name":"` + strings.Repeat("x", MaxBodyBytes+1) + `"}`, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(tc.body))
			err := DecodeJSON(rec, r, &in{})
			if tc.ok {
				require.NoError(t, err)
				return
			}
			var apiErr *APIError
			require.ErrorAs(t, err, &apiErr)
			require.Equal(t, http.StatusBadRequest, apiErr.Status)
		})
	}
}

func TestChain_OuterMostFirst(t *testing.T) {
	var order []string
	mw := func(tag string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, tag)
				next.ServeHTTP(w, r)
			})
		}
	}
	h := Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		order = append(order, "handler")
	}), mw("a"), mw("b"))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, []string{"a", "b", "handler"}, order)
}
