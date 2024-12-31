package engine

import (
	"goss/auth"
	"goss/mod/download"
	"goss/mod/execute"
	"goss/mod/upload"
)

func Download(o *auth.Opts, remotePath, localPath string) error {
	ssh, err := auth.Init(o)
	if err != nil {
		return err
	} else {
		defer ssh.Close()
	}
	err = download.Download(remotePath, localPath, ssh)
	if err != nil {
		return err
	}
	return nil
}

func Upload(o *auth.Opts, remotePath, localPath string) error {
	ssh, err := auth.Init(o)
	if err != nil {
		return err
	} else {
		defer ssh.Close()
	}
	err = upload.Upload(remotePath, localPath, ssh)
	if err != nil {
		return err
	}
	return nil
}

func Execute(o *auth.Opts, pass, cmd string) (string, error) {
	ssh, err := auth.Init(o)
	if err != nil {
		return "", err
	} else {
		defer ssh.Close()
	}
	if pass != "" {
		return execute.SudoExecute(pass, cmd, ssh)
	} else {
		return execute.Execute(cmd, ssh)
	}
}
