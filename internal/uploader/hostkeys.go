package uploader

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"os"
	"slices"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// hostKeyCallback builds the host key verifier for one hop. It also returns
// the host key algorithms that can be derived from the hop's known-hosts
// lines, for the caller to use when the user did not configure any (see
// pinnedHostKeyAlgorithms); the returned list is nil whenever there is
// nothing to derive.
func hostKeyCallback(h hop, log Logger) (ssh.HostKeyCallback, []string, error) {
	want := h.fingerprints
	if len(want) == 0 && h.knownHosts == "" {
		// Unverified connections are an explicit opt-in in v3, per hop: a
		// pinned target behind an unpinned jump host (or vice versa) still
		// fails for the open hop unless that hop opts out too.
		if !h.allowAnyHostKey {
			return nil, nil, fmt.Errorf("the identity of %[1]s cannot be verified: no %[2]s or %[3]s configured. "+
				"Pin the server's keys (run 'ssh-keyscan <server>' and set %[3]s, or convert with 'ssh-keygen -lf -' and set %[2]s), "+
				"or explicitly accept any host key with %[4]s (NOT recommended: allows man-in-the-middle attacks)",
				h.addr, h.names.hostKey, h.names.knownHosts, h.names.allowAny)
		}
		log.Warningf("%s is set; the identity of %s will NOT be verified and man-in-the-middle attacks are possible. "+
			"Pin the server's keys via %s or %s and remove the opt-out.",
			h.names.allowAny, h.addr, h.names.hostKey, h.names.knownHosts)
		return ssh.InsecureIgnoreHostKey(), nil, nil
	}
	for _, fp := range want {
		if !strings.HasPrefix(fp, "SHA256:") {
			return nil, nil, fmt.Errorf("%s must be a SHA256 fingerprint like 'SHA256:...', got %q", h.names.hostKey, fp)
		}
	}
	var khCallback ssh.HostKeyCallback
	if h.knownHosts != "" {
		var err error
		if khCallback, err = knownHostsCallback(h.knownHosts); err != nil {
			return nil, nil, err
		}
	}
	// A key matching either input is accepted, mirroring how multiple
	// fingerprints already OR together: users can pin every key their server
	// presents, in whichever format they have at hand.
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		got := ssh.FingerprintSHA256(key)
		for _, fp := range want {
			if got == fp {
				return nil
			}
		}
		if khCallback != nil && khCallback(hostname, remote, key) == nil {
			return nil
		}
		accepted := append([]string{}, want...)
		if khCallback != nil {
			accepted = append(accepted, "the known-hosts entries")
		}
		// permanentError: a mismatch is a security signal (or a config error),
		// and retrying the connection would present the same key again.
		return permanentError{hostKeyMismatchError(hostname, got, key, accepted, h.names.knownHosts, khCallback == nil)}
	}, pinnedHostKeyAlgorithms(h, khCallback), nil
}

// hostKeyMismatchError reports a rejected host key, including the one fact
// that lets a user tell the common benign cause from a man-in-the-middle
// themselves: the *type* of the key the server presented. A stock OpenSSH
// server has one key of each type and presents whichever the client asks
// for, so "got an ecdsa key while I pinned the ed25519 one" points at a
// different key of the same server, not an impostor.
//
// The hint is added only when the pins are bare fingerprints: fingerprints
// carry no key type, so easySFTP cannot ask the server for the pinned type.
// With known-hosts it can (and does, see pinnedHostKeyAlgorithms), so there
// a mismatch means a key of the pinned type really is different, which is
// the man-in-the-middle signal and must not be softened.
func hostKeyMismatchError(hostname, got string, key ssh.PublicKey, accepted []string, knownHostsOption string, fingerprintsOnly bool) error {
	msg := fmt.Sprintf("host key mismatch for %s: got %s (%s), want one of: %s",
		hostname, got, key.Type(), strings.Join(accepted, ", "))
	if fingerprintsOnly {
		msg += fmt.Sprintf(". The server presented its %s key; if you pinned only one of this server's keys, "+
			"pin this one too, or pin all of its keys, or use %s, which asks the server for the key types you pinned",
			key.Type(), knownHostsOption)
	}
	return errors.New(msg)
}

// pinnedHostKeyAlgorithms derives the host key algorithms the handshake
// should offer from the key types of the known-hosts lines that match this
// hop's address. x/crypto's default preference lists Ed25519 last, so against
// a stock multi-key server the handshake asks for the ECDSA key even when the
// user pinned only the Ed25519 line (issue #282); offering the pinned types
// makes the server present a key the user actually has.
//
// Derivation only happens when known-hosts is the hop's sole pin material:
// combined with fingerprints it could narrow the negotiation away from the
// fingerprint's key type (fingerprints carry no type), which turns today's
// mismatch into a "no common algorithm" handshake failure on servers that
// lack the known-hosts key types.
//
// The probe calls the callback with a throwaway key, which can never match;
// x/crypto answers with a *knownhosts.KeyError whose Want lists every line
// matching the address, of every key type. Both the plain form of each
// pinned type and its certificate form are offered, certificate first, the
// same relative order x/crypto's defaults use: the file's @cert-authority
// marker is not visible through KeyError (keyDBLine.cert is private), so a
// pinned CA key and a pinned plain key cannot be told apart here and both
// negotiation forms are offered for either. Anything
// x/crypto does not support (or classifies as insecure, like ssh-rsa) is
// dropped; users who need those configure algorithms.host_key_algorithms,
// which always wins over this derivation.
func pinnedHostKeyAlgorithms(h hop, khCallback ssh.HostKeyCallback) []string {
	if len(h.fingerprints) > 0 || khCallback == nil {
		return nil
	}
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil
	}
	probeKey, err := ssh.NewPublicKey(priv.Public())
	if err != nil {
		return nil
	}
	probeErr := khCallback(h.addr, stringAddr(h.addr), probeKey)
	var keyErr *knownhosts.KeyError
	if !errors.As(probeErr, &keyErr) {
		// No line matches this address (the handshake will report that), or
		// something unexpected happened: keep x/crypto's defaults.
		return nil
	}
	var algos []string
	for _, known := range keyErr.Want {
		for _, algo := range algorithmsForKeyType(known.Key.Type()) {
			if !slices.Contains(supportedHostKeyAlgorithms, algo) {
				continue
			}
			if !slices.Contains(algos, algo) {
				algos = append(algos, algo)
			}
		}
	}
	if len(algos) == 0 {
		return nil
	}
	return algos
}

// algorithmsForKeyType maps a known_hosts key type to the host key algorithms
// that negotiate a key of that type, in the order x/crypto's own defaults
// prefer them (certificate form first). Unsupported types map to themselves
// so the caller's supported-filter can drop them.
func algorithmsForKeyType(keyType string) []string {
	switch keyType {
	case ssh.KeyAlgoRSA, ssh.CertAlgoRSAv01:
		return []string{ssh.CertAlgoRSASHA256v01, ssh.CertAlgoRSASHA512v01, ssh.KeyAlgoRSASHA256, ssh.KeyAlgoRSASHA512}
	case ssh.KeyAlgoECDSA256:
		return []string{ssh.CertAlgoECDSA256v01, ssh.KeyAlgoECDSA256}
	case ssh.KeyAlgoECDSA384:
		return []string{ssh.CertAlgoECDSA384v01, ssh.KeyAlgoECDSA384}
	case ssh.KeyAlgoECDSA521:
		return []string{ssh.CertAlgoECDSA521v01, ssh.KeyAlgoECDSA521}
	case ssh.KeyAlgoED25519:
		return []string{ssh.CertAlgoED25519v01, ssh.KeyAlgoED25519}
	default:
		return []string{keyType}
	}
}

// supportedHostKeyAlgorithms is every host key algorithm x/crypto implements
// and does not classify as insecure, i.e. exactly what the handshake can be
// asked to negotiate without an explicit algorithms.host_key_algorithms
// opt-in. Computed once from ssh.SupportedAlgorithms, so it stays in sync
// with the library without hardcoding the list.
var supportedHostKeyAlgorithms = ssh.SupportedAlgorithms().HostKeys

// stringAddr turns an address string into a net.Addr for the known-hosts
// probe, mirroring what the handshake passes the callback.
type stringAddr string

func (a stringAddr) Network() string { return "tcp" }
func (a stringAddr) String() string  { return string(a) }

// knownHostsCallback builds a host key verifier from raw OpenSSH known_hosts
// lines (e.g. ssh-keyscan output). x/crypto's knownhosts parser only reads
// files, so the lines are staged in a temp file that is removed again right
// after parsing.
func knownHostsCallback(data string) (ssh.HostKeyCallback, error) {
	f, err := os.CreateTemp("", "easysftp-known-hosts-*")
	if err != nil {
		return nil, fmt.Errorf("staging known-hosts: %w", err)
	}
	defer os.Remove(f.Name())
	_, werr := f.WriteString(data + "\n")
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		return nil, fmt.Errorf("staging known-hosts: %w", werr)
	}
	cb, err := knownhosts.New(f.Name())
	if err != nil {
		return nil, fmt.Errorf("parsing known-hosts: %w", err)
	}
	return cb, nil
}
