package uploader

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/eiserv/easySFTP/internal/config"
)

// lfLine is the shape 'ssh-keyscan <host> | ssh-keygen -lf -' prints for one
// key: bits, the SHA256 fingerprint, the host, and the key type in
// parentheses. The quick start tells users to store exactly this output, so
// host-key must accept it verbatim (issue #235).
func lfLine(srv *testServer) string {
	return fmt.Sprintf("256 %s %s (ED25519)", srv.HostKeySHA256, net.JoinHostPort(srv.Host, strconv.Itoa(srv.Port)))
}

// runWithHostKey runs one upload with the given host-key lines and returns
// the run's error, so every accepted/retransmitted shape below is exercised
// through the same path a user's config takes.
func runWithHostKey(t *testing.T, srv *testServer, fps []string) error {
	t.Helper()
	local := t.TempDir()
	writeTree(t, local, map[string]string{"a.txt": "pinned"})
	cfg := baseConfig(srv)
	cfg.AllowAnyHostKey = false
	cfg.HostKeyFingerprints = fps
	cfg.Uploads = []config.UploadPair{{Local: local, Remote: "/lf"}}
	_, err := Run(context.Background(), cfg, testLogger{t})
	return err
}

// hostKeySetupError asserts a rejected shape fails before any connection is
// attempted (the server counted zero accepts) and returns the error, so a
// future refactor cannot silently turn a bad host-key line into a connect
// attempt against the server.
func hostKeySetupError(t *testing.T, srv *testServer, fps []string) error {
	t.Helper()
	err := runWithHostKey(t, srv, fps)
	if err == nil {
		t.Fatalf("expected host-key %q to be rejected", strings.Join(fps, "\n"))
	}
	if got := atomic.LoadInt32(&srv.accepted); got != 0 {
		t.Fatalf("an invalid host-key line must fail before connecting; the server accepted %d connection(s)", got)
	}
	return err
}

// The rejection path is a permanent config error, not a transient network
// one: pinning the shape and asserting the error class keeps a future
// refactor from turning a bad line into a retryable connect error.
func TestHostKeyInvalidLineIsPermanentError(t *testing.T) {
	srv := startTestServer(t)
	err := hostKeySetupError(t, srv, []string{"garbage"})
	if !strings.Contains(err.Error(), "host-key must be a SHA256 fingerprint") {
		t.Fatalf("expected the SHA256 format error, got %v", err)
	}
	var perm permanentError
	if !errors.As(err, &perm) {
		t.Fatalf("a malformed host-key line must be a permanent error, got %v", err)
	}
}

func TestHostKeyAcceptsSshKeygenLfOutput(t *testing.T) {
	srv := startTestServer(t)

	t.Run("the documented quick-start command's output, verbatim", func(t *testing.T) {
		if err := runWithHostKey(t, srv, []string{lfLine(srv)}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("the whole multi-key output, every line", func(t *testing.T) {
		fps := []string{
			fmt.Sprintf("256 %s sftp.example.com (ED25519)", srv.HostKeySHA256),
			"256 SHA256:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA sftp.example.com (ECDSA)",
			fmt.Sprintf("3072 %s [sftp.example.com]:2222 (RSA)", srv.HostKeySHA256),
		}
		if err := runWithHostKey(t, srv, fps); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("bare fingerprint still works", func(t *testing.T) {
		if err := runWithHostKey(t, srv, []string{srv.HostKeySHA256}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("fingerprint pinned with leading spaces", func(t *testing.T) {
		if err := runWithHostKey(t, srv, []string{"   " + srv.HostKeySHA256 + "  "}); err != nil {
			t.Fatal(err)
		}
	})
}
func TestHostKeyRejectsWithGuidance(t *testing.T) {
	srv := startTestServer(t)
	addr := net.JoinHostPort(srv.Host, strconv.Itoa(srv.Port))
	keyLine := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(srv.HostPubKey)))

	t.Run("plain known_hosts line points at known-hosts", func(t *testing.T) {
		err := hostKeySetupError(t, srv, []string{addr + " " + keyLine})
		if !strings.Contains(err.Error(), "known_hosts line; use known-hosts instead") {
			t.Fatalf("expected the known-hosts redirect, got %v", err)
		}
	})

	t.Run("hashed known_hosts line points at known-hosts", func(t *testing.T) {
		hashed := knownhosts.HashHostname(knownhosts.Normalize(addr))
		err := hostKeySetupError(t, srv, []string{hashed + " " + keyLine})
		if !strings.Contains(err.Error(), "known_hosts line; use known-hosts instead") {
			t.Fatalf("expected the known-hosts redirect for a hashed entry, got %v", err)
		}
	})

	t.Run("MD5 fingerprint names the re-run command", func(t *testing.T) {
		err := hostKeySetupError(t, srv, []string{"256 MD5:26:06:5b:38:64:32:3e:a2:0f:ab:04:36:cd:0d:c4:44 sftp.example.com (ED25519)"})
		if !strings.Contains(err.Error(), "MD5 fingerprints are not accepted; re-run with 'ssh-keygen -E sha256 -lf'") {
			t.Fatalf("expected the MD5 guidance, got %v", err)
		}
	})

	t.Run("public key line names the conversion", func(t *testing.T) {
		err := hostKeySetupError(t, srv, []string{keyLine})
		if !strings.Contains(err.Error(), "public key line; convert it first") {
			t.Fatalf("expected the public-key guidance, got %v", err)
		}
	})

	t.Run("RSA public key line names the conversion too", func(t *testing.T) {
		err := hostKeySetupError(t, srv, []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQ... sftp@example"})
		if !strings.Contains(err.Error(), "public key line; convert it first") {
			t.Fatalf("expected the public-key guidance for an RSA line, got %v", err)
		}
	})

	t.Run("a server literally named ssh-gw is still a known_hosts host", func(t *testing.T) {
		err := hostKeySetupError(t, srv, []string{"ssh-gw.example.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI..."})
		if !strings.Contains(err.Error(), "known_hosts line; use known-hosts instead") {
			t.Fatalf("expected the known-hosts redirect despite the ssh- host name, got %v", err)
		}
	})

	t.Run("two fingerprints on one line are refused, not half-pinned", func(t *testing.T) {
		err := hostKeySetupError(t, srv, []string{srv.HostKeySHA256 + " " + srv.HostKeySHA256})
		if !strings.Contains(err.Error(), "takes one fingerprint per line") {
			t.Fatalf("expected the one-per-line refusal, got %v", err)
		}
	})

	t.Run("garbage keeps the format error", func(t *testing.T) {
		err := hostKeySetupError(t, srv, []string{"md5:abcdef"})
		if !strings.Contains(err.Error(), "must be a SHA256 fingerprint like 'SHA256:...'") {
			t.Fatalf("expected the format error for garbage, got %v", err)
		}
		if strings.Contains(err.Error(), "MD5:") {
			t.Fatalf("lowercase 'md5:' must not match the MD5 guidance, got %v", err)
		}
	})
}

// The jump host hop parses its host-key lines through the same helper with
// connection.proxy.* names; the proxy hop must accept -lf output too, and a
// bad line there must name the proxy option, not the primary one.
func TestProxyHostKeyAcceptsLfOutputAndNamesItsOptions(t *testing.T) {
	target := startTestServer(t)
	jump := startTestJumpServer(t)
	local := t.TempDir()
	writeTree(t, local, map[string]string{"index.html": "x"})

	jumpLf := fmt.Sprintf("256 %s jump.example.com (ED25519)", jump.HostKeySHA256)
	cfg := baseConfig(target)
	cfg.HostKeyFingerprints = []string{target.HostKeySHA256}
	cfg.Proxy = &config.Proxy{
		Server:              jump.Host,
		Port:                jump.Port,
		Username:            testUser,
		Password:            testPassword,
		HostKeyFingerprints: []string{jumpLf},
	}
	cfg.Uploads = []config.UploadPair{{Local: local, Remote: "/www"}}
	if _, err := Run(context.Background(), cfg, testLogger{t}); err != nil {
		t.Fatal(err)
	}

	// A known_hosts line pasted into the proxy's host-key must name
	// connection.proxy.known_hosts, proving the parse went through the
	// proxy hop's names and not the primary hop's.
	cfg.Proxy.HostKeyFingerprints = []string{jump.Host + " ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI..."}
	_, err := Run(context.Background(), cfg, testLogger{t})
	if err == nil || !strings.Contains(err.Error(), "connection.proxy.known_hosts") {
		t.Fatalf("expected the proxy known_hosts guidance, got %v", err)
	}
}
