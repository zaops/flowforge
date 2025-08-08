package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"flowforge/pkg/config"
	"flowforge/pkg/models"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Client Git客户端
type Client struct {
	config *config.Config
}

// NewClient 创建Git客户端
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
	}
}

// CloneOptions 克隆选项
type CloneOptions struct {
	Project   *models.Project
	SSHKey    *models.SSHKey
	TargetDir string
}

// Clone 克隆代码库
func (c *Client) Clone(ctx context.Context, opts CloneOptions) error {
	// 创建目标目录
	if err := os.MkdirAll(opts.TargetDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 设置克隆选项
	cloneOpts := &git.CloneOptions{
		URL:           opts.Project.RepoURL,
		Progress:      nil,
		SingleBranch:  true,
		ReferenceName: plumbing.NewBranchReferenceName(opts.Project.Branch),
		Depth:         1,
	}

	// 设置认证
	auth, err := c.getAuth(opts.Project, opts.SSHKey)
	if err != nil {
		return fmt.Errorf("设置认证失败: %w", err)
	}
	if auth != nil {
		cloneOpts.Auth = auth
	}

	// 设置超时上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(c.config.Deploy.Timeout)*time.Second)
	defer cancel()

	// 执行克隆
	_, err = git.PlainCloneContext(timeoutCtx, opts.TargetDir, false, cloneOpts)
	if err != nil {
		return fmt.Errorf("克隆代码库失败: %w", err)
	}

	return nil
}

// PullOptions 拉取选项
type PullOptions struct {
	Project   *models.Project
	SSHKey    *models.SSHKey
	RepoDir   string
}

// Pull 拉取最新代码
func (c *Client) Pull(ctx context.Context, opts PullOptions) error {
	// 打开仓库
	repo, err := git.PlainOpen(opts.RepoDir)
	if err != nil {
		return fmt.Errorf("打开代码库失败: %w", err)
	}

	// 获取工作区
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("获取工作区失败: %w", err)
	}

	// 设置拉取选项
	pullOpts := &git.PullOptions{
		RemoteName:    "origin",
		SingleBranch:  true,
		ReferenceName: plumbing.NewBranchReferenceName(opts.Project.Branch),
	}

	// 设置认证
	auth, err := c.getAuth(opts.Project, opts.SSHKey)
	if err != nil {
		return fmt.Errorf("设置认证失败: %w", err)
	}
	if auth != nil {
		pullOpts.Auth = auth
	}

	// 设置超时上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(c.config.Deploy.Timeout)*time.Second)
	defer cancel()

	// 执行拉取
	err = worktree.PullContext(timeoutCtx, pullOpts)
	if err == git.NoErrAlreadyUpToDate {
		return nil // 已经是最新的，不是错误
	}
	if err != nil {
		return fmt.Errorf("拉取代码失败: %w", err)
	}

	return nil
}

// GetCommitInfo 获取提交信息
func (c *Client) GetCommitInfo(repoDir string) (string, string, error) {
	// 打开仓库
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return "", "", fmt.Errorf("打开代码库失败: %w", err)
	}

	// 获取HEAD引用
	ref, err := repo.Head()
	if err != nil {
		return "", "", fmt.Errorf("获取HEAD引用失败: %w", err)
	}

	// 获取提交对象
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return "", "", fmt.Errorf("获取提交对象失败: %w", err)
	}

	// 获取分支名称
	branch := ""
	refs, err := repo.References()
	if err == nil {
		refs.ForEach(func(ref *plumbing.Reference) error {
			if ref.Hash() == commit.Hash && ref.Name().IsBranch() {
				branch = ref.Name().Short()
				return nil
			}
			return nil
		})
	}

	return commit.Hash.String(), branch, nil
}

// getAuth 获取认证信息
func (c *Client) getAuth(project *models.Project, sshKey *models.SSHKey) (transport.AuthMethod, error) {
	// 如果使用SSH密钥
	if project.SSHKeyID != nil && sshKey != nil {
		// 创建临时SSH密钥文件
		keyFile := filepath.Join(c.config.SSH.KeysPath, fmt.Sprintf("key_%d", *project.SSHKeyID))
		if err := os.WriteFile(keyFile, []byte(sshKey.PrivateKey), 0600); err != nil {
			return nil, fmt.Errorf("写入SSH密钥文件失败: %w", err)
		}
		defer os.Remove(keyFile) // 使用后删除

		// 创建SSH认证
		publicKeys, err := ssh.NewPublicKeysFromFile("git", keyFile, "")
		if err != nil {
			return nil, fmt.Errorf("创建SSH公钥失败: %w", err)
		}

		return publicKeys, nil
	}

	// 无认证
	return nil, nil
}