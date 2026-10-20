package elydora_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	elydora "github.com/Elydora-Infra/Elydora-Go-SDK/v2"
)

func randomHex(n int) string {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func randomSeedBase64url(t *testing.T) (seedB64 string, pubB64 string) {
	t.Helper()

	seed := make([]byte, ed25519.SeedSize)
	if _, err := rand.Read(seed); err != nil {
		t.Fatalf("read random seed: %v", err)
	}

	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return base64.RawURLEncoding.EncodeToString(seed), base64.RawURLEncoding.EncodeToString(publicKey)
}

func TestSDKCompatibilityAgainstLatestAPI(t *testing.T) {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		t.Skip("API_BASE_URL is not set; skipping SDK compatibility integration test")
	}
	suffix := randomHex(6)
	email := fmt.Sprintf("go-sdk-%s@elydora.test", suffix)
	password := fmt.Sprintf("Go-%s-Password!", suffix)
	agentID := fmt.Sprintf("go-agent-%s", randomHex(8))
	kid := fmt.Sprintf("go-kid-%s", randomHex(8))
	privateKey, publicKey := randomSeedBase64url(t)

	if _, err := elydora.Register(baseURL, email, password, elydora.WithDisplayName("go-sdk-user"), elydora.WithOrgName("go-org-"+suffix)); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	login, err := elydora.Login(baseURL, email, password)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	client, err := elydora.NewClient(&elydora.Config{
		OrgID:      login.User.OrgID,
		AgentID:    agentID,
		PrivateKey: privateKey,
		BaseURL:    baseURL,
		TTLMs:      120000,
		MaxRetries: 1,
		Token:      login.Token,
	})
	if err != nil {
		t.Fatalf("new client failed: %v", err)
	}

	ttl := 1800
	tokenResp, err := client.IssueApiToken(&elydora.IssueApiTokenRequest{TTLSeconds: &ttl})
	if err != nil {
		t.Fatalf("issue api token failed: %v", err)
	}
	if tokenResp.TokenID == "" {
		t.Fatalf("issue api token: token_id missing")
	}

	oldToken := tokenResp.Token
	client.SetToken(oldToken)
	meWithAPI, err := client.GetMe()
	if err != nil {
		t.Fatalf("api token auth failed: %v", err)
	}
	if orgID(meWithAPI.User.OrgID) != login.User.OrgID {
		t.Fatalf("api token org mismatch: got %s want %s", orgID(meWithAPI.User.OrgID), login.User.OrgID)
	}

	rotateResp, err := client.RotateApiToken()
	if err != nil {
		t.Fatalf("rotate api token failed: %v", err)
	}
	if rotateResp.Token == "" {
		t.Fatalf("rotate response missing token: %+v", rotateResp)
	}

	client.SetToken(rotateResp.Token)
	meWithRotated, err := client.GetMe()
	if err != nil {
		t.Fatalf("rotated token auth failed: %v", err)
	}
	if orgID(meWithRotated.User.OrgID) != login.User.OrgID {
		t.Fatalf("rotated token org mismatch: got %s want %s", orgID(meWithRotated.User.OrgID), login.User.OrgID)
	}

	_, err = client.RegisterAgent(&elydora.RegisterAgentRequest{
		AgentID:           agentID,
		DisplayName:       "Go SDK Compat Agent",
		ResponsibleEntity: "sdk-matrix",
		IntegrationType:   elydora.IntegrationTypeSDK,
		Keys: []elydora.RegisterAgentKeyInput{
			{
				KID:       kid,
				PublicKey: publicKey,
				Algorithm: "ed25519",
			},
		},
	})
	if err != nil {
		t.Fatalf("register agent failed: %v", err)
	}

	op, err := client.CreateOperation(&elydora.CreateOperationParams{
		OperationType: "compat.go.submit",
		Subject:       map[string]interface{}{"entity": "sdk", "id": "go-" + suffix},
		Action:        map[string]interface{}{"type": "write", "resource": "matrix"},
		Payload:       map[string]interface{}{"sdk": "go"},
		KID:           kid,
	})
	if err != nil {
		t.Fatalf("create operation failed: %v", err)
	}

	submit, err := client.SubmitOperation(op)
	if err != nil {
		t.Fatalf("submit operation failed: %v", err)
	}
	if submit.Receipt.OperationID != op.OperationID {
		t.Fatalf("receipt operation mismatch: got %s want %s", submit.Receipt.OperationID, op.OperationID)
	}

	got, err := client.GetOperation(op.OperationID)
	if err != nil {
		t.Fatalf("get operation failed: %v", err)
	}
	if got.Operation.OperationID != op.OperationID {
		t.Fatalf("get operation mismatch: got %s want %s", got.Operation.OperationID, op.OperationID)
	}

	verify, err := client.VerifyOperation(op.OperationID)
	if err != nil {
		t.Fatalf("verify operation failed: %v", err)
	}
	if !verify.Valid {
		t.Fatalf("verify operation invalid: %+v", verify)
	}
}

func orgID(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
