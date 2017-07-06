package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/Pterodactyl/wings/config"
	"github.com/Pterodactyl/wings/control"
)

const configFile = "_testdata/config.json"

func TestAuthHandler(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	t.Run("rejects missing token", func(t *testing.T) {
		loadConfiguration(t, false)

		responded, rec := requestMiddlewareWith("c:somepermission", "", "")

		assert.False(t, responded)
		assert.Equal(t, rec.Code, http.StatusBadRequest)
	})

	t.Run("rejects c:* with invalid key", func(t *testing.T) {
		loadConfiguration(t, false)

		responded, rec := requestMiddlewareWith("c:somepermission", "invalidkey", "")

		assert.False(t, responded)
		assert.Equal(t, rec.Code, http.StatusForbidden)
	})

	t.Run("accepts existing c: key", func(t *testing.T) {
		loadConfiguration(t, false)

		responded, rec := requestMiddlewareWith("c:somepermission", "existingkey", "") // TODO: working token

		assert.True(t, responded)
		assert.Equal(t, rec.Code, http.StatusOK)
	})

	t.Run("rejects missing server uuid", func(t *testing.T) {
		loadConfiguration(t, true)

		responded, rec := requestMiddlewareWith("g:test", "existingkey", "")

		assert.False(t, responded)
		assert.Equal(t, rec.Code, http.StatusBadRequest)
	})

	t.Run("rejects not existing server", func(t *testing.T) {
		loadConfiguration(t, true)

		responded, rec := requestMiddlewareWith("g:test", "existingkey", "notexistingserver")

		assert.False(t, responded)
		assert.Equal(t, rec.Code, http.StatusNotFound)
	})

	t.Run("accepts server with existing g: key", func(t *testing.T) {
		loadConfiguration(t, true)

		responded, rec := requestMiddlewareWith("g:test", "existingkey", "existingserver")

		assert.True(t, responded)
		assert.Equal(t, rec.Code, http.StatusOK)
	})

	t.Run("rejects server with not existing g: key", func(t *testing.T) {
		loadConfiguration(t, true)

		responded, rec := requestMiddlewareWith("g:test", "notexistingkey", "existingserver")

		assert.False(t, responded)
		assert.Equal(t, rec.Code, http.StatusForbidden)
	})

	t.Run("rejects server with not existing s: key", func(t *testing.T) {
		loadConfiguration(t, true)

		responded, rec := requestMiddlewareWith("s:test", "notexistingskey", "existingserver")

		assert.False(t, responded)
		assert.Equal(t, rec.Code, http.StatusForbidden)
	})

	t.Run("accepts server with existing s: key with specific permission", func(t *testing.T) {
		loadConfiguration(t, true)

		responded, rec := requestMiddlewareWith("s:test", "existingspecificskey", "existingserver")

		assert.True(t, responded)
		assert.Equal(t, rec.Code, http.StatusOK)
	})

	t.Run("accepts server with existing s: key with gloabl permission", func(t *testing.T) {
		loadConfiguration(t, true)

		responded, rec := requestMiddlewareWith("s:test", "existingglobalskey", "existingserver")

		assert.True(t, responded)
		assert.Equal(t, rec.Code, http.StatusOK)
	})

	t.Run("rejects server with existing s: key without permission", func(t *testing.T) {
		loadConfiguration(t, true)

		responded, rec := requestMiddlewareWith("s:without", "existingspecificskey", "existingserver")

		assert.False(t, responded)
		assert.Equal(t, rec.Code, http.StatusForbidden)
	})
}

func requestMiddlewareWith(neededPermission string, token string, serverUUID string) (responded bool, recorder *httptest.ResponseRecorder) {
	router := gin.New()
	responded = false
	recorder = httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	router.GET("/", AuthHandler(neededPermission), func(c *gin.Context) {
		c.String(http.StatusOK, "Access granted.")
		responded = true
	})

	req.Header.Set(accessTokenHeader, token)
	req.Header.Set(accessServerHeader, serverUUID)
	router.ServeHTTP(recorder, req)
	return
}

func loadConfiguration(t *testing.T, serverConfig bool) {
	if err := config.LoadConfiguration(configFile); err != nil {
		t.Error(err)
		return
	}

	if serverConfig {
		if err := control.LoadServerConfigurations("_testdata/servers/"); err != nil {
			t.Error(err)
		}
	}
}