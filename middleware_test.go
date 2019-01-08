package chirp

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-chi/chi"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareOK(t *testing.T) {
	router := chi.NewRouter()
	type Data struct {
		ID       uuid.UUID `path:"id"`
		Name     string    `json:"name"`
		Part     string    `query:"part"`
		Priority uint8     `json:"priority"`
		Null     string    `json:"-"`
		Hero     string
	}
	router.With(
		Middleware(Data{}),
	).Put("/user/{id}/name", func(w http.ResponseWriter, r *http.Request) {
		data := RequestFromContext(r.Context()).(Data)
		assert.EqualValues(t, uuid.FromStringOrNil("6b245e15-5c88-438b-a170-d8f97460083a"), data.ID)
		assert.Equal(t, "John", data.Name)
		assert.Equal(t, "last", data.Part)
		assert.Empty(t, "", data.Null)
		assert.Equal(t, "Joker", data.Hero)
		assert.EqualValues(t, 5, data.Priority)
	})

	body := bytes.NewBufferString(`{"name": "John", "priority": 5, "Hero": "Joker"}`)
	req := httptest.NewRequest(http.MethodPut, "/user/6b245e15-5c88-438b-a170-d8f97460083a/name?part=last", body)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)
}

func TestMiddlewareFail(t *testing.T) {
	logbuf := bytes.NewBuffer(nil)
	logger := log.New(logbuf, "", 0)
	router := chi.NewRouter()
	type Data struct {
		ID       uuid.UUID `path:"id"`
		Name     string    `json:"name"`
		Part     string    `query:"part"`
		Priority uint8     `json:"priority"`
		Null     string    `json:"-"`
		Hero     string
	}
	router.With(
		Middleware(Data{}, WithLogger(logger), WithInterrupt(400), WithOnError(
			func(err *Error, w http.ResponseWriter, req *http.Request) bool {
				assert.Equal(t, "id", err.Tag())
				assert.Equal(t, "path", err.Part())
				assert.Equal(t, "bad-uuid", err.Source())
				return false
			},
		)),
	).Put("/user/{id}/name", func(w http.ResponseWriter, r *http.Request) {
		require.Fail(t, "request should be interrupted")
	})

	body := bytes.NewBufferString(`{"name": "John", "priority": 5, "Hero": "Joker"}`)
	req := httptest.NewRequest(http.MethodPut, "/user/bad-uuid/name?part=last", body)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	assert.Equal(t, 400, res.Code)
	assert.Equal(t, "path[id](bad-uuid): uuid: incorrect UUID length: bad-uuid\n", res.Body.String())
	assert.Contains(t, logbuf.String(), "path[id](bad-uuid): uuid: incorrect UUID length: bad-uuid\n")
}

func TestMiddlewarePanic(t *testing.T) {
	require.Panics(t, func() {
		Middleware(&struct{}{})
	})
}

func BenchmarkMiddleware(b *testing.B) {
	router := chi.NewRouter()
	type Data struct {
		ID       uuid.UUID `path:"id"`
		Name     string    `json:"name"`
		Part     string    `query:"part"`
		Priority uint8     `json:"priority"`
		Null     string    `json:"-"`
		Hero     string
	}
	router.With(
		Middleware(Data{}),
	).Put("/user/{id}/name", func(w http.ResponseWriter, r *http.Request) {
		data := RequestFromContext(r.Context()).(Data)
		_ = data
	})

	body := bytes.NewBufferString(`{"name": "John", "priority": 5, "Hero": "Joker"}`)
	req := httptest.NewRequest(http.MethodPut, "/user/6b245e15-5c88-438b-a170-d8f97460083a/name?part=last", body)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)
	}
}
