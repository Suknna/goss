package easysftp

import (
	"context"
	"io"
	"os"

	"github.com/kr/fs"
	"github.com/pkg/sftp"
)

type FileSystem interface {
	Stat(path string) (os.FileInfo, error)
	MkdirAll(path string, perm os.FileMode) error
	OpenReader(path string) (io.ReadCloser, error)
	OpenWriter(path string, flag int, perm os.FileMode) (io.WriteCloser, error)
	Walk(path string) (*DirOperation, error)
}

// 目录操作
type DirOperation struct {
	Walker *fs.Walker
	Client *sftp.Client
}

// SFTP接口
type SFTPFileSystem struct {
	client *sftp.Client
	ctx    context.Context
}

// 本地接口
type OSFileSystem struct {
	ctx context.Context
}

func CreateSftpFS(ctx context.Context, sftpClient *sftp.Client) FileSystem {
	// 创建sftp client
	return &SFTPFileSystem{
		client: sftpClient,
		ctx:    ctx,
	}
}

func CreateLocalFS(ctx context.Context) FileSystem {
	return &OSFileSystem{
		ctx: ctx,
	}
}

// SFTP只读文件接口
func (s *SFTPFileSystem) OpenReader(path string) (io.ReadCloser, error) {
	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		return s.client.Open(path)
	}
}

// OS只读文件接口
func (o *OSFileSystem) OpenReader(path string) (io.ReadCloser, error) {
	select {
	case <-o.ctx.Done():
		return nil, o.ctx.Err()
	default:
		return os.Open(path)
	}
}

// SFTP只写文件接口
func (s *SFTPFileSystem) OpenWriter(path string, flag int, perm os.FileMode) (io.WriteCloser, error) {
	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		return s.client.OpenFile(path, flag)
	}
}

// OS只写接口
func (o *OSFileSystem) OpenWriter(path string, flag int, perm os.FileMode) (io.WriteCloser, error) {
	select {
	case <-o.ctx.Done():
		return nil, o.ctx.Err()
	default:
		return os.OpenFile(path, flag, perm)
	}
}

// SFTP获取状态接口
func (s *SFTPFileSystem) Stat(path string) (os.FileInfo, error) {
	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		return s.client.Stat(path)
	}
}

// OS获取状态接口
func (o *OSFileSystem) Stat(path string) (os.FileInfo, error) {
	select {
	case <-o.ctx.Done():
		return nil, o.ctx.Err()
	default:
		return os.Stat(path)
	}
}

// SFTP创建目录接口
func (s *SFTPFileSystem) MkdirAll(path string, perm os.FileMode) error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
		return s.client.MkdirAll(path)
	}
}

// OS创建目录接口
func (o *OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	select {
	case <-o.ctx.Done():
		return o.ctx.Err()
	default:
		return os.MkdirAll(path, perm)
	}
}

func (s *SFTPFileSystem) Walk(path string) (*DirOperation, error) {
	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		return &DirOperation{
			Walker: s.client.Walk(path),
			Client: s.client,
		}, nil
	}

}
func (o *OSFileSystem) Walk(path string) (*DirOperation, error) {
	select {
	case <-o.ctx.Done():
		return nil, o.ctx.Err()
	default:
		return &DirOperation{
			Walker: fs.Walk(path),
			Client: nil,
		}, nil
	}
}
