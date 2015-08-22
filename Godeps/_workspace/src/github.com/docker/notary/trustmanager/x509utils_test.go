package trustmanager

import (
	"crypto/rand"
	"crypto/x509"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCertsToKeys(t *testing.T) {
	// Get root certificate
	rootCA, err := LoadCertFromFile("../fixtures/root-ca.crt")
	assert.NoError(t, err)

	// Get intermediate certificate
	intermediateCA, err := LoadCertFromFile("../fixtures/intermediate-ca.crt")
	assert.NoError(t, err)

	// Get leaf certificate
	leafCert, err := LoadCertFromFile("../fixtures/secure.example.com.crt")
	assert.NoError(t, err)

	// Get our certList with Leaf Cert and Intermediate
	certList := []*x509.Certificate{leafCert, intermediateCA, rootCA}

	// Call CertsToKEys
	keys := CertsToKeys(certList)
	assert.NotNil(t, keys)
	assert.Len(t, keys, 3)

	// Call GetLeafCerts
	newKeys := GetLeafCerts(certList)
	assert.NotNil(t, newKeys)
	assert.Len(t, newKeys, 1)

	// Call GetIntermediateCerts (checks for certs with IsCA true)
	newKeys = GetIntermediateCerts(certList)
	assert.NotNil(t, newKeys)
	assert.Len(t, newKeys, 2)
}

func TestNewCertificate(t *testing.T) {
	cert, err := NewCertificate("docker.com/alpine")
	assert.NoError(t, err)
	assert.Equal(t, cert.Subject.CommonName, "docker.com/alpine")
	assert.True(t, time.Now().Before(cert.NotAfter))
	assert.True(t, time.Now().AddDate(10, 0, 1).After(cert.NotAfter))
}

func TestKeyOperations(t *testing.T) {
	// Generate our ED25519 private key
	edKey, err := GenerateED25519Key(rand.Reader)
	assert.NoError(t, err)

	// Generate our EC private key
	ecKey, err := GenerateECDSAKey(rand.Reader)
	assert.NoError(t, err)

	// Generate our RSA private key
	rsaKey, err := GenerateRSAKey(rand.Reader, 512)

	// Encode our ED private key
	edPEM, err := KeyToPEM(edKey)
	assert.NoError(t, err)

	// Encode our EC private key
	ecPEM, err := KeyToPEM(ecKey)
	assert.NoError(t, err)

	// Encode our RSA private key
	rsaPEM, err := KeyToPEM(rsaKey)
	assert.NoError(t, err)

	// Check to see if ED key it is encoded
	stringEncodedEDKey := string(edPEM)
	assert.True(t, strings.Contains(stringEncodedEDKey, "-----BEGIN ED25519 PRIVATE KEY-----"))

	// Check to see if EC key it is encoded
	stringEncodedECKey := string(ecPEM)
	assert.True(t, strings.Contains(stringEncodedECKey, "-----BEGIN EC PRIVATE KEY-----"))

	// Check to see if RSA key it is encoded
	stringEncodedRSAKey := string(rsaPEM)
	assert.True(t, strings.Contains(stringEncodedRSAKey, "-----BEGIN RSA PRIVATE KEY-----"))

	// Decode our ED Key
	decodedEDKey, err := ParsePEMPrivateKey(edPEM, "")
	assert.NoError(t, err)
	assert.Equal(t, edKey.Private(), decodedEDKey.Private())

	// Decode our EC Key
	decodedECKey, err := ParsePEMPrivateKey(ecPEM, "")
	assert.NoError(t, err)
	assert.Equal(t, ecKey.Private(), decodedECKey.Private())

	// Decode our RSA Key
	decodedRSAKey, err := ParsePEMPrivateKey(rsaPEM, "")
	assert.NoError(t, err)
	assert.Equal(t, rsaKey.Private(), decodedRSAKey.Private())

	// Encrypt our ED Key
	encryptedEDKey, err := EncryptPrivateKey(edKey, "ponies")
	assert.NoError(t, err)

	// Encrypt our EC Key
	encryptedECKey, err := EncryptPrivateKey(ecKey, "ponies")
	assert.NoError(t, err)

	// Encrypt our RSA Key
	encryptedRSAKey, err := EncryptPrivateKey(rsaKey, "ponies")
	assert.NoError(t, err)

	// Check to see if ED key it is encrypted
	stringEncryptedEDKey := string(encryptedEDKey)
	assert.True(t, strings.Contains(stringEncryptedEDKey, "-----BEGIN ED25519 PRIVATE KEY-----"))
	assert.True(t, strings.Contains(stringEncryptedEDKey, "Proc-Type: 4,ENCRYPTED"))

	// Check to see if EC key it is encrypted
	stringEncryptedECKey := string(encryptedECKey)
	assert.True(t, strings.Contains(stringEncryptedECKey, "-----BEGIN EC PRIVATE KEY-----"))
	assert.True(t, strings.Contains(stringEncryptedECKey, "Proc-Type: 4,ENCRYPTED"))

	// Check to see if RSA key it is encrypted
	stringEncryptedRSAKey := string(encryptedRSAKey)
	assert.True(t, strings.Contains(stringEncryptedRSAKey, "-----BEGIN RSA PRIVATE KEY-----"))
	assert.True(t, strings.Contains(stringEncryptedRSAKey, "Proc-Type: 4,ENCRYPTED"))

	// Decrypt our ED Key
	decryptedEDKey, err := ParsePEMPrivateKey(encryptedEDKey, "ponies")
	assert.NoError(t, err)
	assert.Equal(t, edKey.Private(), decryptedEDKey.Private())

	// Decrypt our EC Key
	decryptedECKey, err := ParsePEMPrivateKey(encryptedECKey, "ponies")
	assert.NoError(t, err)
	assert.Equal(t, ecKey.Private(), decryptedECKey.Private())

	// Decrypt our RSA Key
	decryptedRSAKey, err := ParsePEMPrivateKey(encryptedRSAKey, "ponies")
	assert.NoError(t, err)
	assert.Equal(t, rsaKey.Private(), decryptedRSAKey.Private())

}
