package download

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func Download(remotePath, localPath string, ssh *ssh.Client) error {
	sftpClient, err := sftp.NewClient(ssh)
	if err != nil {
		return err
	}
	defer func() {
		if sftpClient != nil {
			sftpClient.Close()
		}
	}()
	info, err := sftpClient.Lstat(remotePath)
	if err != nil {
		return err
	}
	sfs, err := sftpClient.Lstat(remotePath)
	if err != nil {
		return err
	}
	if !sfs.IsDir() {
		localPath = filepath.Join(localPath, filepath.Base(remotePath))
	}
	// 判断是文件还是目录
	if info.IsDir() {
		err := downloadDirectory(remotePath, localPath, sftpClient)
		if err != nil {
			return err
		}
	} else {
		err := downloadFile(remotePath, localPath, sftpClient)
		if err != nil {
			return err
		}
	}
	return nil
}

// 下载目录
func downloadDirectory(remoteDir, localDir string, c *sftp.Client) error {
	w := c.Walk(remoteDir)
	for w.Step() {
		if w.Err() != nil {
			continue
		}
		remotePath := w.Path()
		relPath, err := filepath.Rel(remoteDir, remotePath)
		if err != nil {
			return err
		}
		localPath := filepath.Join(localDir, relPath)
		if w.Stat().IsDir() {
			err = os.MkdirAll(localPath, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			err = downloadFile(remotePath, localPath, c)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// 下载文件
func downloadFile(remoteFile, localFile string, c *sftp.Client) error {
	srcFile, err := c.Open(remoteFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	localFs, err := os.Create(localFile)
	if err != nil {
		return err
	}
	defer localFs.Close()
	_, err = io.Copy(localFs, srcFile)
	return err
}
