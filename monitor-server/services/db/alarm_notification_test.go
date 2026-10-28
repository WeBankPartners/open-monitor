package db

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"xorm.io/xorm"
	"xorm.io/xorm/core"
)

type notificationDB struct {
	payload    string
	status     string
	commits    int
	begins     int
	updates    int
	newAlarm   bool
	bindErr    error
	alarms     map[string]int64
	updatedIDs []int64
}
type notificationConnector struct{ state *notificationDB }

func (c notificationConnector) Connect(context.Context) (driver.Conn, error) {
	return &notificationConn{state: c.state}, nil
}
func (c notificationConnector) Driver() driver.Driver { return notificationDriver{} }

type notificationDriver struct{}

func (notificationDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use connector") }

type notificationConn struct {
	state           *notificationDB
	payload, status string
}

func (*notificationConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (*notificationConn) Close() error { return nil }
func (c *notificationConn) Begin() (driver.Tx, error) {
	c.state.begins++
	c.payload, c.status = c.state.payload, c.state.status
	return c, nil
}
func (c *notificationConn) Commit() error { c.state.commits++; return nil }
func (c *notificationConn) Rollback() error {
	c.state.payload, c.state.status = c.payload, c.status
	return nil
}

type notificationRows struct {
	columns []string
	values  [][]driver.Value
}

func (r *notificationRows) Columns() []string { return r.columns }
func (*notificationRows) Close() error        { return nil }
func (r *notificationRows) Next(dest []driver.Value) error {
	if len(r.values) == 0 {
		return io.EOF
	}
	copy(dest, r.values[0])
	r.values = r.values[1:]
	return nil
}
func (c *notificationConn) QueryContext(_ context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
	switch {
	case strings.Contains(q, "FROM k8s_alert_outbox"):
		return &notificationRows{columns: []string{"payload"}, values: [][]driver.Value{{c.state.payload}}}, nil
	case strings.Contains(q, "SELECT id FROM alarm"):
		rows := &notificationRows{columns: []string{"id"}}
		if c.state.alarms != nil {
			if strings.Contains(q, "alarm_strategy=?") {
				if id := c.state.alarms[args[3].Value.(string)]; id != 0 {
					rows.values = [][]driver.Value{{id}}
				}
			}
		} else if c.state.status == "firing" {
			rows.values = [][]driver.Value{{int64(7)}}
		}
		return rows, nil
	case strings.Contains(q, "SELECT status FROM alarm"):
		return &notificationRows{columns: []string{"status"}, values: [][]driver.Value{{c.state.status}}}, nil
	default:
		return nil, errors.New("unexpected query: " + q)
	}
}
func (c *notificationConn) ExecContext(_ context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
	switch {
	case strings.HasPrefix(q, "insert into alarm_firing"):
		c.state.newAlarm = true
		return notificationResult(11), nil
	case strings.HasPrefix(q, "INSERT INTO alarm("):
		c.state.status = args[2].Value.(string)
		return notificationResult(7), nil
	case strings.HasPrefix(q, "update alarm_firing set alarm_id="):
		return driver.RowsAffected(1), nil
	case strings.HasPrefix(q, "UPDATE alarm SET status="):
		if c.state.status != "firing" {
			return driver.RowsAffected(0), nil
		}
		c.state.status = args[0].Value.(string)
		c.state.updates++
		c.state.updatedIDs = append(c.state.updatedIDs, args[3].Value.(int64))
		return driver.RowsAffected(1), nil
	case strings.HasPrefix(q, "delete from alarm_firing"):
		return driver.RowsAffected(1), nil
	case strings.HasPrefix(q, "UPDATE k8s_alert_outbox SET payload="):
		if c.state.bindErr != nil {
			return nil, c.state.bindErr
		}
		var p map[string]any
		_ = json.Unmarshal([]byte(c.state.payload), &p)
		p["alarm_id"] = args[0].Value
		raw, _ := json.Marshal(p)
		c.state.payload = string(raw)
		return driver.RowsAffected(1), nil
	default:
		c.state.newAlarm = true
		return nil, errors.New("unexpected exec: " + q)
	}
}

type notificationResult int64

func (r notificationResult) LastInsertId() (int64, error) { return int64(r), nil }
func (notificationResult) RowsAffected() (int64, error)   { return 1, nil }
func useNotificationDB(t *testing.T, state *notificationDB) {
	t.Helper()
	db := sql.OpenDB(notificationConnector{state: state})
	engine, err := xorm.NewEngineWithDB("mysql", "", core.FromDB(db))
	if err != nil {
		t.Fatal(err)
	}
	old := x
	x = engine
	t.Cleanup(func() { x = old; _ = engine.Close() })
}

func useCoreNotificationServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/auth/") {
			fmt.Fprintf(w, `{"status":"OK","data":[{"tokenType":"accessToken","token":"access-placeholder","expiration":"%d"}]}`, time.Now().Add(time.Hour).Unix())
			return
		}
		handler(w, r)
	}))
	oldURL, oldKey, oldCode := m.CoreUrl, m.SubSystemKey, m.SubSystemCode
	t.Cleanup(func() { m.CoreUrl, m.SubSystemKey, m.SubSystemCode = oldURL, oldKey, oldCode; server.Close() })
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	m.CoreUrl, m.SubSystemKey, m.SubSystemCode = server.URL, base64.StdEncoding.EncodeToString(raw), "subsystem-a"
	m.InitCoreToken()
	return server
}

func TestRecoveryOutboxReusesAlarmAfterFailedDelivery(t *testing.T) {
	state := &notificationDB{payload: `{"key":{},"value":1}`, status: "firing"}
	useNotificationDB(t, state)
	for attempt := 0; attempt < 2; attempt++ {
		alarm := &m.AlarmHandleObj{AlarmTable: m.AlarmTable{Endpoint: "k8s-object", StrategyId: 1, SMetric: "rule-a", Status: "ok", End: time.Now()}}
		if err := persistOutboxAlarm(context.Background(), alarm, "event-a"); err != nil {
			t.Fatalf("attempt %d: %v", attempt, err)
		}
		if alarm.Id != 7 {
			t.Fatalf("attempt %d uses alarm %d", attempt, alarm.Id)
		}
		// The external delivery failed after the first commit. A later firing
		// must not be closed by the retry of this recovery event.
		if attempt == 0 {
			state.status = "firing"
		}
	}
	if state.commits != 2 || state.updates != 1 || state.newAlarm {
		t.Fatalf("unstable retry: %+v", state)
	}
}

func TestFirstFiringAndOutboxBindingShareTransaction(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprintf("binding_failure=%t", fail), func(t *testing.T) {
			state := &notificationDB{payload: `{}`}
			if fail {
				state.bindErr = errors.New("binding unavailable")
			}
			useNotificationDB(t, state)
			alarm := &m.AlarmHandleObj{AlarmTable: m.AlarmTable{Endpoint: "k8s-object", StrategyId: 1, SMetric: "rule-a", Status: "firing", Start: time.Now()}}
			err := persistOutboxAlarm(context.Background(), alarm, "firing-a")
			if state.begins != 1 || !state.newAlarm {
				t.Fatalf("firing did not share event transaction: %+v", state)
			}
			if fail {
				if !errors.Is(err, state.bindErr) || state.status != "" || state.payload != "{}" || state.commits != 0 {
					t.Fatalf("partial firing survived failed binding: %+v err=%v", state, err)
				}
			} else if err != nil || state.commits != 1 || alarm.Id != 7 || !strings.Contains(state.payload, `"alarm_id":7`) {
				t.Fatalf("first firing was not bound: %+v err=%v", state, err)
			}
		})
	}
}

func TestOutboxAlarmBindingFailureRollsBackRecovery(t *testing.T) {
	state := &notificationDB{payload: `{}`, status: "firing", bindErr: errors.New("binding unavailable")}
	useNotificationDB(t, state)
	alarm := &m.AlarmHandleObj{AlarmTable: m.AlarmTable{Status: "ok", End: time.Now()}}
	if err := persistOutboxAlarm(context.Background(), alarm, "event-a"); !errors.Is(err, state.bindErr) {
		t.Fatalf("original failure lost: %v", err)
	}
	if state.status != "firing" || state.payload != "{}" || state.commits != 0 {
		t.Fatalf("partial recovery committed: %+v", state)
	}
}

func TestRecoveryBindsItsFiringPeriodDespiteLaterFiring(t *testing.T) {
	state := &notificationDB{payload: `{"state":{"StartedAt":"2026-01-01T00:00:00.123456789Z","ObservedAt":"2026-01-01T00:01:00Z"}}`, status: "firing"}
	useNotificationDB(t, state)
	first := &m.AlarmHandleObj{AlarmTable: m.AlarmTable{Endpoint: "k8s-object", StrategyId: 1, SMetric: "rule-a", Status: "firing"}}
	if err := persistOutboxAlarm(context.Background(), first, "firing-a"); err != nil {
		t.Fatal(err)
	}
	if len(first.AlarmStrategy) != 64 {
		t.Fatal("episode identity exceeds alarm_strategy schema")
	}
	// A later period exists before this recovery is delivered. Only the
	// matching earlier period may be selected, regardless of row ordering.
	state.alarms = map[string]int64{first.AlarmStrategy: 7, "later-period": 8}
	state.payload = `{"state":{"StartedAt":"2026-01-01T00:00:00.123457Z","ObservedAt":"2026-01-01T00:02:00Z"}}`
	recovered := &m.AlarmHandleObj{AlarmTable: m.AlarmTable{Endpoint: first.Endpoint, StrategyId: 1, SMetric: first.SMetric, Status: "ok"}}
	if err := persistOutboxAlarm(context.Background(), recovered, "recovery-a"); err != nil {
		t.Fatal(err)
	}
	if recovered.Id != 7 || recovered.AlarmStrategy != first.AlarmStrategy || len(state.updatedIDs) != 1 || state.updatedIDs[0] != 7 {
		t.Fatalf("wrong period recovered: %+v / %+v", recovered, state)
	}
}

func TestStaleFiringDeliveryIsSkippedAfterRecovery(t *testing.T) {
	state := &notificationDB{status: "ok"}
	useNotificationDB(t, state)
	alarm := &m.AlarmHandleObj{AlarmTable: m.AlarmTable{Id: 7, Status: "firing"}}
	// No transport is configured. An obsolete firing must complete before
	// consulting notification configuration or issuing an external request.
	if err := sendAlarmNotification(context.Background(), alarm, "firing-a"); err != nil {
		t.Fatal(err)
	}
}

func TestCoreRetryIdentitySurvivesCallbackReordering(t *testing.T) {
	seen := map[string][]string{}
	useCoreNotificationServer(t, func(w http.ResponseWriter, r *http.Request) {
		var p m.CoreNotifyRequest
		_ = json.NewDecoder(r.Body).Decode(&p)
		seen[p.OperationKey] = append(seen[p.OperationKey], p.EventSeqNo)
		w.Write([]byte(`{"status":"OK"}`))
	})
	for _, keys := range [][]string{{"entity-a^callback-a", "entity-b^callback-b"}, {"entity-b^callback-b", "entity-a^callback-a"}} {
		if err := sendCoreEvents(context.Background(), 7, "ok", keys, "event-a"); err != nil {
			t.Fatal(err)
		}
	}
	if len(seen) != 2 {
		t.Fatalf("unexpected callback count: %v", seen)
	}
	for _, ids := range seen {
		if len(ids) != 2 || ids[0] != ids[1] {
			t.Fatalf("callback identity changed: %v", seen)
		}
	}
	if seen["callback-a"][0] == seen["callback-b"][0] {
		t.Fatal("different callbacks share retry identity")
	}
}

func TestCoreNotificationRejectsHTTPAndBusinessFailures(t *testing.T) {
	for _, tc := range []struct {
		name   string
		code   int
		body   string
		failed bool
	}{
		{"http", 500, `{"status":"OK"}`, true},
		{"business", 200, `{"status":"ERROR","message":"rejected"}`, true},
		{"invalid", 200, `not json`, true},
		{"success", 200, `{"status":"OK"}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var sequences []string
			useCoreNotificationServer(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/platform/v1/operation-events" {
					w.Write([]byte(`{"data":[]}`))
					return
				}
				var p m.CoreNotifyRequest
				_ = json.NewDecoder(r.Body).Decode(&p)
				sequences = append(sequences, p.EventSeqNo)
				w.WriteHeader(tc.code)
				w.Write([]byte(tc.body))
			})
			for i := 0; i < 2; i++ {
				err := sendCoreEvents(context.Background(), 7, "ok", []string{"entity-a^callback-a"}, "event-a")
				if (err != nil) != tc.failed {
					t.Fatalf("unexpected delivery result: %v", err)
				}
			}
			if len(sequences) != 2 || sequences[0] != sequences[1] {
				t.Fatalf("retry identity changed: %v", sequences)
			}
		})
	}
}

func TestCoreNotificationHonorsCancellation(t *testing.T) {
	stop := make(chan struct{})
	useCoreNotificationServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/platform/v1/operation-events" {
			w.Write([]byte(`{"data":[]}`))
			return
		}
		select {
		case <-r.Context().Done():
		case <-stop:
		}
	})
	defer close(stop)
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	started := time.Now()
	err := sendCoreEvents(ctx, 7, "firing", []string{"entity-a^callback-a"}, "event-a")
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) > time.Second {
		t.Fatalf("unbounded Core notification: %v", err)
	}
}
