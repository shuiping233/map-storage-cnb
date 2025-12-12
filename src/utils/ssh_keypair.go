package utils

import (
	"crypto"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	goGitSSH "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"golang.org/x/crypto/ed25519"
	cryptoSSH "golang.org/x/crypto/ssh"
)

const (
	SSHPrivateKeyName = "id_ed25519"
	SSHPublicKeyName  = "id_ed25519.pub"
	SSHKeyPairDir     = "./test_ssh_key"
	GitWorkSpaceDir   = "./test_space"
)

// Generate ed25519 KeyPair,
// return (private key, public key, error)
func GenerateSSHKey() ([]byte, []byte, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	pemBlock, err := cryptoSSH.MarshalPrivateKey(crypto.PrivateKey(priv), "")
	if err != nil {
		panic(err)
	}

	// privKeyBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	// if err != nil {
	// 	return nil, nil, err
	// }
	// privBlock := &pem.Block{
	// 	Type:  "PRIVATE KEY",
	// 	Bytes: privKeyBytes,
	// }

	privResult := pem.EncodeToMemory(pemBlock)

	pubSSH, err := cryptoSSH.NewPublicKey(pub)
	if err != nil {
		return nil, nil, err
	}

	pubResult := cryptoSSH.MarshalAuthorizedKey(pubSSH)

	return privResult, pubResult, nil

	// // If rand is nil, crypto/rand.Reader will be used
	// pub, priv, err := ed25519.GenerateKey(nil)
	// if err != nil {
	// 	panic(err)
	// }

}

func WriteSSHKeyPair(dirPath string, priv []byte, pub []byte) error {

	if !DirExists(dirPath) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return err
		}
	}

	privPath := filepath.Join(dirPath, SSHPrivateKeyName)
	pubPath := filepath.Join(dirPath, SSHPublicKeyName)

	if err := os.WriteFile(privPath, priv, 0600); err != nil {
		return err
	}

	if err := os.WriteFile(pubPath, pub, 0644); err != nil {
		return err
	}

	fmt.Printf("Writing SSH KeyPair to %q and %q ", privPath, pubPath)
	return nil
}

func LoadSSHKeyPair(dirPath string) ([]byte, []byte, error) {

	privPath := filepath.Join(dirPath, SSHPrivateKeyName)
	pubPath := filepath.Join(dirPath, SSHPublicKeyName)

	fmt.Printf("Loading SSH KeyPair from %q and %q ", privPath, pubPath)

	if !FileExists(privPath) {
		return nil, nil, fmt.Errorf("%q not exists", privPath)
	}
	if !FileExists(pubPath) {
		return nil, nil, fmt.Errorf("%q not exists", pubPath)
	}

	privResult, err := os.ReadFile(privPath)
	if err != nil {
		return nil, nil, err
	}

	pubResult, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, nil, err
	}
	return privResult, pubResult, nil
}

func SSHKeyTest(remoteGitRepoUrl string) {
	priv, _, err := LoadSSHKeyPair(SSHKeyPairDir)
	if err != nil {
		log.Printf("Loading SSH Key Error : %v", err)

		priv, pub, err := GenerateSSHKey()
		if err != nil {
			log.Printf("Gen SSH Key Error : %v", err)
		}
		err = WriteSSHKeyPair(SSHKeyPairDir, priv, pub)
		if err != nil {
			log.Printf("Save SSH Key Error : %v", err)
		}
	}

	// sshAuth, err := goGitSSH.NewPublicKeysFromFile("test", filepath.Join(SSHKeyPairDir, SSHPrivateKeyName), "")
	sshAuth, err := goGitSSH.NewPublicKeys("test", priv, "")
	if err != nil {
		log.Printf("NewPublicKeys Error : %v", err)
	}
	sshAuth.ClientConfig()
	_, err = git.PlainClone(GitWorkSpaceDir, false, &git.CloneOptions{
		URL:      remoteGitRepoUrl,
		Auth:     sshAuth,
		Progress: os.Stdout,
	})

	if err != nil {
		log.Printf("Error cloning repository: %v", err)
	}

}
