package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// uploadArtifactRequest performs a multipart artifact upload with the given
// token against /api/v1/developer/apps/{appID}/artifacts.
func uploadArtifactRequest(
	t *testing.T,
	app http.Handler,
	token string,
	appID int64,
	fileName string,
	content string,
) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("artifact", fileName)
	if err != nil {
		t.Fatalf("create artifact form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write artifact form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/developer/apps/"+strconv.FormatInt(appID, 10)+"/artifacts",
		&body,
	)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	return response
}

// publishAndApprove creates an app as the developer and approves it as admin.
func publishAndApprove(t *testing.T, app http.Handler, developerToken, adminToken, name, slug string) int64 {
	t.Helper()
	body := fmt.Sprintf(`{"name":"%s","slug":"%s","tagline":"T.","description":"D.","category":"developer-tools","priceCents":2900,"currency":"USD","iconUrl":"","coverImageUrl":"","demoUrl":"","docsUrl":"","sourceUrl":"","supportUrl":"","tags":[],"version":"2.0.0","releaseNotes":"R."}`, name, slug)
	submitted := performJSONRequestWithToken(t, app, http.MethodPost, "/api/v1/apps", body, developerToken)
	if submitted.Code != http.StatusCreated {
		t.Fatalf("publish status = %d, want %d; body = %s", submitted.Code, http.StatusCreated, submitted.Body.String())
	}
	var submittedBody struct {
		App struct {
			ID int64 `json:"id"`
		} `json:"app"`
	}
	if err := json.NewDecoder(submitted.Body).Decode(&submittedBody); err != nil {
		t.Fatalf("decode published app: %v", err)
	}
	approved := performJSONRequestWithToken(
		t, app, http.MethodPost,
		"/api/v1/admin/apps/"+strconv.FormatInt(submittedBody.App.ID, 10)+"/approve",
		`{"reviewNote":"Approved."}`, adminToken,
	)
	if approved.Code != http.StatusOK {
		t.Fatalf("approve app status = %d, want %d; body = %s", approved.Code, http.StatusOK, approved.Body.String())
	}
	return submittedBody.App.ID
}

func TestArtifactUploadListDeleteAndSecureDownloadFlow(t *testing.T) {
	db, app := newAuthTestApp(t)
	devID := insertVerifiedUser(t, db, "dev@example.com", "old-password")
	insertVerifiedUser(t, db, "buyer@example.com", "old-password")
	insertVerifiedUser(t, db, "other@example.com", "old-password")
	adminID := insertVerifiedUser(t, db, "admin@example.com", "old-password")
	if _, err := db.Exec("UPDATE users SET role = 'admin' WHERE id = ?", adminID); err != nil {
		t.Fatalf("mark admin: %v", err)
	}
	if _, err := db.Exec("INSERT INTO developers(user_id, display_name) VALUES (?, ?)", devID, "Dev Studio"); err != nil {
		t.Fatalf("insert developer: %v", err)
	}

	developerToken := loginToken(t, app, "dev@example.com", "old-password")
	buyerToken := loginToken(t, app, "buyer@example.com", "old-password")
	otherToken := loginToken(t, app, "other@example.com", "old-password")
	adminToken := loginToken(t, app, "admin@example.com", "old-password")

	appID := publishAndApprove(t, app, developerToken, adminToken, "Artifact App", "artifact-app")

	// --- Buyers cannot upload artifacts. ---
	forbidden := uploadArtifactRequest(t, app, buyerToken, appID, "x.zip", "buyer upload")
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("buyer upload status = %d, want %d", forbidden.Code, http.StatusForbidden)
	}

	// --- Developer uploads an artifact. ---
	uploaded := uploadArtifactRequest(t, app, developerToken, appID, "release-v2.0.0.zip", "ZIP-PAYLOAD")
	if uploaded.Code != http.StatusCreated {
		t.Fatalf("upload status = %d, want %d; body = %s", uploaded.Code, http.StatusCreated, uploaded.Body.String())
	}
	var uploadedBody struct {
		Artifact struct {
			ID             int64  `json:"id"`
			AppID          int64  `json:"appId"`
			FileName       string `json:"fileName"`
			SizeBytes      int64  `json:"sizeBytes"`
			ChecksumSHA256 string `json:"checksumSha256"`
		} `json:"artifact"`
	}
	if err := json.NewDecoder(uploaded.Body).Decode(&uploadedBody); err != nil {
		t.Fatalf("decode uploaded artifact: %v", err)
	}
	artifactID := uploadedBody.Artifact.ID
	if uploadedBody.Artifact.AppID != appID || uploadedBody.Artifact.FileName != "release-v2.0.0.zip" ||
		uploadedBody.Artifact.SizeBytes != int64(len("ZIP-PAYLOAD")) {
		t.Fatalf("uploaded artifact = %+v, want matching metadata", uploadedBody.Artifact)
	}

	// --- Developer lists artifacts. ---
	listResp := performJSONRequestWithToken(
		t, app, http.MethodGet,
		"/api/v1/developer/apps/"+strconv.FormatInt(appID, 10)+"/artifacts",
		"", developerToken,
	)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list artifacts status = %d, want %d", listResp.Code, http.StatusOK)
	}
	var listBody struct {
		Artifacts []struct {
			ID       int64  `json:"id"`
			FileName string `json:"fileName"`
		} `json:"artifacts"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&listBody); err != nil {
		t.Fatalf("decode list artifacts: %v", err)
	}
	if len(listBody.Artifacts) != 1 || listBody.Artifacts[0].ID != artifactID {
		t.Fatalf("list artifacts = %+v, want one artifact", listBody.Artifacts)
	}

	// --- Buy: checkout, confirm payment, get delivery. ---
	checkout := performJSONRequestWithToken(
		t, app, http.MethodPost, "/api/v1/marketplace/apps/artifact-app/checkout",
		`{}`, buyerToken,
	)
	if checkout.Code != http.StatusCreated {
		t.Fatalf("checkout status = %d, want %d; body = %s", checkout.Code, http.StatusCreated, checkout.Body.String())
	}
	var checkoutBody struct {
		Order struct {
			ID int64 `json:"id"`
		} `json:"order"`
	}
	if err := json.NewDecoder(checkout.Body).Decode(&checkoutBody); err != nil {
		t.Fatalf("decode checkout: %v", err)
	}
	confirm := performJSONRequestWithToken(
		t, app, http.MethodPost,
		"/api/v1/orders/"+strconv.FormatInt(checkoutBody.Order.ID, 10)+"/confirm-payment",
		`{}`, buyerToken,
	)
	if confirm.Code != http.StatusOK {
		t.Fatalf("confirm payment status = %d, want %d; body = %s", confirm.Code, http.StatusOK, confirm.Body.String())
	}

	myEnts := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/me/entitlements", "", buyerToken)
	var myEntsBody struct {
		Entitlements []struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"entitlements"`
	}
	if err := json.NewDecoder(myEnts.Body).Decode(&myEntsBody); err != nil {
		t.Fatalf("decode my entitlements: %v", err)
	}
	if len(myEntsBody.Entitlements) != 1 {
		t.Fatalf("my entitlements = %+v, want one", myEntsBody.Entitlements)
	}
	entitlementID := myEntsBody.Entitlements[0].ID

	delivery := performJSONRequestWithToken(
		t, app, http.MethodGet,
		"/api/v1/entitlements/"+strconv.FormatInt(entitlementID, 10)+"/delivery",
		"", buyerToken,
	)
	if delivery.Code != http.StatusOK {
		t.Fatalf("delivery status = %d, want %d; body = %s", delivery.Code, http.StatusOK, delivery.Body.String())
	}
	var deliveryBody struct {
		Artifacts []struct {
			ID       int64  `json:"id"`
			FileName string `json:"fileName"`
		} `json:"artifacts"`
		DeliveryToken string `json:"deliveryToken"`
	}
	if err := json.NewDecoder(delivery.Body).Decode(&deliveryBody); err != nil {
		t.Fatalf("decode delivery: %v", err)
	}
	if len(deliveryBody.Artifacts) != 1 || deliveryBody.Artifacts[0].ID != artifactID {
		t.Fatalf("delivery artifacts = %+v, want one matching artifact", deliveryBody.Artifacts)
	}
	if deliveryBody.DeliveryToken == "" {
		t.Fatalf("delivery token missing")
	}
	token := deliveryBody.DeliveryToken

	// --- Secure download: no token is rejected. ---
	noToken := performJSONRequest(
		t, app, http.MethodGet,
		"/api/v1/artifacts/"+strconv.FormatInt(artifactID, 10)+"/download",
		"",
	)
	if noToken.Code != http.StatusUnauthorized {
		t.Fatalf("download without token status = %d, want %d", noToken.Code, http.StatusUnauthorized)
	}

	// --- Secure download: invalid token is rejected. ---
	badToken := performJSONRequest(
		t, app, http.MethodGet,
		"/api/v1/artifacts/"+strconv.FormatInt(artifactID, 10)+"/download?token=bogus",
		"",
	)
	if badToken.Code != http.StatusUnauthorized {
		t.Fatalf("download with bogus token status = %d, want %d", badToken.Code, http.StatusUnauthorized)
	}

	// --- Secure download: valid token serves the artifact content. ---
	download := performJSONRequest(
		t, app, http.MethodGet,
		"/api/v1/artifacts/"+strconv.FormatInt(artifactID, 10)+"/download?token="+token,
		"",
	)
	if download.Code != http.StatusOK {
		t.Fatalf("download status = %d, want %d; body = %s", download.Code, http.StatusOK, download.Body.String())
	}
	if download.Body.String() != "ZIP-PAYLOAD" {
		t.Fatalf("download body = %q, want %q", download.Body.String(), "ZIP-PAYLOAD")
	}

	// --- Developer deletes the artifact; download then fails. ---
	delResp := performJSONRequestWithToken(
		t, app, http.MethodDelete,
		"/api/v1/developer/apps/"+strconv.FormatInt(appID, 10)+"/artifacts/"+strconv.FormatInt(artifactID, 10),
		"", developerToken,
	)
	if delResp.Code != http.StatusOK {
		t.Fatalf("delete artifact status = %d, want %d; body = %s", delResp.Code, http.StatusOK, delResp.Body.String())
	}
	afterDelete := performJSONRequest(
		t, app, http.MethodGet,
		"/api/v1/artifacts/"+strconv.FormatInt(artifactID, 10)+"/download?token="+token,
		"",
	)
	if afterDelete.Code != http.StatusNotFound {
		t.Fatalf("download after delete status = %d, want %d", afterDelete.Code, http.StatusNotFound)
	}

	// --- Foreign developer cannot delete another's artifact. ---
	uploaded2 := uploadArtifactRequest(t, app, developerToken, appID, "second.zip", "SECOND")
	if uploaded2.Code != http.StatusCreated {
		t.Fatalf("second upload status = %d, want %d; body = %s", uploaded2.Code, http.StatusCreated, uploaded2.Body.String())
	}
	var uploaded2Body struct {
		Artifact struct {
			ID int64 `json:"id"`
		} `json:"artifact"`
	}
	if err := json.NewDecoder(uploaded2.Body).Decode(&uploaded2Body); err != nil {
		t.Fatalf("decode second upload: %v", err)
	}
	foreignDelete := performJSONRequestWithToken(
		t, app, http.MethodDelete,
		"/api/v1/developer/apps/"+strconv.FormatInt(appID, 10)+"/artifacts/"+strconv.FormatInt(uploaded2Body.Artifact.ID, 10),
		"", otherToken,
	)
	if foreignDelete.Code != http.StatusForbidden {
		t.Fatalf("foreign delete status = %d, want %d", foreignDelete.Code, http.StatusForbidden)
	}
}

func TestArtifactDownloadRejectsForeignAppOwnership(t *testing.T) {
	db, app := newAuthTestApp(t)
	devID := insertVerifiedUser(t, db, "dev@example.com", "old-password")
	insertVerifiedUser(t, db, "buyer@example.com", "old-password")
	adminID := insertVerifiedUser(t, db, "admin@example.com", "old-password")
	if _, err := db.Exec("UPDATE users SET role = 'admin' WHERE id = ?", adminID); err != nil {
		t.Fatalf("mark admin: %v", err)
	}
	if _, err := db.Exec("INSERT INTO developers(user_id, display_name) VALUES (?, ?)", devID, "Dev Two"); err != nil {
		t.Fatalf("insert developer: %v", err)
	}

	developerToken := loginToken(t, app, "dev@example.com", "old-password")
	buyerToken := loginToken(t, app, "buyer@example.com", "old-password")
	adminToken := loginToken(t, app, "admin@example.com", "old-password")

	appOne := publishAndApprove(t, app, developerToken, adminToken, "App One", "app-one")
	appTwo := publishAndApprove(t, app, developerToken, adminToken, "App Two", "app-two")

	// Upload an artifact to each app.
	upOne := uploadArtifactRequest(t, app, developerToken, appOne, "one.zip", "ONE")
	if upOne.Code != http.StatusCreated {
		t.Fatalf("upload app-one status = %d, want %d; body = %s", upOne.Code, http.StatusCreated, upOne.Body.String())
	}
	var upOneBody struct {
		Artifact struct {
			ID int64 `json:"id"`
		} `json:"artifact"`
	}
	if err := json.NewDecoder(upOne.Body).Decode(&upOneBody); err != nil {
		t.Fatalf("decode app-one upload: %v", err)
	}
	upTwo := uploadArtifactRequest(t, app, developerToken, appTwo, "two.zip", "TWO")
	if upTwo.Code != http.StatusCreated {
		t.Fatalf("upload app-two status = %d, want %d; body = %s", upTwo.Code, http.StatusCreated, upTwo.Body.String())
	}
	var upTwoBody struct {
		Artifact struct {
			ID int64 `json:"id"`
		} `json:"artifact"`
	}
	if err := json.NewDecoder(upTwo.Body).Decode(&upTwoBody); err != nil {
		t.Fatalf("decode app-two upload: %v", err)
	}

	// Buyer purchases app-one only.
	checkout := performJSONRequestWithToken(
		t, app, http.MethodPost, "/api/v1/marketplace/apps/app-one/checkout", `{}`, buyerToken,
	)
	if checkout.Code != http.StatusCreated {
		t.Fatalf("checkout status = %d, want %d; body = %s", checkout.Code, http.StatusCreated, checkout.Body.String())
	}
	var checkoutBody struct {
		Order struct {
			ID int64 `json:"id"`
		} `json:"order"`
	}
	if err := json.NewDecoder(checkout.Body).Decode(&checkoutBody); err != nil {
		t.Fatalf("decode checkout: %v", err)
	}
	confirm := performJSONRequestWithToken(
		t, app, http.MethodPost,
		"/api/v1/orders/"+strconv.FormatInt(checkoutBody.Order.ID, 10)+"/confirm-payment",
		`{}`, buyerToken,
	)
	if confirm.Code != http.StatusOK {
		t.Fatalf("confirm payment status = %d, want %d; body = %s", confirm.Code, http.StatusOK, confirm.Body.String())
	}

	myEnts := performJSONRequestWithToken(t, app, http.MethodGet, "/api/v1/me/entitlements", "", buyerToken)
	var myEntsBody struct {
		Entitlements []struct {
			ID int64 `json:"id"`
		} `json:"entitlements"`
	}
	if err := json.NewDecoder(myEnts.Body).Decode(&myEntsBody); err != nil {
		t.Fatalf("decode my entitlements: %v", err)
	}
	if len(myEntsBody.Entitlements) != 1 {
		t.Fatalf("my entitlements = %+v, want one", myEntsBody.Entitlements)
	}

	delivery := performJSONRequestWithToken(
		t, app, http.MethodGet,
		"/api/v1/entitlements/"+strconv.FormatInt(myEntsBody.Entitlements[0].ID, 10)+"/delivery",
		"", buyerToken,
	)
	if delivery.Code != http.StatusOK {
		t.Fatalf("delivery status = %d, want %d; body = %s", delivery.Code, http.StatusOK, delivery.Body.String())
	}
	var deliveryBody struct {
		DeliveryToken string `json:"deliveryToken"`
	}
	if err := json.NewDecoder(delivery.Body).Decode(&deliveryBody); err != nil {
		t.Fatalf("decode delivery: %v", err)
	}

	// Buyer can download app-one's artifact.
	dlOne := performJSONRequest(
		t, app, http.MethodGet,
		"/api/v1/artifacts/"+strconv.FormatInt(upOneBody.Artifact.ID, 10)+"/download?token="+deliveryBody.DeliveryToken,
		"",
	)
	if dlOne.Code != http.StatusOK || dlOne.Body.String() != "ONE" {
		t.Fatalf("download app-one artifact status = %d, body = %q, want 200/ONE", dlOne.Code, dlOne.Body.String())
	}

	// Buyer cannot download app-two's artifact (not entitled to that app).
	dlTwo := performJSONRequest(
		t, app, http.MethodGet,
		"/api/v1/artifacts/"+strconv.FormatInt(upTwoBody.Artifact.ID, 10)+"/download?token="+deliveryBody.DeliveryToken,
		"",
	)
	if dlTwo.Code != http.StatusForbidden {
		t.Fatalf("download foreign-app artifact status = %d, want %d", dlTwo.Code, http.StatusForbidden)
	}
}
