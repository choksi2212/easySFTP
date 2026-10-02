package uploader

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"net"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/eiserv/easySFTP/internal/config"
)

// multiKeyServer is the issue #282 fixture: one server, two host key types.
// The Ed25519 key is the one the user pinned (srv.HostPubKey); the ECDSA key
// is the one Go's default client preference would make the server present.
func multiKeyServer(t *testing.T) (*testServer, string) {
	t.Helper()
	srv := startTestServer(t, withExtraECDSAHostKey())
	return srv, net.JoinHostPort(srv.Host, strconv.Itoa(srv.Port))
}

// A user who pins only the Ed25519 line of a multi-key server must still
// connect: easySFTP asks the server for the key types the user pinned.
func TestKnownHostsSinglePinnedKeyOnMultiKeyServer(t *testing.T) {
	srv, addr := multiKeyServer(t)
	local := t.TempDir()
	writeTree(t, local, map[string]string{"a.txt": "one"})

	run := func(t *testing.T, knownHosts string) {
		t.Helper()
		cfg := baseConfig(srv)
		cfg.AllowAnyHostKey = false
		cfg.KnownHosts = knownHosts
		cfg.Uploads = []config.UploadPair{{Local: local, Remote: "/single"}}
		if _, err := Run(context.Background(), cfg, testLogger{t}); err != nil {
			t.Fatal(err)
		}
	}

	// What "ssh-keyscan <host>" prints for one of the server's keys.
	t.Run("keyscan line for the pinned key", func(t *testing.T) {
		run(t, knownhosts.Line([]string{addr}, srv.HostPubKey))
	})
	// The other shape ssh-keyscan prints: hashed hostname.
	t.Run("hashed entry for the pinned key", func(t *testing.T) {
		hashed := knownhosts.HashHostname(knownhosts.Normalize(addr))
		run(t, hashed+" "+strings.TrimSpace(string(ssh.MarshalAuthorizedKey(srv.HostPubKey))))
	})
}

// An explicit algorithms.host_key_algorithms list stays authoritative: it
// can still ask for a type the user did not pin, which fails as before.
func TestExplicitHostKeyAlgorithmsBeatDerivation(t *testing.T) {
	srv, addr := multiKeyServer(t)
	local := t.TempDir()
	writeTree(t, local, map[string]string{"a.txt": "one"})

	cfg := baseConfig(srv)
	cfg.AllowAnyHostKey = false
	cfg.KnownHosts = knownhosts.Line([]string{addr}, srv.HostPubKey)
	// Ask for the ECDSA key the user did not pin, plus the pinned type last:
	// the server happily presents its ECDSA key first.
	cfg.Algorithms.HostKeyAlgorithms = []string{ssh.KeyAlgoECDSA256, ssh.KeyAlgoED25519}
	cfg.Uploads = []config.UploadPair{{Local: local, Remote: "/explicit"}}
	_, err := Run(context.Background(), cfg, testLogger{t})
	if err == nil || !strings.Contains(err.Error(), "host key mismatch") {
		t.Fatalf("expected a mismatch with an explicit algorithm list, got %v", err)
	}
	// The server must actually have presented the ECDSA key, i.e. the
	// explicit list really was honored (not silently derived around).
	if !strings.Contains(err.Error(), ssh.KeyAlgoECDSA256) {
		t.Fatalf("expected the error to name the presented key's type, got %v", err)
	}
}

// The real security property must survive: a server presenting a key that
// matches none of the pins still fails, and the error now names the
// presented key's type so the mismatch is diagnosable.
func TestMultiKeyServerRealMismatchStillFails(t *testing.T) {
	srv, _ := multiKeyServer(t)
	local := t.TempDir()
	writeTree(t, local, map[string]string{"a.txt": "one"})

	_, otherPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	otherPub, err := ssh.NewPublicKey(otherPriv.Public())
	if err != nil {
		t.Fatal(err)
	}

	cfg := baseConfig(srv)
	cfg.AllowAnyHostKey = false
	cfg.HostKeyFingerprints = []string{ssh.FingerprintSHA256(otherPub)}
	cfg.Uploads = []config.UploadPair{{Local: local, Remote: "/mismatch"}}
	_, err = Run(context.Background(), cfg, testLogger{t})
	if err == nil || !strings.Contains(err.Error(), "host key mismatch") {
		t.Fatalf("expected a host key mismatch, got %v", err)
	}
	// The presented key's type is the fact that separates "another key of
	// this server" from "an impostor".
	if !strings.Contains(err.Error(), "("+ssh.KeyAlgoECDSA256+")") {
		t.Fatalf("expected the mismatch error to name the presented key type, got %v", err)
	}
	// Fingerprint-only pins cannot ask for a type, so the error should also
	// tell the user what to pin next.
	if !strings.Contains(err.Error(), "pin this one too") {
		t.Fatalf("expected the fingerprint-only hint in the error, got %v", err)
	}
}

// Derivation must not leak: known-hosts lines for a different host do not
// influence the algorithms offered to this server.
func TestDerivationIgnoresOtherHostsEntries(t *testing.T) {
	srv, _ := multiKeyServer(t)
	local := t.TempDir()
	writeTree(t, local, map[string]string{"a.txt": "one"})

	cfg := baseConfig(srv)
	cfg.AllowAnyHostKey = false
	// Only a line for some other host: the connection must fail (nothing is
	// pinned for this one), not succeed by accident.
	cfg.KnownHosts = knownhosts.Line([]string{"other.example.com:22"}, srv.HostPubKey)
	cfg.Uploads = []config.UploadPair{{Local: local, Remote: "/leak"}}
	_, err := Run(context.Background(), cfg, testLogger{t})
	if err == nil || !strings.Contains(err.Error(), "host key mismatch") {
		t.Fatalf("expected a mismatch, got %v", err)
	}
}

// A pinned ECDSA line derives ECDSA algorithms even though the server also
// has an Ed25519 key: the client asks for the pinned type.
func TestPinnedECDSAKeyDerivesECDSAAlgorithms(t *testing.T) {
	srv, addr := multiKeyServer(t)
	if len(srv.ExtraHostKeys) == 0 {
		t.Fatal("fixture did not add an extra host key")
	}
	local := t.TempDir()
	writeTree(t, local, map[string]string{"a.txt": "one"})

	cfg := baseConfig(srv)
	cfg.AllowAnyHostKey = false
	cfg.KnownHosts = knownhosts.Line([]string{addr}, srv.ExtraHostKeys[0])
	cfg.Uploads = []config.UploadPair{{Local: local, Remote: "/ecdsa"}}
	if _, err := Run(context.Background(), cfg, testLogger{t}); err != nil {
		t.Fatal(err)
	}
}
