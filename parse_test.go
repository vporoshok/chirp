package chirp

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	router := chi.NewRouter()
	router.Put("/user/{id}/name", func(w http.ResponseWriter, r *http.Request) {
		var data struct {
			ID       uuid.UUID `path:"id"`
			Name     string    `json:"name"`
			Part     string    `query:"part"`
			Priority uint8     `json:"priority"`
			Null     string    `json:"-"`
			Hero     string
		}
		require.NoError(t, Parse(r, &data))
		assert.EqualValues(t, uuid.FromStringOrNil("6b245e15-5c88-438b-a170-d8f97460083a"), data.ID)
		assert.Equal(t, "John", data.Name)
		assert.Equal(t, "last", data.Part)
		assert.Empty(t, "", data.Null)
		assert.Equal(t, "Joker", data.Hero)
		assert.EqualValues(t, 5, data.Priority)
	})

	t.Run("json", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name": "John", "priority": 5, "Hero": "Joker"}`)
		req := httptest.NewRequest(http.MethodPut, "/user/6b245e15-5c88-438b-a170-d8f97460083a/name?part=last", body)
		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)
	})

	t.Run("form", func(t *testing.T) {
		form := &url.Values{}
		form.Add("name", "John")
		form.Add("priority", "5")
		form.Add("Hero", "Joker")
		body := bytes.NewBufferString(form.Encode())
		req := httptest.NewRequest(http.MethodPut, "/user/6b245e15-5c88-438b-a170-d8f97460083a/name?part=last", body)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)
	})
}

func BenchmarkParse(b *testing.B) {
	router := chi.NewRouter()
	type Data struct {
		ID       uuid.UUID `path:"id"`
		Name     string    `json:"name"`
		Part     string    `query:"part"`
		Priority uint8     `json:"priority"`
		Null     string    `json:"-"`
		Hero     string
	}
	router.Put("/user/{id}/name", func(w http.ResponseWriter, r *http.Request) {
		data := Data{}
		Parse(r, &data)
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
