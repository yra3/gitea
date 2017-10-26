// Copyright 2016 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routers

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"code.gitea.io/git"
	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/models/migrations"
	"code.gitea.io/gitea/modules/cache"
	"code.gitea.io/gitea/modules/cron"
	"code.gitea.io/gitea/modules/highlight"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/mailer"
	"code.gitea.io/gitea/modules/markup"
	"code.gitea.io/gitea/modules/plugins"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/ssh"

	macaron "gopkg.in/macaron.v1"
)

func checkRunMode() {
	switch setting.Cfg.Section("").Key("RUN_MODE").String() {
	case "prod":
		macaron.Env = macaron.PROD
		macaron.ColorLog = false
		setting.ProdMode = true
	default:
		git.Debug = true
	}
	log.Info("Run Mode: %s", strings.Title(macaron.Env))
}

// NewServices init new services
func NewServices() {
	setting.NewServices()
	mailer.NewContext()
	cache.NewContext()
}

// GlobalInit is for global configuration reload-able.
func GlobalInit() {
	setting.NewContext()
	log.Trace("Custom path: %s", setting.CustomPath)
	log.Trace("Log path: %s", setting.LogRootPath)
	models.LoadConfigs()
	NewServices()

	if setting.InstallLock {
		highlight.NewContext()
		markup.Init()

		if err := models.NewEngine(migrations.Migrate); err != nil {
			log.Fatal(4, "Failed to initialize ORM engine: %v", err)
		}
		models.HasEngine = true
		models.InitOAuth2()

		models.LoadRepoConfig()
		models.NewRepoContext()

		// Booting long running goroutines.
		cron.NewContext()
		models.InitIssueIndexer()
		models.InitSyncMirrors()
		models.InitDeliverHooks()
		models.InitTestPullRequests()
		log.NewGitLogger(path.Join(setting.LogRootPath, "http.log"))
	}
	if models.EnableSQLite3 {
		log.Info("SQLite3 Supported")
	}
	if models.EnableTiDB {
		log.Info("TiDB Supported")
	}
	checkRunMode()

	if setting.InstallLock && setting.SSH.StartBuiltinServer {
		ssh.Listen(setting.SSH.ListenHost, setting.SSH.ListenPort, setting.SSH.ServerCiphers)
		log.Info("SSH server started on %s:%d. Cipher list (%v)", setting.SSH.ListenHost, setting.SSH.ListenPort, setting.SSH.ServerCiphers)
	}

	// TODO settings if setting.StartPlugins {
	//TODO define in settings
	pluginPath := "./plugins"
	pManager := plugins.GetManager()

	//Create folder if not exist
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		os.Mkdir(pluginPath, 0755)
	}

	err := filepath.Walk(pluginPath, func(path string, f os.FileInfo, err error) error {
		//log.Debug("Potential plugin found : %v %v %v", path, f, err)
		if f != nil && !f.IsDir() && f.Mode() == 0755 {
			log.Debug("Potential plugin found : %s", path)
			pID := pManager.Add(path)
			//TODO testing and start via db config / via admin panel
			pManager.Start(pID)
		}
		return nil
	})
	if err != nil {
		log.Fatal(7, "Failed to walk plugin folder '%s' : %v", pluginPath, err)
	}
	//}
}
