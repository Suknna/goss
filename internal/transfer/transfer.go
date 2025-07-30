package transfer

import (
	"context"
	"fmt"
	"goss/internal/config"
	"goss/pkg/easysftp"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Transfer struct {
	host       string
	client     *sftp.Client
	remotePath string
	localPath  string
	flag       int
	auto       bool
}
type result struct {
	written int64
	err     error
}

func NewTransferHandler(remotePath, localPath string, policy config.FileTransferPolicy, conn *ssh.Client) (*Transfer, error) {
	var (
		flag int
		auto bool
	)
	switch policy {
	case config.Always:
		flag = os.O_TRUNC | os.O_WRONLY | os.O_CREATE
	case config.Never:
		flag = os.O_CREATE | os.O_WRONLY | os.O_EXCL
	}
	// 创建sftp连接
	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		return nil, err
	}
	return &Transfer{
		host:       strings.Split(conn.Conn.RemoteAddr().String(), ":")[0],
		client:     sftpClient,
		remotePath: remotePath,
		localPath:  localPath,
		flag:       flag,
		auto:       auto,
	}, nil
}

func (t *Transfer) Download(ctx context.Context) (string, error) {
	// 收集
	ch := make(chan result)
	// 创建src和dst
	src := easysftp.CreateSftpFS(ctx, t.client)
	dst := easysftp.CreateLocalFS(ctx)
	// 获取src文件的信息
	srcInfo, err := src.Stat(t.remotePath)
	if err != nil {
		return "", fmt.Errorf("")
	}
	go func() {
		defer close(ch)
		if srcInfo.IsDir() {
			b, err := t.dirHandle(t.remotePath, t.localPath, src, dst, srcInfo)
			ch <- result{
				written: b,
				err:     err,
			}

		} else {
			b, err := t.fileHandle(t.remotePath, t.localPath, src, dst, srcInfo)
			ch <- result{
				written: b,
				err:     err,
			}
		}
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case r := <-ch:
		return fmt.Sprint(humanize.Bytes(uint64(r.written))), r.err
	}
}

func (t *Transfer) Upload(ctx context.Context) (string, error) {
	ch := make(chan result)
	// 创建src和dst
	src := easysftp.CreateLocalFS(ctx)
	dst := easysftp.CreateSftpFS(ctx, t.client)
	// 获取src文件的信息
	srcInfo, err := src.Stat(t.localPath)
	if err != nil {
		return "", err
	}
	go func() {
		if srcInfo.IsDir() {
			b, err := t.dirHandle(t.localPath, t.remotePath, src, dst, srcInfo)
			ch <- result{
				written: b,
				err:     err,
			}
		} else {
			b, err := t.fileHandle(t.localPath, t.remotePath, src, dst, srcInfo)
			ch <- result{
				written: b,
				err:     err,
			}
		}
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case r := <-ch:
		return fmt.Sprint(humanize.Bytes(uint64(r.written))), r.err
	}
}

func (t *Transfer) dirHandle(srcPath, dstPath string, src, dst easysftp.FileSystem, fileInfo os.FileInfo) (int64, error) {
	var written int64
	// 以源端为主决定目标端文件
	if err := dst.MkdirAll(t.localPath, fileInfo.Mode()); err != nil {
		return 0, err
	}
	// 递归查询文件
	srcDirOperation, err := src.Walk(srcPath)
	if err != nil {
		return 0, err
	}
	w := srcDirOperation.Walker
	for w.Step() {
		if w.Err() != nil {
			return 0, err
		}
		currentSrcPath := filepath.Clean(w.Path())
		currentDstPath := filepath.Join(dstPath, filepath.Base(currentSrcPath))
		if w.Stat().IsDir() {
			// 创建目录
			if err := dst.MkdirAll(currentDstPath, w.Stat().Mode()); err != nil {
				return 0, err
			}
		} else {
			b, err := t.fileHandle(currentSrcPath, currentDstPath, src, dst, fileInfo)
			if err != nil {
				return 0, err
			}
			written += b
		}
	}
	return written, nil
}

func (t *Transfer) fileHandle(srcPath, dstPath string, src, dst easysftp.FileSystem, fileInfo os.FileInfo) (int64, error) {
	// 确保父目录存在
	if err := dst.MkdirAll(filepath.Dir(dstPath), fileInfo.Mode()); err != nil {
		return 0, err
	}
	// 创建reader和writer
	reader, err := src.OpenReader(srcPath)
	if err != nil {
		return 0, err
	}
	defer reader.Close()
	writer, err := dst.OpenWriter(dstPath, t.flag, fileInfo.Mode())
	if err != nil {
		return 0, err
	}
	defer writer.Close()

	// 使用带进度条的拷贝
	return io.Copy(writer, reader)
}

func GenerateTimestampedFilename(host string, p string) string {
	suffix := fmt.Sprintf("%s_%s", host, time.Now().Format("2006-01-02T150405Z0700"))
	dir, file := filepath.Split(p)
	fileSlice := strings.Split(file, ".")
	file = fileSlice[0] + suffix + fileSlice[1]
	return path.Join(dir, file)
}
