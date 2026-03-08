package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

var (
	tenantID      = mustEnv("ENTRA_TENANT_ID")
	mcpClientID   = mustEnv("MCP_SERVER_CLIENT_ID")
	listenAddr    = envOr("LISTEN_ADDR", ":3001")
	jwksURL       = fmt.Sprintf("https://login.microsoftonline.com/%s/discovery/v2.0/keys", tenantID)
	issuer        = fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", tenantID)
	requiredScope = "mcp.tools"
)

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return v
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ---------------------------------------------------------------------------
// Per-user notes store
// ---------------------------------------------------------------------------

type Note struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type NotesStore struct {
	mu    sync.RWMutex
	notes map[string][]Note // keyed by user sub claim
}

func NewNotesStore() *NotesStore {
	return &NotesStore{notes: make(map[string][]Note)}
}

func (s *NotesStore) Save(userSub, title, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notes[userSub] = append(s.notes[userSub], Note{
		Title:     title,
		Content:   content,
		CreatedAt: time.Now().Format(time.RFC3339),
	})
}

func (s *NotesStore) List(userSub string) []Note {
	s.mu.RLock()
	defer s.mu.RUnlock()
	notes := s.notes[userSub]
	if notes == nil {
		return []Note{}
	}
	return notes
}

func (s *NotesStore) Delete(userSub, title string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	notes := s.notes[userSub]
	for i, n := range notes {
		if n.Title == title {
			s.notes[userSub] = append(notes[:i], notes[i+1:]...)
			return true
		}
	}
	return false
}

var store = NewNotesStore()

// ---------------------------------------------------------------------------
// JWT verification for Entra ID
// ---------------------------------------------------------------------------

func newJWKSVerifier() auth.TokenVerifier {
	// Create a keyfunc that fetches and caches JWKS from Entra.
	jwks, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		log.Fatalf("failed to create JWKS keyfunc: %v", err)
	}

	appIDURI := fmt.Sprintf("api://%s", mcpClientID)

	return func(ctx context.Context, tokenString string, _ *http.Request) (*auth.TokenInfo, error) {
		// Parse with Entra's JWKS for RSA signature validation.
		// Accept both v1 and v2 issuers from Entra.
		issuerV1 := fmt.Sprintf("https://sts.windows.net/%s/", tenantID)
		token, err := jwt.Parse(tokenString, jwks.Keyfunc,
			jwt.WithExpirationRequired(),
		)
		if err != nil {
			log.Printf("JWT parse error: %v", err)
			return nil, fmt.Errorf("%w: %v", auth.ErrInvalidToken, err)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			return nil, fmt.Errorf("%w: invalid claims", auth.ErrInvalidToken)
		}

		// Validate issuer — accept both v1 and v2 format.
		tokenIssuer := claimStr(claims, "iss")
		if tokenIssuer != issuer && tokenIssuer != issuerV1 {
			log.Printf("JWT issuer mismatch: got %q, want %q or %q", tokenIssuer, issuer, issuerV1)
			return nil, fmt.Errorf("%w: invalid issuer %q", auth.ErrInvalidToken, tokenIssuer)
		}

		// Validate audience manually to accept both client ID and App ID URI.
		aud := claimStr(claims, "aud")
		if aud != mcpClientID && aud != appIDURI {
			log.Printf("JWT audience mismatch: got %q, want %q or %q", aud, mcpClientID, appIDURI)
			return nil, fmt.Errorf("%w: invalid audience %q", auth.ErrInvalidToken, aud)
		}

		// Extract scopes — Entra puts them in "scp" as a space-separated string.
		scopes := extractScopes(claims)

		// Extract user identity claims.
		extra := map[string]any{
			"preferred_username": claimStr(claims, "preferred_username"),
			"name":              claimStr(claims, "name"),
			"sub":               claimStr(claims, "sub"),
		}

		exp, _ := token.Claims.GetExpirationTime()

		return &auth.TokenInfo{
			Scopes:     scopes,
			Expiration: exp.Time,
			UserID:     claimStr(claims, "sub"),
			Extra:      extra,
		}, nil
	}
}

func extractScopes(claims jwt.MapClaims) []string {
	scp, ok := claims["scp"].(string)
	if !ok || scp == "" {
		return nil
	}
	return strings.Fields(scp)
}

func claimStr(claims jwt.MapClaims, key string) string {
	v, _ := claims[key].(string)
	return v
}

// ---------------------------------------------------------------------------
// Tool argument types
// ---------------------------------------------------------------------------

type echoArgs struct {
	Message string `json:"message" jsonschema:"the message to echo back"`
}

type saveNoteArgs struct {
	Title   string `json:"title" jsonschema:"title of the note"`
	Content string `json:"content" jsonschema:"content of the note"`
}

type deleteNoteArgs struct {
	Title string `json:"title" jsonschema:"title of the note to delete"`
}

// ---------------------------------------------------------------------------
// Tool handlers
// ---------------------------------------------------------------------------

func echoHandler(_ context.Context, req *mcp.CallToolRequest, args echoArgs) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: args.Message}},
	}, nil, nil
}

func getServerTimeHandler(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: time.Now().Format(time.RFC3339)}},
	}, nil, nil
}

func whoamiHandler(_ context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	info := req.Extra.TokenInfo
	data := map[string]any{
		"preferred_username": info.Extra["preferred_username"],
		"name":              info.Extra["name"],
		"sub":               info.Extra["sub"],
		"scopes":            info.Scopes,
	}
	out, _ := json.MarshalIndent(data, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(out)}},
	}, nil, nil
}

func saveNoteHandler(_ context.Context, req *mcp.CallToolRequest, args saveNoteArgs) (*mcp.CallToolResult, any, error) {
	userSub := req.Extra.TokenInfo.UserID
	store.Save(userSub, args.Title, args.Content)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Saved note '%s'.", args.Title)}},
	}, nil, nil
}

func listNotesHandler(_ context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	userSub := req.Extra.TokenInfo.UserID
	notes := store.List(userSub)
	out, _ := json.MarshalIndent(notes, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(out)}},
	}, nil, nil
}

func deleteNoteHandler(_ context.Context, req *mcp.CallToolRequest, args deleteNoteArgs) (*mcp.CallToolResult, any, error) {
	userSub := req.Extra.TokenInfo.UserID
	if store.Delete(userSub, args.Title) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Deleted note '%s'.", args.Title)}},
		}, nil, nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Note '%s' not found.", args.Title)}},
	}, nil, nil
}

// ---------------------------------------------------------------------------
// Server setup
// ---------------------------------------------------------------------------

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "homelab-mcp-server"}, nil)

	// Register tools.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "echo",
		Description: "Echo a message back",
	}, echoHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_server_time",
		Description: "Get the current server time",
	}, getServerTimeHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "whoami",
		Description: "Get the authenticated user's identity from the JWT",
	}, whoamiHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "save_note",
		Description: "Save a personal note (scoped to the authenticated user)",
	}, saveNoteHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_notes",
		Description: "List all personal notes for the authenticated user",
	}, listNotesHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_note",
		Description: "Delete a personal note by title",
	}, deleteNoteHandler)

	// Streamable HTTP handler.
	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return server
	}, nil)

	// JWT auth middleware.
	verifier := newJWKSVerifier()
	resourceMetadataURL := fmt.Sprintf("http://localhost%s/.well-known/oauth-protected-resource", listenAddr)
	jwtAuth := auth.RequireBearerToken(verifier, &auth.RequireBearerTokenOptions{
		Scopes:              []string{requiredScope},
		ResourceMetadataURL: resourceMetadataURL,
	})

	// Protected resource metadata (unauthenticated).
	metadata := &oauthex.ProtectedResourceMetadata{
		Resource: fmt.Sprintf("http://localhost%s", listenAddr),
		AuthorizationServers: []string{
			fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", tenantID),
		},
		ScopesSupported:        []string{requiredScope},
		BearerMethodsSupported: []string{"header"},
	}

	mux := http.NewServeMux()
	mux.Handle("/.well-known/oauth-protected-resource", auth.ProtectedResourceMetadataHandler(metadata))
	mux.Handle("/mcp", jwtAuth(handler))

	log.Printf("MCP server listening on %s", listenAddr)
	log.Printf("Tools: echo, get_server_time, whoami, save_note, list_notes, delete_note")
	log.Fatal(http.ListenAndServe(listenAddr, mux))
}
