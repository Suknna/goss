package upload

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func Upload(remotePath, localPath string, ssh *ssh.Client) error {
	// 业务处理逻辑
	sftpClient, err := sftp.NewClient(ssh)
	if err != nil {
		return err
	}
	defer func() {
		if sftpClient != nil {
			sftpClient.Close()
		}
	}()
	f, err := os.Stat(localPath)
	if err != nil {
		return err
	}
	remoteFileStat, err := sftpClient.Lstat(remotePath)
	if err != nil {
		return err
	}
	if remoteFileStat.IsDir() {
		remotePath = filepath.Join(remotePath, filepath.Base(localPath))
	}
	if f.IsDir() {
		if err := uploadDirectory(localPath, remotePath, sftpClient); err != nil {
			return err
		}
	} else {
		if err := uploadFile(localPath, remotePath, sftpClient); err != nil {
			return err
		}
	}
	return nil
}

func uploadDirectory(localDir, remoteDir string, c *sftp.Client) error {
	err := filepath.Walk(localDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}
		remotePath := filepath.Join(remoteDir, relPath)
		remotePath = filepath.Clean(remotePath)
		if info.IsDir() {
			if err := c.MkdirAll(remotePath); err != nil {
				return err
			}
		} else {
			if err := uploadFile(path, remotePath, c); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func uploadFile(localFile, remoteFile string, c *sftp.Client) error {
	srcFile, err := os.Open(localFile)
	if err != nil {
		log.Println(localFile)
		return err
	}
	defer srcFile.Close()
	dstFile, err := c.Create(remoteFile)
	if err != nil {
		log.Println(remoteFile)
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	return err
}
