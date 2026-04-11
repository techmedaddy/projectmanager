package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskflow/backend/internal/auth"
	"taskflow/backend/internal/config"
	"taskflow/backend/internal/db"
	"taskflow/backend/internal/tasks"
)

const (
	testJWTSecret      = "integration-test-secret-value"
	seedUserID         = "11111111-1111-1111-1111-111111111111"
	seedProjectID      = "22222222-2222-2222-2222-222222222222"
	seedTaskID         = "33333333-3333-3333-3333-333333333333"
	seedUserEmail      = "test@example.com"
	seedUserPassword   = "password123"
	secondUserPassword = "password123"
)

var (
	integrationSetupOnce sync.Once
	integrationHandler   http.Handler
	integrationDB        *db.Database
	integrationExecPool  *pgxpool.Pool
	integrationSetupErr  error
	integrationSkip      string
)

func TestMain(m *testing.M) {
	code := m.Run()

	if integrationExecPool != nil {
		integrationExecPool.Close()
	}

	if integrationDB != nil {
		integrationDB.Close()
	}

	os.Exit(code)
}

func TestRegisterLoginFlow(t *testing.T) {
	handler := setupIntegrationTest(t)

	registerResp := performJSONRequest(t, handler, http.MethodPost, "/auth/register", auth.RegisterRequest{
		Name:     "Integration User",
		Email:    "  NewUser@Example.com  ",
		Password: "password123",
	}, nil)
	if registerResp.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %s", registerResp.Code, registerResp.Body.String())
	}

	var created auth.RegisterResponse
	decodeJSONBody(t, registerResp, &created)

	if created.User.Email != "newuser@example.com" {
		t.Fatalf("normalized email = %q, want %q", created.User.Email, "newuser@example.com")
	}

	loginResp := performJSONRequest(t, handler, http.MethodPost, "/auth/login", auth.LoginRequest{
		Email:    "newuser@example.com",
		Password: "password123",
	}, nil)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", loginResp.Code, loginResp.Body.String())
	}

	var login auth.LoginResponse
	decodeJSONBody(t, loginResp, &login)

	if login.AccessToken == "" {
		t.Fatal("login access token is empty")
	}

	meResp := performJSONRequest(t, handler, http.MethodGet, "/auth/me", nil, map[string]string{
		"Authorization": "Bearer " + login.AccessToken,
	})
	if meResp.Code != http.StatusOK {
		t.Fatalf("me status = %d, body = %s", meResp.Code, meResp.Body.String())
	}

	var current auth.CurrentUserResponse
	decodeJSONBody(t, meResp, &current)

	if current.User.Email != "newuser@example.com" {
		t.Fatalf("current user email = %q, want %q", current.User.Email, "newuser@example.com")
	}
}

func TestProtectedRouteRejectsMissingToken(t *testing.T) {
	handler := setupIntegrationTest(t)

	resp := performJSONRequest(t, handler, http.MethodGet, "/projects", nil, nil)
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body = %s", resp.Code, http.StatusUnauthorized, resp.Body.String())
	}

	assertErrorMessage(t, resp, "unauthenticated")
}

func TestCreateTaskInProject(t *testing.T) {
	handler := setupIntegrationTest(t)
	token := loginSeedUser(t, handler)

	resp := performJSONRequest(t, handler, http.MethodPost, "/projects/"+seedProjectID+"/tasks", tasks.CreateRequest{
		Title:    "Integration Task",
		Status:   ptrTaskStatus(tasks.StatusInProgress),
		Priority: ptrTaskPriority(tasks.PriorityHigh),
		DueDate:  ptrString("2026-04-25"),
	}, map[string]string{
		"Authorization": "Bearer " + token,
	})
	if resp.Code != http.StatusCreated {
		t.Fatalf("create task status = %d, body = %s", resp.Code, resp.Body.String())
	}

	var created tasks.Response
	decodeJSONBody(t, resp, &created)

	if created.CreatorID != seedUserID {
		t.Fatalf("creator_id = %q, want %q", created.CreatorID, seedUserID)
	}

	if created.ProjectID != seedProjectID {
		t.Fatalf("project_id = %q, want %q", created.ProjectID, seedProjectID)
	}

	if created.Status != tasks.StatusInProgress {
		t.Fatalf("status = %q, want %q", created.Status, tasks.StatusInProgress)
	}

	if created.Priority != tasks.PriorityHigh {
		t.Fatalf("priority = %q, want %q", created.Priority, tasks.PriorityHigh)
	}

	if created.DueDate == nil || *created.DueDate != "2026-04-25" {
		t.Fatalf("due_date = %v, want %q", created.DueDate, "2026-04-25")
	}
}

func TestDeleteTaskAuthorization(t *testing.T) {
	handler := setupIntegrationTest(t)
	token := registerAndLoginUser(t, handler, "Other User", "other@example.com", secondUserPassword)

	resp := performJSONRequest(t, handler, http.MethodDelete, "/tasks/"+seedTaskID, nil, map[string]string{
		"Authorization": "Bearer " + token,
	})
	if resp.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", resp.Code, http.StatusForbidden, resp.Body.String())
	}

	assertErrorMessage(t, resp, "forbidden")
}

func TestProjectsVisibilityWhenUserIsAssignee(t *testing.T) {
	handler := setupIntegrationTest(t)
	ownerToken := loginSeedUser(t, handler)
	assigneeToken := registerAndLoginUser(
		t,
		handler,
		"Assignee User",
		"assignee-visibility@example.com",
		secondUserPassword,
	)

	assigneeMe := performJSONRequest(t, handler, http.MethodGet, "/auth/me", nil, map[string]string{
		"Authorization": "Bearer " + assigneeToken,
	})
	if assigneeMe.Code != http.StatusOK {
		t.Fatalf("assignee me status = %d, body = %s", assigneeMe.Code, assigneeMe.Body.String())
	}

	var assigneeCurrent auth.CurrentUserResponse
	decodeJSONBody(t, assigneeMe, &assigneeCurrent)
	assigneeID := assigneeCurrent.User.ID

	createdTaskResp := performJSONRequest(t, handler, http.MethodPost, "/projects/"+seedProjectID+"/tasks", tasks.CreateRequest{
		Title:      "Visibility assignee task",
		AssigneeID: ptrString(assigneeID),
	}, map[string]string{
		"Authorization": "Bearer " + ownerToken,
	})
	if createdTaskResp.Code != http.StatusCreated {
		t.Fatalf("create task status = %d, body = %s", createdTaskResp.Code, createdTaskResp.Body.String())
	}

	listResp := performJSONRequest(t, handler, http.MethodGet, "/projects", nil, map[string]string{
		"Authorization": "Bearer " + assigneeToken,
	})
	if listResp.Code != http.StatusOK {
		t.Fatalf("list projects status = %d, body = %s", listResp.Code, listResp.Body.String())
	}

	var listBody struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	decodeJSONBody(t, listResp, &listBody)

	if !containsProjectID(listBody.Items, seedProjectID) {
		t.Fatalf("expected project %s to be visible to assignee", seedProjectID)
	}

	detailResp := performJSONRequest(t, handler, http.MethodGet, "/projects/"+seedProjectID, nil, map[string]string{
		"Authorization": "Bearer " + assigneeToken,
	})
	if detailResp.Code != http.StatusOK {
		t.Fatalf("project detail status = %d, body = %s", detailResp.Code, detailResp.Body.String())
	}
}

func TestProjectsVisibilityWhenUserIsCreatorButNotAssignee(t *testing.T) {
	handler := setupIntegrationTest(t)
	ownerToken := loginSeedUser(t, handler)
	creatorToken := registerAndLoginUser(
		t,
		handler,
		"Creator User",
		"creator-visibility@example.com",
		secondUserPassword,
	)

	creatorMe := performJSONRequest(t, handler, http.MethodGet, "/auth/me", nil, map[string]string{
		"Authorization": "Bearer " + creatorToken,
	})
	if creatorMe.Code != http.StatusOK {
		t.Fatalf("creator me status = %d, body = %s", creatorMe.Code, creatorMe.Body.String())
	}

	var creatorCurrent auth.CurrentUserResponse
	decodeJSONBody(t, creatorMe, &creatorCurrent)
	creatorID := creatorCurrent.User.ID

	bootstrapResp := performJSONRequest(t, handler, http.MethodPost, "/projects/"+seedProjectID+"/tasks", tasks.CreateRequest{
		Title:      "Bootstrap creator access",
		AssigneeID: ptrString(creatorID),
	}, map[string]string{
		"Authorization": "Bearer " + ownerToken,
	})
	if bootstrapResp.Code != http.StatusCreated {
		t.Fatalf("bootstrap task status = %d, body = %s", bootstrapResp.Code, bootstrapResp.Body.String())
	}

	var bootstrapTask tasks.Response
	decodeJSONBody(t, bootstrapResp, &bootstrapTask)

	creatorTaskResp := performJSONRequest(t, handler, http.MethodPost, "/projects/"+seedProjectID+"/tasks", tasks.CreateRequest{
		Title:      "Creator-only task",
		AssigneeID: ptrString(seedUserID),
	}, map[string]string{
		"Authorization": "Bearer " + creatorToken,
	})
	if creatorTaskResp.Code != http.StatusCreated {
		t.Fatalf("creator task status = %d, body = %s", creatorTaskResp.Code, creatorTaskResp.Body.String())
	}

	deleteBootstrapResp := performJSONRequest(t, handler, http.MethodDelete, "/tasks/"+bootstrapTask.ID, nil, map[string]string{
		"Authorization": "Bearer " + ownerToken,
	})
	if deleteBootstrapResp.Code != http.StatusNoContent {
		t.Fatalf("delete bootstrap status = %d, body = %s", deleteBootstrapResp.Code, deleteBootstrapResp.Body.String())
	}

	listResp := performJSONRequest(t, handler, http.MethodGet, "/projects", nil, map[string]string{
		"Authorization": "Bearer " + creatorToken,
	})
	if listResp.Code != http.StatusOK {
		t.Fatalf("list projects status = %d, body = %s", listResp.Code, listResp.Body.String())
	}

	var listBody struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	decodeJSONBody(t, listResp, &listBody)

	if !containsProjectID(listBody.Items, seedProjectID) {
		t.Fatalf("expected project %s to remain visible to creator", seedProjectID)
	}

	detailResp := performJSONRequest(t, handler, http.MethodGet, "/projects/"+seedProjectID, nil, map[string]string{
		"Authorization": "Bearer " + creatorToken,
	})
	if detailResp.Code != http.StatusOK {
		t.Fatalf("project detail status = %d, body = %s", detailResp.Code, detailResp.Body.String())
	}
}

func TestProjectsVisibilityDeniedWhenUserIsNotOwnerCreatorOrAssignee(t *testing.T) {
	handler := setupIntegrationTest(t)
	outsiderToken := registerAndLoginUser(
		t,
		handler,
		"Outsider User",
		"outsider-visibility@example.com",
		secondUserPassword,
	)

	listResp := performJSONRequest(t, handler, http.MethodGet, "/projects", nil, map[string]string{
		"Authorization": "Bearer " + outsiderToken,
	})
	if listResp.Code != http.StatusOK {
		t.Fatalf("list projects status = %d, body = %s", listResp.Code, listResp.Body.String())
	}

	var listBody struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	decodeJSONBody(t, listResp, &listBody)

	if containsProjectID(listBody.Items, seedProjectID) {
		t.Fatalf("did not expect project %s to be visible to outsider", seedProjectID)
	}

	detailResp := performJSONRequest(t, handler, http.MethodGet, "/projects/"+seedProjectID, nil, map[string]string{
		"Authorization": "Bearer " + outsiderToken,
	})
	if detailResp.Code != http.StatusForbidden {
		t.Fatalf("project detail status = %d, want %d; body = %s", detailResp.Code, http.StatusForbidden, detailResp.Body.String())
	}

	assertErrorMessage(t, detailResp, "forbidden")
}

func setupIntegrationTest(t *testing.T) http.Handler {
	t.Helper()

	testConfig := integrationConfig(t)
	if integrationSkip != "" {
		t.Skip(integrationSkip)
	}

	integrationSetupOnce.Do(func() {
		integrationExecPool, integrationSetupErr = newIntegrationExecPool(context.Background(), testConfig.DatabaseURL)
		if integrationSetupErr != nil {
			return
		}

		if integrationSetupErr = resetIntegrationSchema(context.Background(), integrationExecPool); integrationSetupErr != nil {
			return
		}

		integrationDB, integrationSetupErr = db.New(context.Background(), testConfig.DatabaseURL)
		if integrationSetupErr != nil {
			return
		}

		logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
		app := newApplication(logger, testConfig, integrationDB)
		integrationHandler = newRouter(app)
	})

	if integrationSkip != "" {
		t.Skip(integrationSkip)
	}

	if integrationSetupErr != nil {
		t.Fatalf("integration setup failed: %v", integrationSetupErr)
	}

	if err := reseedIntegrationDatabase(context.Background(), integrationExecPool); err != nil {
		t.Fatalf("reseed integration database: %v", err)
	}

	return integrationHandler
}

func integrationConfig(t *testing.T) config.Config {
	t.Helper()

	testDatabaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if testDatabaseURL == "" {
		integrationSkip = "set TEST_DATABASE_URL to run integration tests"
		return config.Config{}
	}

	productionDatabaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if productionDatabaseURL != "" && productionDatabaseURL == testDatabaseURL {
		t.Fatal("TEST_DATABASE_URL must point to a dedicated test database and must not equal DATABASE_URL")
	}

	return config.Config{
		AppPort:        0,
		DatabaseURL:    testDatabaseURL,
		JWTSecret:      testJWTSecret,
		JWTExpiryHours: 24,
		BcryptCost:     12,
	}
}

func newIntegrationExecPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse test pgx config: %w", err)
	}

	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create test pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping test database: %w", err)
	}

	return pool, nil
}

func resetIntegrationSchema(ctx context.Context, pool *pgxpool.Pool) error {
	downFiles, err := sqlFiles("migrations", "*_down.sql")
	if err != nil {
		return err
	}

	sort.Sort(sort.Reverse(sort.StringSlice(downFiles)))
	for _, path := range downFiles {
		if err := execSQLFile(ctx, pool, path); err != nil {
			return fmt.Errorf("run down migration %s: %w", filepath.Base(path), err)
		}
	}

	upFiles, err := sqlFiles("migrations", "*_up.sql")
	if err != nil {
		return err
	}

	sort.Strings(upFiles)
	for _, path := range upFiles {
		if err := execSQLFile(ctx, pool, path); err != nil {
			return fmt.Errorf("run up migration %s: %w", filepath.Base(path), err)
		}
	}

	return nil
}

func reseedIntegrationDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `TRUNCATE TABLE tasks, projects, users RESTART IDENTITY CASCADE`); err != nil {
		return fmt.Errorf("truncate test tables: %w", err)
	}

	seedPath := filepath.Join(projectRoot(), "seed", "001_seed.sql")
	if err := execSQLFile(ctx, pool, seedPath); err != nil {
		return fmt.Errorf("execute seed sql: %w", err)
	}

	return nil
}

func sqlFiles(dirName, pattern string) ([]string, error) {
	paths, err := filepath.Glob(filepath.Join(projectRoot(), dirName, pattern))
	if err != nil {
		return nil, fmt.Errorf("glob sql files: %w", err)
	}

	return paths, nil
}

func execSQLFile(ctx context.Context, pool *pgxpool.Pool, path string) error {
	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read sql file: %w", err)
	}

	if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("exec sql: %w", err)
	}

	return nil
}

func projectRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime caller unavailable")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func performJSONRequest(t *testing.T, handler http.Handler, method, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	var requestBody io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		requestBody = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, requestBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func decodeJSONBody(t *testing.T, recorder *httptest.ResponseRecorder, dst any) {
	t.Helper()

	if err := json.Unmarshal(recorder.Body.Bytes(), dst); err != nil {
		t.Fatalf("decode response body: %v; body = %s", err, recorder.Body.String())
	}
}

func assertErrorMessage(t *testing.T, recorder *httptest.ResponseRecorder, want string) {
	t.Helper()

	var body struct {
		Error string `json:"error"`
	}
	decodeJSONBody(t, recorder, &body)

	if body.Error != want {
		t.Fatalf("error = %q, want %q", body.Error, want)
	}
}

func loginSeedUser(t *testing.T, handler http.Handler) string {
	t.Helper()

	resp := performJSONRequest(t, handler, http.MethodPost, "/auth/login", auth.LoginRequest{
		Email:    seedUserEmail,
		Password: seedUserPassword,
	}, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("seed login status = %d, body = %s", resp.Code, resp.Body.String())
	}

	var login auth.LoginResponse
	decodeJSONBody(t, resp, &login)
	return login.AccessToken
}

func registerAndLoginUser(t *testing.T, handler http.Handler, name, email, password string) string {
	t.Helper()

	registerResp := performJSONRequest(t, handler, http.MethodPost, "/auth/register", auth.RegisterRequest{
		Name:     name,
		Email:    email,
		Password: password,
	}, nil)
	if registerResp.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %s", registerResp.Code, registerResp.Body.String())
	}

	loginResp := performJSONRequest(t, handler, http.MethodPost, "/auth/login", auth.LoginRequest{
		Email:    email,
		Password: password,
	}, nil)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", loginResp.Code, loginResp.Body.String())
	}

	var login auth.LoginResponse
	decodeJSONBody(t, loginResp, &login)
	return login.AccessToken
}

func ptrString(value string) *string {
	return &value
}

func ptrTaskStatus(value tasks.Status) *tasks.Status {
	return &value
}

func ptrTaskPriority(value tasks.Priority) *tasks.Priority {
	return &value
}

func containsProjectID(projects []struct {
	ID string `json:"id"`
}, target string) bool {
	for _, project := range projects {
		if project.ID == target {
			return true
		}
	}

	return false
}
