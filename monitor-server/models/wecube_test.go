package models

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type coreTokenTestTransport func(*http.Request) (*http.Response, error)

func (f coreTokenTestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func prepareCoreTokenTest(t *testing.T) string {
	t.Helper()
	oldURL, oldCode, oldKey := CoreUrl, SubSystemCode, SubSystemKey
	oldAccess, oldRefresh := coreRequestToken, coreRefreshToken
	oldAccessExpiry, oldRefreshExpiry := coreRequestTokenExpTime, coreRefreshTokenExpTime
	oldClient := coreTokenHTTPClient
	client := *oldClient
	coreTokenHTTPClient = &client
	t.Cleanup(func() {
		CoreUrl, SubSystemCode, SubSystemKey = oldURL, oldCode, oldKey
		coreRequestToken, coreRefreshToken = oldAccess, oldRefresh
		coreRequestTokenExpTime, coreRefreshTokenExpTime = oldAccessExpiry, oldRefreshExpiry
		coreTokenHTTPClient = oldClient
	})
	CoreUrl, SubSystemCode = "http://core.example.invalid", "subsystem-a"
	coreRequestToken, coreRefreshToken = "", ""
	coreRequestTokenExpTime, coreRefreshTokenExpTime = time.Time{}, time.Time{}
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	SubSystemKey = base64.StdEncoding.EncodeToString(raw)
	return SubSystemKey
}

func tokenTestResponse() *http.Response {
	body := fmt.Sprintf(`{"status":"OK","data":[{"tokenType":"accessToken","token":"access-placeholder","expiration":"%d"},{"tokenType":"refreshToken","token":"refresh-placeholder","expiration":"%d"}]}`, time.Now().Add(time.Hour).Unix(), time.Now().Add(2*time.Hour).Unix())
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}
}

func TestCoreTokenRequestsHaveDeadlines(t *testing.T) {
	key := prepareCoreTokenTest(t)
	coreTokenHTTPClient.Transport = coreTokenTestTransport(func(req *http.Request) (*http.Response, error) {
		if _, ok := req.Context().Deadline(); !ok {
			t.Errorf("Core token request has no deadline: %s", req.URL.Path)
		}
		return tokenTestResponse(), nil
	})
	if err := requestCoreToken(context.Background(), key); err != nil {
		t.Fatal("Core login failed")
	}
	if err := refreshToken(context.Background()); err != nil {
		t.Fatal("Core token refresh failed")
	}
}

func TestGetCoreTokenReturnsAccessTokenAfterLogin(t *testing.T) {
	prepareCoreTokenTest(t)
	coreTokenHTTPClient.Transport = coreTokenTestTransport(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.Path != "/auth/v1/api/login" {
			t.Fatal("expected Core login request")
		}
		return tokenTestResponse(), nil
	})
	if GetCoreToken() != "Bearer access-placeholder" {
		t.Fatal("Core login did not return the access token")
	}
}

func TestGetCoreTokenContextRejectsInvalidResponses(t *testing.T) {
	validExpiration := fmt.Sprint(time.Now().Add(time.Hour).Unix())
	for _, operation := range []string{"login", "refresh"} {
		for _, tc := range []struct {
			name, status, token, expiration string
			httpStatus                      int
		}{
			{"http", "OK", "access-placeholder", validExpiration, http.StatusInternalServerError},
			{"business", "ERROR", "access-placeholder", validExpiration, http.StatusOK},
			{"missing_access", "OK", "access-placeholder", validExpiration, http.StatusOK},
			{"missing_token", "OK", "", validExpiration, http.StatusOK},
			{"blank_token", "OK", " ", validExpiration, http.StatusOK},
			{"expired", "OK", "access-placeholder", "1", http.StatusOK},
			{"invalid_expiration", "OK", "access-placeholder", "invalid", http.StatusOK},
			{"invalid_expiration_suffix", "OK", "access-placeholder", validExpiration + "invalid", http.StatusOK},
		} {
			t.Run(operation+"/"+tc.name, func(t *testing.T) {
				prepareCoreTokenTest(t)
				coreRequestToken = "Bearer old-access-placeholder"
				coreRequestTokenExpTime = time.Unix(1, 0)
				if operation == "refresh" {
					coreRefreshToken = "Bearer old-refresh-placeholder"
					coreRefreshTokenExpTime = time.Now().Add(time.Hour)
				}
				oldRefresh, oldRefreshExpiry := coreRefreshToken, coreRefreshTokenExpTime
				coreTokenHTTPClient.Transport = coreTokenTestTransport(func(req *http.Request) (*http.Response, error) {
					body := fmt.Sprintf(`{"status":%q,"data":[{"tokenType":"accessToken","token":%q,"expiration":%q},{"tokenType":"refreshToken","token":"new-refresh-placeholder","expiration":%q}]}`, tc.status, tc.token, tc.expiration, validExpiration)
					if tc.name == "missing_access" {
						body = `{"status":"OK","data":[]}`
					}
					return &http.Response{StatusCode: tc.httpStatus, Body: io.NopCloser(strings.NewReader(body))}, nil
				})
				token, err := GetCoreTokenContext(context.Background())
				if err == nil || token != "" {
					t.Error("invalid Core response was accepted as a successful token acquisition")
				}
				if coreRequestToken != "Bearer old-access-placeholder" || !coreRequestTokenExpTime.Equal(time.Unix(1, 0)) || coreRefreshToken != oldRefresh || !coreRefreshTokenExpTime.Equal(oldRefreshExpiry) {
					t.Error("failed Core response changed the token cache")
				}
			})
		}
	}
}

func TestGetCoreTokenContextAcceptsAccessOnlyResponses(t *testing.T) {
	for _, milliseconds := range []bool{false, true} {
		t.Run(fmt.Sprintf("milliseconds=%t", milliseconds), func(t *testing.T) {
			prepareCoreTokenTest(t)
			expires := time.Now().Add(time.Hour)
			expiration := expires.Unix()
			if milliseconds {
				expiration = expires.UnixMilli()
			}
			coreTokenHTTPClient.Transport = coreTokenTestTransport(func(*http.Request) (*http.Response, error) {
				body := fmt.Sprintf(`{"status":"OK","data":[{"tokenType":"accessToken","token":"access-placeholder","expiration":"%d"}]}`, expiration)
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			token, err := GetCoreTokenContext(context.Background())
			if err != nil || token != "Bearer access-placeholder" {
				t.Fatal("valid access-only response was rejected")
			}
		})
	}
}

func TestGetCoreTokenSerializesConcurrentRequests(t *testing.T) {
	for _, operation := range []string{"login", "refresh"} {
		t.Run(operation, func(t *testing.T) {
			prepareCoreTokenTest(t)
			if operation == "refresh" {
				coreRefreshToken = "Bearer refresh-placeholder"
				coreRefreshTokenExpTime = time.Now().Add(time.Hour)
			}
			var requests int32
			release := make(chan struct{})
			firstRequest := make(chan struct{}, 1)
			coreTokenHTTPClient.Transport = coreTokenTestTransport(func(req *http.Request) (*http.Response, error) {
				atomic.AddInt32(&requests, 1)
				select {
				case firstRequest <- struct{}{}:
				default:
				}
				<-release
				return tokenTestResponse(), nil
			})
			const callers = 16
			start := make(chan struct{})
			results := make(chan string, callers)
			var ready sync.WaitGroup
			ready.Add(callers)
			for i := 0; i < callers; i++ {
				go func() {
					ready.Done()
					<-start
					results <- GetCoreToken()
				}()
			}
			ready.Wait()
			close(start)
			<-firstRequest
			// Keep the first response in flight while the other callers reach the
			// cache, so they must wait for the refresh rather than start their own.
			time.AfterFunc(20*time.Millisecond, func() { close(release) })
			for i := 0; i < callers; i++ {
				if token := <-results; token != "Bearer access-placeholder" {
					t.Error("concurrent caller did not receive the access token")
				}
			}
			if got := atomic.LoadInt32(&requests); got != 1 {
				t.Fatalf("concurrent callers made %d token requests; expected one", got)
			}
		})
	}
}

func TestGetCoreTokenContextCancelsWaitingCaller(t *testing.T) {
	prepareCoreTokenTest(t)
	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	coreTokenHTTPClient.Transport = coreTokenTestTransport(func(req *http.Request) (*http.Response, error) {
		entered <- struct{}{}
		<-release
		return tokenTestResponse(), nil
	})
	first := make(chan error, 1)
	go func() {
		_, err := GetCoreTokenContext(context.Background())
		first <- err
	}()
	<-entered
	ctx, cancel := context.WithCancel(context.Background())
	waiting := make(chan error, 1)
	go func() {
		_, err := GetCoreTokenContext(ctx)
		waiting <- err
	}()
	cancel()
	var waitErr error
	completed := false
	select {
	case waitErr = <-waiting:
		completed = true
	case <-time.After(500 * time.Millisecond):
		t.Error("waiting Core token caller ignored cancellation")
	}
	close(release)
	if err := <-first; err != nil {
		t.Fatal("initial Core token request failed")
	}
	if !completed {
		waitErr = <-waiting
	}
	if !errors.Is(waitErr, context.Canceled) {
		t.Fatalf("expected waiting caller cancellation, got %v", waitErr)
	}
}

func TestGetCoreTokenContextCancelsTokenRequest(t *testing.T) {
	for _, operation := range []string{"login", "refresh"} {
		t.Run(operation, func(t *testing.T) {
			prepareCoreTokenTest(t)
			if operation == "refresh" {
				coreRefreshToken = "Bearer refresh-placeholder"
				coreRefreshTokenExpTime = time.Now().Add(time.Hour)
			}
			entered := make(chan struct{}, 2)
			var requests int32
			coreTokenHTTPClient.Transport = coreTokenTestTransport(func(req *http.Request) (*http.Response, error) {
				atomic.AddInt32(&requests, 1)
				entered <- struct{}{}
				<-req.Context().Done()
				return nil, req.Context().Err()
			})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() {
				_, err := GetCoreTokenContext(ctx)
				done <- err
			}()
			<-entered
			cancel()
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("expected Core token request cancellation, got %v", err)
				}
			case <-time.After(500 * time.Millisecond):
				t.Fatal("Core token request ignored cancellation")
			}
			if got := atomic.LoadInt32(&requests); got != 1 {
				t.Fatalf("cancellation started %d token requests; expected one", got)
			}
		})
	}
}

func TestCoreTokenRequestsTimeOutOnStalledResponses(t *testing.T) {
	for _, operation := range []string{"login", "refresh"} {
		for _, stage := range []string{"headers", "body"} {
			t.Run(operation+"/"+stage, func(t *testing.T) {
				key := prepareCoreTokenTest(t)
				release := make(chan struct{})
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					if stage == "body" {
						w.Write([]byte(`{"status":`))
						w.(http.Flusher).Flush()
					}
					select {
					case <-req.Context().Done():
					case <-release:
					}
				}))
				defer server.Close()
				defer close(release)
				CoreUrl = server.URL
				coreTokenHTTPClient = &http.Client{Timeout: 25 * time.Millisecond}
				done := make(chan error, 1)
				go func() {
					if operation == "login" {
						done <- requestCoreToken(context.Background(), key)
					} else {
						done <- refreshToken(context.Background())
					}
				}()
				select {
				case err := <-done:
					if !errors.Is(err, context.DeadlineExceeded) {
						t.Fatal("stalled Core token response did not return a timeout")
					}
				case <-time.After(500 * time.Millisecond):
					t.Fatal("stalled Core token response blocked beyond the request timeout")
				}
			})
		}
	}
}
