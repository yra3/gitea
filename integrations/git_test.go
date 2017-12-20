// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"code.gitea.io/git"
	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/setting"
	api "code.gitea.io/sdk/gitea"

	"github.com/Unknwon/com"
	"github.com/stretchr/testify/assert"
)

func onGiteaRun(t *testing.T, callback func(*testing.T, *url.URL)) {
	s := http.Server{
		Handler: mac,
	}

	u, err := url.Parse(setting.AppURL)
	assert.NoError(t, err)
	listener, err := net.Listen("tcp", u.Host)
	assert.NoError(t, err)

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		s.Shutdown(ctx)
		cancel()
	}()

	go s.Serve(listener)

	//TODO add SSH internal server

	callback(t, u)
}

func generateCommit(repoPath, email, fullName string) error {
	//Generate random file
	data := make([]byte, 1024)
	_, err := rand.Read(data)
	if err != nil {
		return err
	}
	tmpFile, err := ioutil.TempFile(repoPath, "data-file-")
	if err != nil {
		return err
	}
	defer tmpFile.Close()
	_, err = tmpFile.Write(data)
	if err != nil {
		return err
	}

	//Commit
	err = git.AddChanges(repoPath, false, filepath.Base(tmpFile.Name()))
	if err != nil {
		return err
	}
	err = git.CommitChanges(repoPath, git.CommitChangesOptions{
		Committer: &git.Signature{
			Email: email,
			Name:  fullName,
			When:  time.Now(),
		},
		Author: &git.Signature{
			Email: email,
			Name:  fullName,
			When:  time.Now(),
		},
		Message: fmt.Sprintf("Testing commit @ %v", time.Now()),
	})
	return err
}

func TestGit(t *testing.T) {
	prepareTestEnv(t)

	onGiteaRun(t, func(t *testing.T, u *url.URL) {
		u.Path = "user2/repo1.git"

		t.Run("HTTP", func(t *testing.T) {
			dstPath, err := ioutil.TempDir("", "repo-tmp-17")
			assert.NoError(t, err)
			defer os.RemoveAll(dstPath)
			t.Run("Standard", func(t *testing.T) {
				t.Run("CloneNoLogin", func(t *testing.T) {
					dstLocalPath, err := ioutil.TempDir("", "repo1")
					assert.NoError(t, err)
					defer os.RemoveAll(dstLocalPath)
					err = git.Clone(u.String(), dstLocalPath, git.CloneRepoOptions{})
					assert.NoError(t, err)
					assert.True(t, com.IsExist(filepath.Join(dstLocalPath, "README.md")))
				})

				t.Run("CreateRepo", func(t *testing.T) {
					session := loginUser(t, "user2")
					req := NewRequestWithJSON(t, "POST", "/api/v1/user/repos", &api.CreateRepoOption{
						AutoInit:    true,
						Description: "Temporary repo",
						Name:        "repo-tmp-17",
						Private:     false,
						Gitignores:  "",
						License:     "WTFPL",
						Readme:      "Default",
					})
					session.MakeRequest(t, req, http.StatusCreated)
				})

				u.Path = "user2/repo-tmp-17.git"
				u.User = url.UserPassword("user2", userPassword)
				t.Run("Clone", func(t *testing.T) {
					err = git.Clone(u.String(), dstPath, git.CloneRepoOptions{})
					assert.NoError(t, err)
					assert.True(t, com.IsExist(filepath.Join(dstPath, "README.md")))
				})

				t.Run("PushCommit", func(t *testing.T) {
					err = generateCommit(dstPath, "user2@example.com", "User Two")
					assert.NoError(t, err)
					//Push
					err = git.Push(dstPath, git.PushOptions{
						Branch: "master",
						Remote: u.String(),
						Force:  false,
					})
					assert.NoError(t, err)
				})
			})
			t.Run("LFS", func(t *testing.T) {
				t.Run("PushCommit", func(t *testing.T) {
					//Setup git LFS
					_, err = git.NewCommand("lfs").AddArguments("install").RunInDir(dstPath)
					assert.NoError(t, err)
					_, err = git.NewCommand("lfs").AddArguments("track", "data-file-*").RunInDir(dstPath)
					assert.NoError(t, err)
					err = git.AddChanges(dstPath, false, ".gitattributes")
					assert.NoError(t, err)

					err = generateCommit(dstPath, "user2@example.com", "User Two")
					//Push
					u.User = url.UserPassword("user2", userPassword)
					err = git.Push(dstPath, git.PushOptions{
						Branch: "master",
						Remote: u.String(),
						Force:  false,
					})
					assert.NoError(t, err)
				})
				t.Run("Locks", func(t *testing.T) {
					_, err = git.NewCommand("remote").AddArguments("set-url", "origin", u.String()).RunInDir(dstPath) //TODO add test ssh git-lfs-creds
					assert.NoError(t, err)
					_, err = git.NewCommand("lfs").AddArguments("locks").RunInDir(dstPath)
					assert.NoError(t, err)
					_, err = git.NewCommand("lfs").AddArguments("lock", "README.md").RunInDir(dstPath)
					assert.NoError(t, err)
					_, err = git.NewCommand("lfs").AddArguments("locks").RunInDir(dstPath)
					assert.NoError(t, err)
					_, err = git.NewCommand("lfs").AddArguments("unlock", "README.md").RunInDir(dstPath)
					assert.NoError(t, err)
				})
			})
		})
		t.Run("SSH", func(t *testing.T) {
			//Setup remote link
			u.Scheme = "ssh"
			u.User = url.User("git")
			u.Host = "localhost:22" //TODO setup port
			log.Println(u.String())

			//Setup key
			session := loginUser(t, "user1")
			keyOwner := models.AssertExistsAndLoadBean(t, &models.User{Name: "user2"}).(*models.User)
			urlStr := fmt.Sprintf("/api/v1/admin/users/%s/keys", keyOwner.Name)

			key, err := rsa.GenerateKey(rand.Reader, 2048)
			assert.NoError(t, err)
			keyPath, err := ioutil.TempFile("", "user-tmp-key")
			assert.NoError(t, err)
			saveKey(t, keyPath.Name(), key)
			defer os.Remove(keyPath.Name())
			savePubKey(t, keyPath.Name()+".pub", key.PublicKey)
			defer os.Remove(keyPath.Name())

			dataPubKey, err := ioutil.ReadFile(keyPath.Name() + ".pub")
			assert.NoError(t, err)
			fmt.Print(string(dataPubKey))
			req := NewRequestWithValues(t, "POST", urlStr, map[string]string{
				"key":   string(dataPubKey),
				"title": "test-key",
			})
			//resp :=
			session.MakeRequest(t, req, http.StatusCreated)

			//Setup clone folder
			dstPath, err := ioutil.TempDir("", "repo-tmp-18")
			assert.NoError(t, err)
			defer os.RemoveAll(dstPath)
		})
	})
}

//From https://gist.github.com/sdorra/1c95de8cb80da31610d2ad767cd6f251
func savePubKey(t *testing.T, fileName string, pubkey rsa.PublicKey) {
	asn1Bytes, err := asn1.Marshal(pubkey)
	assert.NoError(t, err)

	f, err := os.Create(fileName)
	defer f.Close()
	assert.NoError(t, err)

	f.WriteString("ssh-rsa ")
	b64 := base64.NewEncoder(base64.StdEncoding, f)
	defer b64.Close()
	_, err = b64.Write(asn1Bytes)
	assert.NoError(t, err)

}

func saveKey(t *testing.T, fileName string, key *rsa.PrivateKey) {
	outFile, err := os.Create(fileName)
	assert.NoError(t, err)
	defer outFile.Close()

	var privateKey = &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	err = pem.Encode(outFile, privateKey)
	assert.NoError(t, err)
}

/*


// user1 is an admin user
session := loginUser(t, "user1")
keyOwner := models.AssertExistsAndLoadBean(t, &models.User{Name: "user2"}).(*models.User)

urlStr := fmt.Sprintf("/api/v1/admin/users/%s/keys", keyOwner.Name)
req := NewRequestWithValues(t, "POST", urlStr, map[string]string{
	"key":   "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQDAu7tvIvX6ZHrRXuZNfkR3XLHSsuCK9Zn3X58lxBcQzuo5xZgB6vRwwm/QtJuF+zZPtY5hsQILBLmF+BZ5WpKZp1jBeSjH2G7lxet9kbcH+kIVj0tPFEoyKI9wvWqIwC4prx/WVk2wLTJjzBAhyNxfEq7C9CeiX9pQEbEqJfkKCQ== nocomment\n",
	"title": "test-key",
})
resp := session.MakeRequest(t, req, http.StatusCreated)

*/
