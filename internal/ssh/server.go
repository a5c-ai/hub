package ssh

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

// SSHServer represents the SSH server for git operations
type SSHServer struct {
	config            *ssh.ServerConfig
	listener          net.Listener
	port              int
	hostKeyPath       string
	repositoryService RepositoryService
	gitService        GitShellService
	logger            *logrus.Logger
	db                *gorm.DB
}

// GitShellService defines git shell operations
type GitShellService interface {
	HandleGitCommand(ctx context.Context, command string, repoPath string, stdin io.Reader, stdout, stderr io.Writer) error
}

// SSHServerConfig holds SSH server configuration
type SSHServerConfig struct {
	Port        int    `mapstructure:"port"`
	HostKeyPath string `mapstructure:"host_key_path"`
}

// NewSSHServer creates a new SSH server instance
func NewSSHServer(
	config SSHServerConfig,
	repositoryService RepositoryService,
	gitService GitShellService,
	logger *logrus.Logger,
	db *gorm.DB,
) (*SSHServer, error) {
	server := &SSHServer{
		port:              config.Port,
		hostKeyPath:       config.HostKeyPath,
		repositoryService: repositoryService,
		gitService:        gitService,
		logger:            logger,
		db:                db,
	}

	// Initialize SSH server config
	if err := server.initializeConfig(); err != nil {
		return nil, fmt.Errorf("failed to initialize SSH config: %w", err)
	}

	return server, nil
}

// initializeConfig sets up the SSH server configuration
func (s *SSHServer) initializeConfig() error {
	config := &ssh.ServerConfig{
		PublicKeyCallback: s.authenticatePublicKey,
		ServerVersion:     "SSH-2.0-Hub-Git-Server",
	}

	// Load or generate host key
	hostKey, err := s.loadOrGenerateHostKey()
	if err != nil {
		return fmt.Errorf("failed to load host key: %w", err)
	}

	config.AddHostKey(hostKey)
	s.config = config

	return nil
}

// loadOrGenerateHostKey loads existing host key or generates a new one
func (s *SSHServer) loadOrGenerateHostKey() (ssh.Signer, error) {
	if s.hostKeyPath == "" {
		s.hostKeyPath = "./ssh_host_key"
	}

	// Try to load existing key
	if _, err := os.Stat(s.hostKeyPath); err == nil {
		keyData, err := os.ReadFile(s.hostKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read host key file: %w", err)
		}

		key, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			s.logger.Warn("Failed to parse existing host key, generating new one")
		} else {
			s.logger.Info("Loaded existing SSH host key")
			return key, nil
		}
	}

	// Generate new key
	s.logger.Info("Generating new SSH host key")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Convert to PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Save to file
	keyFile, err := os.Create(s.hostKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create host key file: %w", err)
	}
	defer keyFile.Close()

	if err := pem.Encode(keyFile, privateKeyPEM); err != nil {
		return nil, fmt.Errorf("failed to encode private key: %w", err)
	}

	// Set restrictive permissions
	if err := os.Chmod(s.hostKeyPath, 0600); err != nil {
		s.logger.WithError(err).Warn("Failed to set host key file permissions")
	}

	// Convert to SSH key
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer from private key: %w", err)
	}

	s.logger.Info("Generated and saved new SSH host key")
	return signer, nil
}

// authenticatePublicKey authenticates users by their SSH public keys
func (s *SSHServer) authenticatePublicKey(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	username := conn.User()
	keyType := key.Type()

	s.logger.WithFields(logrus.Fields{
		"username": username,
		"key_type": keyType,
		"remote":   conn.RemoteAddr(),
	}).Debug("SSH authentication attempt")

	// Look up user by username
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		s.logger.WithError(err).WithField("username", username).Debug("User not found")
		return nil, fmt.Errorf("authentication failed")
	}

	// Get user's SSH keys
	var sshKeys []models.SSHKey
	if err := s.db.Where("user_id = ? AND active = ?", user.ID, true).Find(&sshKeys).Error; err != nil {
		s.logger.WithError(err).WithField("user_id", user.ID).Error("Failed to fetch SSH keys")
		return nil, fmt.Errorf("authentication failed")
	}

	// Check if any key matches
	for _, sshKey := range sshKeys {
		publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(sshKey.KeyData))
		if err != nil {
			s.logger.WithError(err).WithField("key_id", sshKey.ID).Debug("Failed to parse stored public key")
			continue
		}

		// Compare keys by comparing their marshaled bytes
		if bytes.Equal(key.Marshal(), publicKey.Marshal()) {
			s.logger.WithFields(logrus.Fields{
				"username": username,
				"key_name": sshKey.Title,
				"key_id":   sshKey.ID,
			}).Info("SSH authentication successful")

			// Update last used time
			s.db.Model(&sshKey).Update("last_used_at", time.Now())

			return &ssh.Permissions{
				Extensions: map[string]string{
					"user_id":  user.ID.String(),
					"username": user.Username,
					"key_id":   sshKey.ID.String(),
				},
			}, nil
		}
	}

	s.logger.WithField("username", username).Debug("No matching SSH key found")
	return nil, fmt.Errorf("authentication failed")
}

// Start starts the SSH server
func (s *SSHServer) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}

	s.listener = listener
	s.logger.WithField("port", s.port).Info("SSH server started")

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return nil
				default:
					s.logger.WithError(err).Error("Failed to accept SSH connection")
					continue
				}
			}

			go s.handleConnection(ctx, conn)
		}
	}
}

// Stop stops the SSH server
func (s *SSHServer) Stop() error {
	if s.listener != nil {
		s.logger.Info("Stopping SSH server")
		return s.listener.Close()
	}
	return nil
}

// handleConnection handles an SSH connection
func (s *SSHServer) handleConnection(ctx context.Context, netConn net.Conn) {
	defer netConn.Close()

	// Perform SSH handshake
	conn, chans, reqs, err := ssh.NewServerConn(netConn, s.config)
	if err != nil {
		s.logger.WithError(err).Debug("Failed to perform SSH handshake")
		return
	}
	defer conn.Close()

	username := conn.Permissions.Extensions["username"]
	userID := conn.Permissions.Extensions["user_id"]

	s.logger.WithFields(logrus.Fields{
		"username": username,
		"user_id":  userID,
		"remote":   conn.RemoteAddr(),
	}).Info("SSH connection established")

	// Handle out-of-band requests
	go ssh.DiscardRequests(reqs)

	// Handle channels
	for newChannel := range chans {
		go s.handleChannel(ctx, newChannel, conn.Permissions)
	}
}

// handleChannel handles an SSH channel
func (s *SSHServer) handleChannel(ctx context.Context, newChannel ssh.NewChannel, perms *ssh.Permissions) {
	if newChannel.ChannelType() != "session" {
		newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
		return
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		s.logger.WithError(err).Error("Failed to accept SSH channel")
		return
	}
	defer channel.Close()

	// Handle channel requests
	go func() {
		for req := range requests {
			switch req.Type {
			case "exec":
				s.handleExec(ctx, req, channel, perms)
			default:
				if req.WantReply {
					req.Reply(false, nil)
				}
			}
		}
	}()
}

// handleExec handles SSH exec requests (git commands)
func (s *SSHServer) handleExec(ctx context.Context, req *ssh.Request, channel ssh.Channel, perms *ssh.Permissions) {
	if !req.WantReply {
		return
	}

	// Parse command
	command := string(req.Payload[4:]) // Skip length prefix
	s.logger.WithFields(logrus.Fields{
		"username": perms.Extensions["username"],
		"command":  command,
	}).Info("SSH exec request")

	// Only allow git commands
	if !s.isValidGitCommand(command) {
		s.logger.WithField("command", command).Warn("Invalid git command")
		req.Reply(false, nil)
		return
	}

	req.Reply(true, nil)

	// Execute git command
	if err := s.executeGitCommand(ctx, command, channel, perms); err != nil {
		s.logger.WithError(err).Error("Failed to execute git command")
		channel.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{Status: 1}))
	} else {
		channel.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{Status: 0}))
	}
}

// isValidGitCommand checks if the command is a valid git command
func (s *SSHServer) isValidGitCommand(command string) bool {
	parts := strings.Fields(command)
	if len(parts) < 2 {
		return false
	}

	gitCommand := parts[0]
	return gitCommand == "git-upload-pack" || gitCommand == "git-receive-pack"
}

// executeGitCommand executes a git command
func (s *SSHServer) executeGitCommand(ctx context.Context, command string, channel ssh.Channel, perms *ssh.Permissions) error {
	parts := strings.Fields(command)
	if len(parts) < 2 {
		return fmt.Errorf("invalid command format")
	}

	gitCommand := parts[0]
	repoPath := strings.Trim(parts[1], "'\"")

	// Remove leading slash and .git suffix
	repoPath = strings.TrimPrefix(repoPath, "/")
	repoPath = strings.TrimSuffix(repoPath, ".git")

	// Parse owner/repo from path
	pathParts := strings.Split(repoPath, "/")
	if len(pathParts) != 2 {
		return fmt.Errorf("invalid repository path format: %s", repoPath)
	}

	owner := pathParts[0]
	repoName := pathParts[1]

	// Get repository
	repo, err := s.repositoryService.Get(ctx, owner, repoName)
	if err != nil {
		return fmt.Errorf("repository not found: %s/%s", owner, repoName)
	}

	// Get repository filesystem path
	actualRepoPath, err := s.repositoryService.GetRepositoryPath(ctx, repo.ID)
	if err != nil {
		return fmt.Errorf("failed to get repository path: %w", err)
	}

	// Execute git command
	return s.gitService.HandleGitCommand(ctx, gitCommand, actualRepoPath, channel, channel, channel)
}
