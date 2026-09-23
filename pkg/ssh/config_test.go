package ssh

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SSHConfigTestSuite struct {
	suite.Suite
}

func TestSSHConfigSuite(t *testing.T) {
	suite.Run(t, new(SSHConfigTestSuite))
}

func (s *SSHConfigTestSuite) TestAddHostSection() {
	tests := []struct {
		name                   string
		config                 string
		execPath               string
		host                   string
		user                   string
		context                string
		workspace              string
		workdir                string
		command                string
		gpgagent               bool
		devPodHome             string
		provider               string
		expected               string
		// skipConfigContainsCheck disables the blanket assert.Contains(result,
		// config) check for cases where the new block is inserted inside the
		// existing config, splitting it.
		skipConfigContainsCheck bool
	}{
		{
			name:       "Basic host addition",
			config:     "",
			execPath:   "/path/to/exec",
			host:       "testhost",
			user:       "testuser",
			context:    "testcontext",
			workspace:  "testworkspace",
			workdir:    "",
			command:    "",
			gpgagent:   false,
			devPodHome: "",
			provider:   "",
			expected: `# DevPod Start testhost
Host testhost
  ForwardAgent yes
  LogLevel error
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null
  HostKeyAlgorithms rsa-sha2-256,rsa-sha2-512,ssh-rsa
  ProxyCommand "/path/to/exec" ssh --stdio --context testcontext --user testuser testworkspace
  User testuser
# DevPod End testhost`,
		},
		{
			name:       "AWS provider with ConnectTimeout",
			config:     "",
			execPath:   "/path/to/exec",
			host:       "testhost",
			user:       "testuser",
			context:    "testcontext",
			workspace:  "testworkspace",
			workdir:    "",
			command:    "",
			gpgagent:   false,
			devPodHome: "",
			provider:   "aws",
			expected: `# DevPod Start testhost
Host testhost
  ForwardAgent yes
  LogLevel error
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null
  HostKeyAlgorithms rsa-sha2-256,rsa-sha2-512,ssh-rsa
  ConnectTimeout 60
  ProxyCommand "/path/to/exec" ssh --stdio --context testcontext --user testuser testworkspace
  User testuser
# DevPod End testhost`,
		},
		{
			name:       "Basic host addition with DEVPOD_HOME",
			config:     "",
			execPath:   "/path/to/exec",
			host:       "testhost",
			user:       "testuser",
			context:    "testcontext",
			workspace:  "testworkspace",
			workdir:    "",
			command:    "",
			gpgagent:   false,
			devPodHome: "C:\\\\White Space\\devpod\\test",
			provider:   "",
			expected: `# DevPod Start testhost
Host testhost
  ForwardAgent yes
  LogLevel error
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null
  HostKeyAlgorithms rsa-sha2-256,rsa-sha2-512,ssh-rsa
  ProxyCommand "/path/to/exec" ssh --stdio --context testcontext --user testuser testworkspace --devpod-home "C:\\White Space\devpod\test"
  User testuser
# DevPod End testhost`,
		},
		{
			name:       "Host addition with workdir",
			config:     "",
			execPath:   "/path/to/exec",
			host:       "testhost",
			user:       "testuser",
			context:    "testcontext",
			workspace:  "testworkspace",
			workdir:    "/path/to/workdir",
			command:    "",
			gpgagent:   false,
			devPodHome: "",
			provider:   "",
			expected: `# DevPod Start testhost
Host testhost
  ForwardAgent yes
  LogLevel error
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null
  HostKeyAlgorithms rsa-sha2-256,rsa-sha2-512,ssh-rsa
  ProxyCommand "/path/to/exec" ssh --stdio --context testcontext --user testuser testworkspace --workdir "/path/to/workdir"
  User testuser
# DevPod End testhost`,
		},
		{
			name:       "Host addition with gpg agent",
			config:     "",
			execPath:   "/path/to/exec",
			host:       "testhost",
			user:       "testuser",
			context:    "testcontext",
			workspace:  "testworkspace",
			workdir:    "",
			command:    "",
			gpgagent:   true,
			devPodHome: "",
			provider:   "",
			expected: `# DevPod Start testhost
Host testhost
  ForwardAgent yes
  LogLevel error
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null
  HostKeyAlgorithms rsa-sha2-256,rsa-sha2-512,ssh-rsa
  ProxyCommand "/path/to/exec" ssh --stdio --context testcontext --user testuser testworkspace --gpg-agent-forwarding
  User testuser
# DevPod End testhost`,
		},
		{
			name:       "Host addition with custom command",
			config:     "",
			execPath:   "/path/to/exec",
			host:       "testhost",
			user:       "testuser",
			context:    "testcontext",
			workspace:  "testworkspace",
			workdir:    "",
			command:    "ssh -W %h:%p bastion",
			gpgagent:   false,
			devPodHome: "",
			provider:   "",
			expected: `# DevPod Start testhost
Host testhost
  ForwardAgent yes
  LogLevel error
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null
  HostKeyAlgorithms rsa-sha2-256,rsa-sha2-512,ssh-rsa
  ProxyCommand "ssh -W %h:%p bastion"
  User testuser
# DevPod End testhost`,
		},
		{
			name: "Host addition to existing config",
			config: `Host existinghost
  User existinguser`,
			execPath:   "/path/to/exec",
			host:       "testhost",
			user:       "testuser",
			context:    "testcontext",
			workspace:  "testworkspace",
			workdir:    "",
			command:    "",
			gpgagent:   false,
			devPodHome: "",
			provider:   "",
			expected: `# DevPod Start testhost
Host testhost
  ForwardAgent yes
  LogLevel error
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null
  HostKeyAlgorithms rsa-sha2-256,rsa-sha2-512,ssh-rsa
  ProxyCommand "/path/to/exec" ssh --stdio --context testcontext --user testuser testworkspace
  User testuser
# DevPod End testhost
Host existinghost
  User existinguser`,
		},
		{
			name: "Host addition to existing config with DevPod host",
			// Config where an existing DevPod block is at the very top (no
			// preceding blank line). The new block must be inserted before the
			// existing one, directly after its Start marker line — because our
			// fix correctly does not treat "# DevPod Start" as a backtrack
			// comment, so the insert position lands right at "Host
			// existingtesthost" (index 1) and the new block goes in there.
			config: "# DevPod Start existingtesthost\nHost existingtesthost\n  ForwardAgent yes\n  LogLevel error\n  StrictHostKeyChecking no\n  UserKnownHostsFile /dev/null\n  HostKeyAlgorithms rsa-sha2-256,rsa-sha2-512,ssh-rsa\n  ProxyCommand \"/path/to/exec\" ssh --stdio --context testcontext --user testuser testworkspace\n  User testuser\n# DevPod End existingtesthost\n\nHost existinghost\n  User existinguser",
			execPath:   "/path/to/exec",
			host:       "testhost",
			user:       "testuser",
			context:    "testcontext",
			workspace:  "testworkspace",
			workdir:    "",
			command:    "",
			gpgagent:   false,
			devPodHome: "",
			provider:   "",
			expected:   "# DevPod Start existingtesthost\n# DevPod Start testhost\nHost testhost\n  ForwardAgent yes\n  LogLevel error\n  StrictHostKeyChecking no\n  UserKnownHostsFile /dev/null\n  HostKeyAlgorithms rsa-sha2-256,rsa-sha2-512,ssh-rsa\n  ProxyCommand \"/path/to/exec\" ssh --stdio --context testcontext --user testuser testworkspace\n  User testuser\n# DevPod End testhost\nHost existingtesthost\n  ForwardAgent yes\n  LogLevel error\n  StrictHostKeyChecking no\n  UserKnownHostsFile /dev/null\n  HostKeyAlgorithms rsa-sha2-256,rsa-sha2-512,ssh-rsa\n  ProxyCommand \"/path/to/exec\" ssh --stdio --context testcontext --user testuser testworkspace\n  User testuser\n# DevPod End existingtesthost\n\nHost existinghost\n  User existinguser",
			skipConfigContainsCheck: true,
		},
		{
			// Regression: lowercase "host" entries (valid SSH config) were
			// invisible to the old case-sensitive check, causing the insertion
			// point to land inside an existing DevPod block instead of before
			// the first host stanza.
			name:       "Host addition with lowercase host entries",
			config:     "host 192.168.1.1\n  User alice\n  Port 22\n\nhost myserver\n  User bob",
			execPath:   "/path/to/exec",
			host:       "testhost",
			user:       "testuser",
			context:    "testcontext",
			workspace:  "testworkspace",
			workdir:    "",
			command:    "",
			gpgagent:   false,
			devPodHome: "",
			provider:   "",
			expected:   "# DevPod Start testhost\nHost testhost\n  ForwardAgent yes\n  LogLevel error\n  StrictHostKeyChecking no\n  UserKnownHostsFile /dev/null\n  HostKeyAlgorithms rsa-sha2-256,rsa-sha2-512,ssh-rsa\n  ProxyCommand \"/path/to/exec\" ssh --stdio --context testcontext --user testuser testworkspace\n  User testuser\n# DevPod End testhost\nhost 192.168.1.1\n  User alice\n  Port 22\n\nhost myserver\n  User bob",
		},
		{
			// Regression: when an existing DevPod block was present, its
			// "# DevPod Start" marker was counted as a comment line and the
			// backtrack moved the insert position into it, corrupting the block.
			name:       "Host addition does not corrupt existing DevPod block followed by plain host",
			config:     "host plain-host\n  User alice\n\n# DevPod Start existing.devpod\nHost existing.devpod\n  ForwardAgent yes\n  LogLevel error\n  StrictHostKeyChecking no\n  UserKnownHostsFile /dev/null\n  HostKeyAlgorithms rsa-sha2-256,rsa-sha2-512,ssh-rsa\n  ProxyCommand \"/path/to/exec\" ssh --stdio --context ctx --user user existing\n  User user\n# DevPod End existing.devpod",
			execPath:   "/path/to/exec",
			host:       "testhost",
			user:       "testuser",
			context:    "testcontext",
			workspace:  "testworkspace",
			workdir:    "",
			command:    "",
			gpgagent:   false,
			devPodHome: "",
			provider:   "",
			expected:   "# DevPod Start testhost\nHost testhost\n  ForwardAgent yes\n  LogLevel error\n  StrictHostKeyChecking no\n  UserKnownHostsFile /dev/null\n  HostKeyAlgorithms rsa-sha2-256,rsa-sha2-512,ssh-rsa\n  ProxyCommand \"/path/to/exec\" ssh --stdio --context testcontext --user testuser testworkspace\n  User testuser\n# DevPod End testhost\nhost plain-host\n  User alice\n\n# DevPod Start existing.devpod\nHost existing.devpod\n  ForwardAgent yes\n  LogLevel error\n  StrictHostKeyChecking no\n  UserKnownHostsFile /dev/null\n  HostKeyAlgorithms rsa-sha2-256,rsa-sha2-512,ssh-rsa\n  ProxyCommand \"/path/to/exec\" ssh --stdio --context ctx --user user existing\n  User user\n# DevPod End existing.devpod",
		},
		{
			name: "Host addition after top level includes",
			config: `Include ~/config1

Include ~/config2



Include ~/config3`,
			execPath:   "/path/to/exec",
			host:       "testhost",
			user:       "testuser",
			context:    "testcontext",
			workspace:  "testworkspace",
			workdir:    "",
			command:    "",
			gpgagent:   false,
			devPodHome: "",
			provider:   "",
			expected: `Include ~/config1

Include ~/config2



Include ~/config3
# DevPod Start testhost
Host testhost
  ForwardAgent yes
  LogLevel error
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null
  HostKeyAlgorithms rsa-sha2-256,rsa-sha2-512,ssh-rsa
  ProxyCommand "/path/to/exec" ssh --stdio --context testcontext --user testuser testworkspace
  User testuser
# DevPod End testhost`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := addHostSection(tt.config, tt.execPath, addHostParams{
				path:       "",
				host:       tt.host,
				user:       tt.user,
				context:    tt.context,
				workspace:  tt.workspace,
				workdir:    tt.workdir,
				command:    tt.command,
				gpgagent:   tt.gpgagent,
				devPodHome: tt.devPodHome,
				provider:   tt.provider,
			})

			assert.NoError(s.T(), err)
			assert.Equal(s.T(), tt.expected, result)
			assert.Contains(s.T(), result, MarkerEndPrefix+tt.host)
			assert.Contains(s.T(), result, "Host "+tt.host)
			assert.Contains(s.T(), result, "User "+tt.user)

			if tt.command != "" {
				assert.Contains(s.T(), result, "ProxyCommand \""+tt.command+"\"")
			}

			if tt.workdir != "" {
				assert.Contains(s.T(), result, "--workdir \""+tt.workdir+"\"")
			}

			if tt.gpgagent {
				assert.Contains(s.T(), result, "--gpg-agent-forwarding")
			}

			if tt.config != "" && !tt.skipConfigContainsCheck {
					assert.Contains(s.T(), result, tt.config)
				}
		})
	}
}
