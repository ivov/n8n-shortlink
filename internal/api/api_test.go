package api_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ivov/n8n-shortlink/internal/api"
	"github.com/ivov/n8n-shortlink/internal/config"
	"github.com/ivov/n8n-shortlink/internal/db"
	"github.com/ivov/n8n-shortlink/internal/db/entities"
	"github.com/ivov/n8n-shortlink/internal/errors"
	"github.com/ivov/n8n-shortlink/internal/log"
	"github.com/ivov/n8n-shortlink/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPI(t *testing.T) {
	// ------------------------
	//         setup
	// ------------------------

	cfg := config.Config{
		DB:  struct{ FilePath string }{FilePath: ":memory:"},
		Env: "testing",
	}

	config.SetupDotDir()

	logger, err := log.NewLogger(cfg.Env)
	require.NoError(t, err)

	dbConn, err := db.SetupTestDB()
	require.NoError(t, err)
	defer dbConn.Close()

	api := &api.API{
		Config:           &cfg,
		Logger:           &logger,
		ShortlinkService: &services.ShortlinkService{DB: dbConn, Logger: &logger},
		VisitService:     &services.VisitService{DB: dbConn, Logger: &logger},
	}

	api.InitMetrics("test-commit-sha")

	server := httptest.NewServer(api.Routes())
	defer server.Close()

	// ------------------------
	//         utils
	// ------------------------

	storeShortlink := func(candidate entities.Shortlink) entities.Shortlink {
		body, err := json.Marshal(candidate)
		require.NoError(t, err)

		resp, err := http.Post(server.URL+"/shortlink", "application/json", bytes.NewBuffer(body))

		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response struct {
			Data json.RawMessage `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		var result entities.Shortlink
		err = json.Unmarshal(response.Data, &result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Slug)

		return result
	}

	noFollowRedirectClient := http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse // do not follow redirect so we can inspect it
		},
	}

	type ErrorResponse struct {
		Error struct {
			Message string `json:"message"`
			Doc     string `json:"doc"`
			Code    string `json:"code"`
			Trace   string `json:"trace"`
		} `json:"error"`
	}

	toErrorResponse := func(body io.ReadCloser) ErrorResponse {
		var errorResponse ErrorResponse
		err := json.NewDecoder(body).Decode(&errorResponse)
		require.NoError(t, err)

		return errorResponse
	}

	assertChallengeShown := func(resp *http.Response) {
		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		bodyString := string(bodyBytes)

		assert.Contains(t, bodyString, "<html")
		assert.Contains(t, bodyString, "Password required")
	}

	// ------------------------
	//         debug
	// ------------------------

	t.Run("debug", func(t *testing.T) {
		t.Run("should report health status", func(t *testing.T) {
			resp, err := http.Get(server.URL + "/health")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var health struct {
				Status      string `json:"status"`
				Environment string `json:"environment"`
				Version     string `json:"version"`
			}
			err = json.NewDecoder(resp.Body).Decode(&health)
			require.NoError(t, err)

			assert.Equal(t, "ok", health.Status)
			assert.Equal(t, "testing", health.Environment)
		})

		t.Run("should expose expvars metrics", func(t *testing.T) {
			resp, err := http.Get(server.URL + "/debug/vars")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var vars map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&vars)
			require.NoError(t, err)

			assert.Contains(t, vars, "uptime_seconds")
			uptime, ok := vars["uptime_seconds"].(float64)
			assert.True(t, ok, "uptime_seconds should be a float64")
			assert.GreaterOrEqual(t, uptime, float64(0), "uptime should be non-negative")

			assert.Contains(t, vars, "commit_sha")
			commitSha, ok := vars["commit_sha"].(string)
			assert.True(t, ok, "commit_sha should be a string")
			assert.NotEmpty(t, commitSha, "commit_sha should not be empty")

			assert.Contains(t, vars, "timestamp")
			timestamp, ok := vars["timestamp"].(float64)
			assert.True(t, ok, "timestamp should be a float64")
			assert.Greater(t, timestamp, float64(0), "timestamp should be positive")

			assert.Contains(t, vars, "goroutines")
			goroutines, ok := vars["goroutines"].(float64)
			assert.True(t, ok, "goroutines should be a float64")
			assert.Greater(t, goroutines, float64(0), "goroutines should be positive")

			assert.Contains(t, vars, "uptime_seconds")
			uptime, ok = vars["uptime_seconds"].(float64)
			assert.True(t, ok, "uptime_seconds should be a float64")
			assert.GreaterOrEqual(t, uptime, float64(0), "uptime should be non-negative")

			assert.Contains(t, vars, "cmdline")
			assert.IsType(t, []interface{}{}, vars["cmdline"])

			assert.Contains(t, vars, "memstats")
			assert.IsType(t, map[string]interface{}{}, vars["memstats"])

			assert.Contains(t, vars, "total_requests_received")
			assert.IsType(t, float64(0), vars["total_requests_received"])

			memstats, ok := vars["memstats"].(map[string]interface{})
			require.True(t, ok, "memstats should be a map")

			assert.Contains(t, memstats, "Alloc")
			assert.IsType(t, float64(0), memstats["Alloc"])

			assert.Contains(t, memstats, "TotalAlloc")
			assert.IsType(t, float64(0), memstats["TotalAlloc"])

			assert.Contains(t, memstats, "Sys")
			assert.IsType(t, float64(0), memstats["Sys"])

			assert.Contains(t, vars, "total_responses_sent_by_status")
			assert.IsType(t, map[string]interface{}{}, vars["total_responses_sent_by_status"])

			assert.Contains(t, vars, "in_flight_requests")
			assert.IsType(t, float64(0), vars["in_flight_requests"])

			assert.Contains(t, vars, "total_processing_time_ms")
			assert.IsType(t, float64(0), vars["total_processing_time_ms"])

			assert.Contains(t, vars, "total_responses_sent")
			assert.IsType(t, float64(0), vars["total_responses_sent"])
		})

		t.Run("should expose Prometheus metrics", func(t *testing.T) {
			resp, err := http.Get(server.URL + "/metrics")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Contains(t, string(body), "total_requests_received")
			assert.Contains(t, string(body), "total_responses_sent")
			assert.Contains(t, string(body), "in_flight_requests")
			assert.Contains(t, string(body), "total_processing_time_ms")
			assert.Contains(t, string(body), "total_responses_sent_by_status")
		})
	})

	// ------------------------
	//      base use cases
	// ------------------------

	t.Run("base use cases", func(t *testing.T) {
		t.Run("should create URL shortlink and redirect on retrieval", func(t *testing.T) {
			candidate := entities.Shortlink{
				Kind:    "url",
				Content: "https://example.com",
			}

			result := storeShortlink(candidate)

			assert.Equal(t, candidate.Kind, result.Kind)
			assert.Equal(t, candidate.Content, result.Content)

			resp, err := noFollowRedirectClient.Get(server.URL + "/" + result.Slug)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMovedPermanently, resp.StatusCode)
			assert.Equal(t, candidate.Content, resp.Header.Get("Location"))
		})

		t.Run("should create workflow shortlink and serve on retrieval", func(t *testing.T) {
			candidate := entities.Shortlink{
				Kind:    "workflow",
				Content: `{"nodes":[{"type":"n8n-nodes-base.start","typeVersion":1,"position":[250,300]}]}`,
			}

			result := storeShortlink(candidate)

			assert.Equal(t, candidate.Kind, result.Kind)
			assert.Equal(t, candidate.Content, result.Content)

			resp, err := http.Get(server.URL + "/" + result.Slug)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

			workflowStr, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// JSON -> map
			var workflowMap map[string]interface{}
			err = json.Unmarshal([]byte(workflowStr), &workflowMap)
			require.NoError(t, err)

			// validate content
			assert.Contains(t, workflowMap, "nodes")
			nodes, ok := workflowMap["nodes"].([]interface{})
			assert.True(t, ok)
			assert.Len(t, nodes, 1)
			node := nodes[0].(map[string]interface{})
			assert.Equal(t, "n8n-nodes-base.start", node["type"])
			assert.Equal(t, float64(1), node["typeVersion"])
			position, ok := node["position"].([]interface{})
			assert.True(t, ok)
			assert.Equal(t, []interface{}{float64(250), float64(300)}, position)
		})

		t.Run("should return 404 header + page on retrieval of inexistent slug", func(t *testing.T) {
			resp, err := http.Get(server.URL + "/" + "inexistent")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusNotFound, resp.StatusCode)

			bodyBytes, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			bodyString := string(bodyBytes)

			assert.Contains(t, bodyString, "Page not found")
		})

		t.Run("should record visit on retrieval", func(t *testing.T) {
			candidate := entities.Shortlink{
				Kind:    "url",
				Content: "https://example.com/visit-test",
			}

			result := storeShortlink(candidate)

			const referer = "https://test-referer.com"
			const userAgent = "TestUserAgent/1.0"
			req, err := http.NewRequest("GET", server.URL+"/"+result.Slug, nil)
			require.NoError(t, err)
			req.Header.Set("Referer", referer)
			req.Header.Set("User-Agent", userAgent)

			resp, err := noFollowRedirectClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMovedPermanently, resp.StatusCode)

			var visit entities.Visit
			err = dbConn.Get(&visit, "SELECT * FROM visits WHERE slug = ? ORDER BY ts DESC LIMIT 1;", result.Slug)
			require.NoError(t, err)

			assert.Equal(t, result.Slug, visit.Slug)
			assert.Equal(t, referer, visit.Referer)
			assert.Equal(t, userAgent, visit.UserAgent)
			assert.NotZero(t, visit.TS)
		})
	})

	// ------------------------
	//      custom slug
	// ------------------------

	t.Run("custom slug", func(t *testing.T) {
		t.Run("should create custom-slug shortlink and redirect on retrieval", func(t *testing.T) {
			candidate := entities.Shortlink{
				Slug:    "my-custom-slug",
				Kind:    "url",
				Content: "https://example.org",
			}

			result := storeShortlink(candidate)

			assert.Equal(t, candidate.Slug, result.Slug)
			assert.Equal(t, candidate.Kind, result.Kind)
			assert.Equal(t, candidate.Content, result.Content)

			resp, err := noFollowRedirectClient.Get(server.URL + "/" + result.Slug)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMovedPermanently, resp.StatusCode)
			assert.Equal(t, candidate.Content, resp.Header.Get("Location"))
		})
	})

	// ------------------------
	//   creation validation
	// ------------------------

	t.Run("creation payload validation", func(t *testing.T) {
		t.Run("should reject on invalid creation payload", func(t *testing.T) {
			testCases := []struct {
				name               string
				shortlink          entities.Shortlink
				expectedStatusCode int
				expectedErrorCode  string
			}{
				{
					name:               "Content as empty string",
					shortlink:          entities.Shortlink{}, // content defaults to empty string
					expectedStatusCode: http.StatusBadRequest,
					expectedErrorCode:  errors.ToCode[errors.ErrContentMalformed],
				},
				{
					name:              "Content as neither URL nor JSON",
					shortlink:         entities.Shortlink{Content: "not-a-url"},
					expectedErrorCode: errors.ToCode[errors.ErrContentMalformed],
				},
				{
					name:              "Password too short",
					shortlink:         entities.Shortlink{Content: "https://example.com", Password: "1234567"},
					expectedErrorCode: errors.ToCode[errors.ErrPasswordTooShort],
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					body, err := json.Marshal(tc.shortlink)
					require.NoError(t, err)

					resp, err := http.Post(server.URL+"/shortlink", "application/json", bytes.NewBuffer(body))
					require.NoError(t, err)
					defer resp.Body.Close()

					errorResponse := toErrorResponse(resp.Body)

					assert.Equal(t, errorResponse.Error.Code, tc.expectedErrorCode)
				})
			}
		})

		t.Run("should reject on duplicate custom slug in payload", func(t *testing.T) {
			candidate := entities.Shortlink{
				Slug:    "some-custom-slug",
				Kind:    "url",
				Content: "https://example.com",
			}

			result := storeShortlink(candidate)

			duplicate := entities.Shortlink{
				Slug:    result.Slug, // already exists
				Content: "https://other-example.com",
			}

			body, err := json.Marshal(duplicate)
			require.NoError(t, err)
			resp, err := http.Post(server.URL+"/shortlink", "application/json", bytes.NewBuffer(body))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

			errorResponse := toErrorResponse(resp.Body)
			assert.Equal(t, errors.ToCode[errors.ErrSlugTaken], errorResponse.Error.Code)

			// verify that original shortlink is still intact
			resp, err = noFollowRedirectClient.Get(server.URL + "/" + candidate.Slug)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMovedPermanently, resp.StatusCode)
			assert.Equal(t, "https://example.com", resp.Header.Get("Location"))
		})

		t.Run("should reject on invalid custom slug", func(t *testing.T) {
			testCases := []struct {
				errorCode string
				slug      string
			}{
				{errors.ToCode[errors.ErrSlugTooShort], "abc"},
				{errors.ToCode[errors.ErrSlugTooLong], strings.Repeat("a", 513)},
				{errors.ToCode[errors.ErrSlugMisformatted], "abc+def"},
				{errors.ToCode[errors.ErrSlugReserved], "health"},
			}

			for _, tc := range testCases {
				shortlink := entities.Shortlink{
					Slug:    tc.slug,
					Content: "https://example.com",
				}

				body, _ := json.Marshal(shortlink)
				resp, err := http.Post(server.URL+"/shortlink", "application/json", bytes.NewBuffer(body))
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

				errorResponse := toErrorResponse(resp.Body)

				assert.Equal(t, errorResponse.Error.Code, tc.errorCode)
			}
		})

		t.Run("should reject payload >= 5 MB", func(t *testing.T) {
			tooBig := strings.Repeat("a", 5*1024*1024) // 5 MB
			tooBigCandidate := entities.Shortlink{
				Content: "https://example.com?" + tooBig,
			}

			body, err := json.Marshal(tooBigCandidate)
			require.NoError(t, err)

			resp, err := http.Post(server.URL+"/shortlink", "application/json", bytes.NewBuffer(body))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

			errorResponse := toErrorResponse(resp.Body)
			assert.Equal(t, errorResponse.Error.Code, errors.ToCode[errors.ErrPayloadTooLarge])

			// retry with smaller payload

			rightSized := strings.Repeat("a", 4*1024*1024) // 4 MB
			rightSizedCandidate := entities.Shortlink{
				Kind:    "url",
				Content: "https://example.com?" + rightSized,
			}

			result := storeShortlink(rightSizedCandidate)

			assert.NotEmpty(t, result.Slug)
			assert.Equal(t, rightSizedCandidate.Kind, result.Kind)
			assert.Equal(t, rightSizedCandidate.Content, result.Content)
		})
	})

	t.Run("rate limiting", func(t *testing.T) {
		t.Run("should enforce rate limiting", func(t *testing.T) {
			// enable rate limiting only for this test
			originalConfig := *api.Config
			api.Config.RateLimiter.Enabled = true
			api.Config.RateLimiter.RPS = 2
			api.Config.RateLimiter.Burst = 2
			defer func() {
				api.Config = &originalConfig
			}()

			shortlink := entities.Shortlink{
				Content: "https://example.com",
			}
			body, err := json.Marshal(shortlink)
			require.NoError(t, err)

			// util to make a request
			makeRequest := func() (*http.Response, error) {
				return http.Post(server.URL+"/shortlink", "application/json", bytes.NewBuffer(body))
			}

			// make requests up to the limit
			for i := 0; i < api.Config.RateLimiter.Burst; i++ {
				resp, err := makeRequest()
				require.NoError(t, err)
				defer resp.Body.Close()
				assert.Equal(t, http.StatusCreated, resp.StatusCode)
			}

			resp, err := makeRequest() // should be rate limited
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

			errorResponse := toErrorResponse(resp.Body)

			assert.Equal(t, "You have exceeded the rate limit. Please wait and retry later.", errorResponse.Error.Message)

			time.Sleep(time.Second) // wait for rate limit to reset

			resp, err = makeRequest() // should succeed
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		})
	})

	t.Run("password protection", func(t *testing.T) {
		t.Run("should store password-protected shortlink", func(t *testing.T) {
			candidate := entities.Shortlink{
				Kind:     "url",
				Content:  "https://example.com",
				Password: "securepass123",
			}

			result := storeShortlink(candidate)

			var storedPassword string
			err = dbConn.Get(&storedPassword, "SELECT password FROM shortlinks WHERE slug = ?", result.Slug)
			require.NoError(t, err)
			assert.NotEmpty(t, storedPassword)
			assert.NotContains(t, storedPassword, "securepass123") // has been hashed

			assert.NotEmpty(t, result.Slug)
			assert.Equal(t, candidate.Kind, result.Kind)
			assert.Equal(t, candidate.Content, result.Content)
			assert.Empty(t, result.Password) // not returned in response
		})

		t.Run("should show challenge for password-protected shortlink", func(t *testing.T) {
			result := storeShortlink(entities.Shortlink{
				Kind:     "url",
				Content:  "https://example.com/protected",
				Password: "securepass123",
			})

			resp, err := http.Get(server.URL + "/" + result.Slug)
			require.NoError(t, err)
			defer resp.Body.Close()

			assertChallengeShown(resp)
		})

		t.Run("should return original URL if correct password", func(t *testing.T) {
			plainPassword := "securepass123"
			candidate := entities.Shortlink{
				Kind:     "url",
				Content:  "https://example.com/protected",
				Password: plainPassword,
			}
			result := storeShortlink(candidate)

			req, err := http.NewRequest("GET", server.URL+"/"+result.Slug, nil)
			require.NoError(t, err)

			auth := base64.StdEncoding.EncodeToString([]byte(plainPassword))
			req.Header.Add("Authorization", "Basic "+auth)

			resp, err := noFollowRedirectClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			bodyString := string(bodyBytes)

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
			assert.Equal(t, "{\"url\":\"https://example.com/protected\"}\n", bodyString)
		})

		t.Run("should deny access with incorrect password", func(t *testing.T) {
			candidate := entities.Shortlink{
				Kind:     "url",
				Content:  "https://example.com/protected",
				Password: "securepass123",
			}
			result := storeShortlink(candidate)

			req, err := http.NewRequest("GET", server.URL+"/"+result.Slug, nil)
			require.NoError(t, err)

			auth := base64.StdEncoding.EncodeToString([]byte("wrongpass"))
			req.Header.Add("Authorization", "Basic "+auth)

			resp, err := noFollowRedirectClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})

		t.Run("should show challenge on missing Authorization header", func(t *testing.T) {
			candidate := entities.Shortlink{
				Kind:     "url",
				Content:  "https://example.com/protected",
				Password: "securepass123",
			}
			result := storeShortlink(candidate)

			req, err := http.NewRequest("GET", server.URL+"/"+result.Slug, nil)
			require.NoError(t, err)

			resp, err := noFollowRedirectClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assertChallengeShown(resp)
		})

		t.Run("should reject on malformed Authorization header", func(t *testing.T) {
			candidate := entities.Shortlink{
				Kind:     "url",
				Content:  "https://example.com/protected",
				Password: "securepass123",
			}
			result := storeShortlink(candidate)

			req, err := http.NewRequest("GET", server.URL+"/"+result.Slug, nil)
			require.NoError(t, err)

			auth := base64.StdEncoding.EncodeToString([]byte("wrongpass"))
			req.Header.Add("Authorization", "B_a_s_i_c "+auth)

			resp, err := noFollowRedirectClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})

		t.Run("should reject on missing password in Authorization header", func(t *testing.T) {
			candidate := entities.Shortlink{
				Kind:     "url",
				Content:  "https://example.com/protected",
				Password: "securepass123",
			}
			result := storeShortlink(candidate)

			req, err := http.NewRequest("GET", server.URL+"/"+result.Slug, nil)
			require.NoError(t, err)

			auth := base64.StdEncoding.EncodeToString([]byte(""))
			req.Header.Add("Authorization", "B_a_s_i_c "+auth)

			resp, err := noFollowRedirectClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	})

	t.Run("content validation", func(t *testing.T) {
		testCases := []struct {
			name        string
			content     string
			shouldBlock bool
		}{
			{"legitimate URL", "https://example.com", false},
			{"n8n workflow", "https://n8n.io/workflows/123", false},
			{"github repo", "https://github.com/user/repo", false},

			{"cpanel phishing", "https://oauth-us-est-25.178-128-96-243.cpanel.site/?access", true},
			{"payment scam", "https://depop.order-payment2232321.cyou/245918330", true},
			{"screenconnect", "https://digslhrizxde.screenconnect.com/Bin/ScreenConnect.ClientSetup.exe", true},
			{"signin phishing", "https://evil.com/signin-portal", true},
			{"delivery scam", "https://fake-fedex.com/delivery-notification", true},

			{"workflow with signin", `{"nodes":[{"name":"signin-node","type":"webhook"}]}`, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				shortlink := entities.Shortlink{Content: tc.content}

				body, err := json.Marshal(shortlink)
				require.NoError(t, err)

				resp, err := http.Post(server.URL+"/shortlink", "application/json", bytes.NewBuffer(body))
				require.NoError(t, err)
				defer resp.Body.Close()

				if tc.shouldBlock {
					assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

					var errorResponse ErrorResponse
					err = json.NewDecoder(resp.Body).Decode(&errorResponse)
					require.NoError(t, err)
					assert.Equal(t, "CONTENT_BLOCKED", errorResponse.Error.Code)
				} else {
					assert.Equal(t, http.StatusCreated, resp.StatusCode)
				}
			})
		}
	})
}
