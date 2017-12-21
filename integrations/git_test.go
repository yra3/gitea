// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"context"
	"crypto/rand"
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
	"code.gitea.io/gitea/modules/ssh"
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
	go ssh.Listen(setting.SSH.ListenHost, setting.SSH.ListenPort, setting.SSH.ServerCiphers, setting.SSH.ServerKeyExchanges, setting.SSH.ServerMACs)

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
			u.Host = fmt.Sprintf("%s:%d", setting.SSH.ListenHost, setting.SSH.ListenPort)
			u.Path = "user2/repo-tmp-18.git"
			log.Println(u.String()) //TODO remove debug

			//Setup key
			keyFolder, err := ioutil.TempDir("", "tmp-key-folder")
			assert.NoError(t, err)
			defer os.RemoveAll(keyFolder)
			_, _, err = com.ExecCmd("ssh-keygen", "-f", filepath.Join(keyFolder, "my-testing-key"), "-t", "rsa", "-N", "")
			assert.NoError(t, err)

			session := loginUser(t, "user1")
			keyOwner := models.AssertExistsAndLoadBean(t, &models.User{Name: "user2"}).(*models.User)
			urlStr := fmt.Sprintf("/api/v1/admin/users/%s/keys", keyOwner.Name)

			dataPubKey, err := ioutil.ReadFile(filepath.Join(keyFolder, "my-testing-key.pub"))
			assert.NoError(t, err)
			fmt.Print(string(dataPubKey)) //TODO remove debug
			req := NewRequestWithValues(t, "POST", urlStr, map[string]string{
				"key":   string(dataPubKey),
				"title": "test-key",
			})
			session.MakeRequest(t, req, http.StatusCreated)

			//Setup ssh wrapper
			sshWrapper, err := ioutil.TempFile("./integrations/", "tmp-ssh-wrapper")
			sshWrapper.WriteString("#!/bin/sh\n\n")
			sshWrapper.WriteString("ssh -i \"" + filepath.Join(keyFolder, "my-testing-key") + "\" $* \n\n")
			err = sshWrapper.Chmod(os.ModePerm)
			assert.NoError(t, err)
			sshWrapper.Close()

			//Setup clone folder
			dstPath, err := ioutil.TempDir("", "repo-tmp-18")
			assert.NoError(t, err)
			defer os.RemoveAll(dstPath)

			t.Run("Standard", func(t *testing.T) {
				t.Run("CreateRepo", func(t *testing.T) {
					session := loginUser(t, "user2")
					req := NewRequestWithJSON(t, "POST", "/api/v1/user/repos", &api.CreateRepoOption{
						AutoInit:    true,
						Description: "Temporary repo",
						Name:        "repo-tmp-18",
						Private:     false,
						Gitignores:  "",
						License:     "WTFPL",
						Readme:      "Default",
					})
					session.MakeRequest(t, req, http.StatusCreated)
				})

				t.Run("Clone", func(t *testing.T) {
					err = git.Clone(u.String(), dstPath, git.CloneRepoOptions{})
					_, err = git.NewCommand("clone").AddArguments("--config", "core.sshCommand=./"+sshWrapper.Name(), u.String(), dstPath).Run() //TODO improve wrapper and add to make clean
					assert.NoError(t, err)
					assert.True(t, com.IsExist(filepath.Join(dstPath, "README.md")))
				})
				//time.Sleep(5 * time.Minute)
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

		})
	})
}
