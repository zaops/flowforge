package ssh

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"flowforge/pkg/config"
	"flowforge/pkg/models"
	"golang.org/x/crypto/ssh"
)

// Client SSH客户端
type Client struct {
	config *config.Config
}

// Manager SSH管理器
type Manager struct {
	client *Client
	config *config.Config
}

// Manager SSH管理器
type Manager struct {
	client *Client
	config *config.Config
}

// NewClient 创建SSH客户端
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
	}
}

// NewManager 创建SSH管理器
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		client: NewClient(cfg),
		config: cfg,
	}
}

// GetClient 获取SSH客户端
func (m *Manager) GetClient() *Client {
	return m.client
}

package ssh

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"flowforge/pkg/config"
	"flowforge/pkg/models"
	"golang.org/x/crypto/ssh"
)


// GenerateKeyPair 生成SSH密钥对
func (c *Client) GenerateKeyPair(bits int, passphrase string) (string, string, error) {
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", fmt.Errorf("生成RSA密钥对失败: %w", err)
	}

	// 将私钥转换为PEM格式
	var privateKeyPEM bytes.Buffer
	if passphrase == "" {
		// 不加密
		privateKeyBlock := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}
		if err := pem.Encode(&privateKeyPEM, privateKeyBlock); err != nil {
			return "", "", fmt.Errorf("编码私钥失败: %w", err)
		}
	} else {
		// 使用密码加密
		privateKeyBlock, err := x509.EncryptPEMBlock(
			rand.Reader,
			"RSA PRIVATE KEY",
			x509.MarshalPKCS1PrivateKey(privateKey),
			[]byte(passphrase),
			x509.PEMCipherAES256,
		)
		if err != nil {
			return "", "", fmt.Errorf("加密私钥失败: %w", err)
		}
		if err := pem.Encode(&privateKeyPEM, privateKeyBlock); err != nil {
			return "", "", fmt.Errorf("编码私钥失败: %w", err)
		}
	}

	// 生成公钥
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("生成公钥失败: %w", err)
	}

	// 将公钥转换为授权密钥格式
	publicKeyString := string(ssh.MarshalAuthorizedKey(publicKey))

	return privateKeyPEM.String(), publicKeyString, nil
}

// TestConnection 测试SSH连接
func (c *Client) TestConnection(host string, port int, username string, privateKey string, passphrase string) error {
	// 解析私钥
	var signer ssh.Signer
	var err error
	if passphrase == "" {
		signer, err = ssh.ParsePrivateKey([]byte(privateKey))
	} else {
		signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(privateKey), []byte(passphrase))
	}
	if err != nil {
		return fmt.Errorf("解析私钥失败: %w", err)
	}

	// 创建SSH客户端配置
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 仅用于测试，生产环境应使用已知主机密钥
		Timeout:         time.Duration(c.config.SSH.Timeout) * time.Second,
	}

	// 连接到SSH服务器
	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("SSH连接失败: %w", err)
	}
	defer client.Close()

	// 创建会话
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer session.Close()

	// 执行简单命令
	var stdout bytes.Buffer
	session.Stdout = &stdout
	if err := session.Run("echo Connection successful"); err != nil {
		return fmt.Errorf("执行命令失败: %w", err)
	}

	return nil
}

// ExecuteCommand 执行SSH命令
func (c *Client) ExecuteCommand(sshKey *models.SSHKey, host string, port int, username string, command string) (string, error) {
	// 创建临时SSH密钥文件
	keyFile := filepath.Join(c.config.SSH.KeysPath, fmt.Sprintf("key_%d", sshKey.ID))
	if err := os.WriteFile(keyFile, []byte(sshKey.PrivateKey), 0600); err != nil {
		return "", fmt.Errorf("写入SSH密钥文件失败: %w", err)
	}
	defer os.Remove(keyFile) // 使用后删除

	// 解析私钥
	var signer ssh.Signer
	var err error
	// SSHKey模型中没有密码字段，假设私钥没有密码保护
	signer, err = ssh.ParsePrivateKey([]byte(sshKey.PrivateKey))
	if err != nil {
		return "", fmt.Errorf("解析私钥失败: %w", err)
	}

	// 创建SSH客户端配置
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 仅用于测试，生产环境应使用已知主机密钥
		Timeout:         time.Duration(c.config.SSH.Timeout) * time.Second,
	}

	// 连接到SSH服务器
	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return "", fmt.Errorf("SSH连接失败: %w", err)
	}
	defer client.Close()

	// 创建会话
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer session.Close()

	// 执行命令
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	if err := session.Run(command); err != nil {
		return stderr.String(), fmt.Errorf("执行命令失败: %w", err)
	}

	return stdout.String(), nil
}

// CopyFile 通过SCP复制文件
func (c *Client) CopyFile(sshKey *models.SSHKey, host string, port int, username string, localPath string, remotePath string) error {
	// 解析私钥
	var signer ssh.Signer
	var err error
	// SSHKey模型中没有密码字段，假设私钥没有密码保护
	signer, err = ssh.ParsePrivateKey([]byte(sshKey.PrivateKey))
	if err != nil {
		return fmt.Errorf("解析私钥失败: %w", err)
	}

	// 创建SSH客户端配置
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 仅用于测试，生产环境应使用已知主机密钥
		Timeout:         time.Duration(c.config.SSH.Timeout) * time.Second,
	}

	// 连接到SSH服务器
	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("SSH连接失败: %w", err)
	}
	defer client.Close()

	// 创建会话
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer session.Close()

	// 打开本地文件
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %w", err)
	}
	defer localFile.Close()

	// 获取文件信息
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 设置管道
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()

		// 发送文件头
		fmt.Fprintf(w, "C%#o %d %s\n", fileInfo.Mode().Perm(), fileInfo.Size(), filepath.Base(remotePath))

		// 发送文件内容
		io.Copy(w, localFile)

		// 发送结束标记
		fmt.Fprint(w, "\x00")
	}()

	// 执行SCP命令
	if err := session.Run(fmt.Sprintf("scp -t %s", filepath.Dir(remotePath))); err != nil {
		return fmt.Errorf("执行SCP命令失败: %w", err)
	}

	return nil
}