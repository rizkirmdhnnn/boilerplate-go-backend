package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testContext creates a Gin context with a fresh ResponseRecorder.
func testContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	return c, w
}

func TestOK(t *testing.T) {
	c, w := testContext()
	OK(c, map[string]string{"key": "value"})

	assert.Equal(t, 200, w.Code)
	var resp APIResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, map[string]interface{}{"key": "value"}, resp.Data)
}

func TestCreated(t *testing.T) {
	c, w := testContext()
	Created(c, map[string]int{"id": 1})

	assert.Equal(t, 201, w.Code)
	var resp APIResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
}

func TestMessage(t *testing.T) {
	c, w := testContext()
	Message(c, "hello")

	assert.Equal(t, 200, w.Code)
	var resp APIResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, "hello", resp.Message)
}

func TestError(t *testing.T) {
	c, w := testContext()
	Error(c, 400, "bad request")

	assert.Equal(t, 400, w.Code)
	var resp APIResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Success)
	assert.Equal(t, "bad request", resp.Error.Message)
}

func TestValidationError(t *testing.T) {
	c, w := testContext()
	ValidationError(c, "invalid field")

	assert.Equal(t, 422, w.Code)
	var resp APIResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Success)
}

func TestNotFound(t *testing.T) {
	c, w := testContext()
	NotFound(c, "")

	assert.Equal(t, 404, w.Code)
	var resp APIResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Success)
	assert.Equal(t, "resource not found", resp.Error.Message)
}

func TestInternalError(t *testing.T) {
	c, w := testContext()
	InternalError(c)

	assert.Equal(t, 500, w.Code)
	var resp APIResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Success)
}

func TestUnauthorized(t *testing.T) {
	c, w := testContext()
	Unauthorized(c, "")

	assert.Equal(t, 401, w.Code)
	var resp APIResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Success)
}

func TestForbidden(t *testing.T) {
	c, w := testContext()
	Forbidden(c, "no access")

	assert.Equal(t, 403, w.Code)
	var resp APIResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Success)
	assert.Equal(t, "no access", resp.Error.Message)
}

func TestPaginated(t *testing.T) {
	c, w := testContext()
	Paginated(c, []string{"a", "b"}, 1, 10, 2)

	assert.Equal(t, 200, w.Code)
	var resp APIResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, 1, resp.Meta.Page)
	assert.Equal(t, 10, resp.Meta.PerPage)
	assert.Equal(t, 2, resp.Meta.Total)
	assert.Equal(t, 1, resp.Meta.TotalPages)
}

func TestErrorDetail(t *testing.T) {
	c, w := testContext()
	ErrorDetail(c, 409, "CONFLICT", "duplicate", "email already in use")

	assert.Equal(t, 409, w.Code)
	var resp APIResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Success)
	assert.Equal(t, "CONFLICT", resp.Error.Code)
	assert.Equal(t, "duplicate", resp.Error.Message)
	assert.Equal(t, "email already in use", resp.Error.Details)
}
