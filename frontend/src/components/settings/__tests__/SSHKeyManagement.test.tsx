import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { SSHKeyManagement } from '../SSHKeyManagement';
import { sshKeyApi } from '@/lib/api';

// Mock the API
jest.mock('@/lib/api', () => ({
  sshKeyApi: {
    getSSHKeys: jest.fn(),
    createSSHKey: jest.fn(),
    deleteSSHKey: jest.fn(),
  },
}));

const mockSSHKeys = [
  {
    id: '1',
    title: 'My laptop key',
    fingerprint: 'SHA256:abc123def456',
    key_type: 'ssh-ed25519',
    last_used_at: '2024-01-15T10:30:00Z',
    created_at: '2024-01-01T00:00:00Z',
  },
  {
    id: '2',
    title: 'Work desktop',
    fingerprint: 'SHA256:xyz789uvw012',
    key_type: 'ssh-rsa',
    last_used_at: null,
    created_at: '2024-01-10T00:00:00Z',
  },
];

describe('SSHKeyManagement', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders loading state initially', () => {
    (sshKeyApi.getSSHKeys as jest.Mock).mockImplementation(() => new Promise(() => {}));

    render(<SSHKeyManagement />);

    expect(screen.getByText('Loading SSH keys...')).toBeInTheDocument();
  });

  test('renders SSH keys list when loaded', async () => {
    (sshKeyApi.getSSHKeys as jest.Mock).mockResolvedValue(mockSSHKeys);

    render(<SSHKeyManagement />);

    await waitFor(() => {
      expect(screen.getByText('My laptop key')).toBeInTheDocument();
      expect(screen.getByText('Work desktop')).toBeInTheDocument();
    });
  });

  test('renders empty state when no SSH keys', async () => {
    (sshKeyApi.getSSHKeys as jest.Mock).mockResolvedValue([]);

    render(<SSHKeyManagement />);

    await waitFor(() => {
      expect(screen.getByText('No SSH keys')).toBeInTheDocument();
      expect(
        screen.getByText(
          "You haven't added any SSH keys yet. Add one to securely access your repositories."
        )
      ).toBeInTheDocument();
    });
  });

  test('shows error message when loading fails', async () => {
    (sshKeyApi.getSSHKeys as jest.Mock).mockRejectedValue({
      response: { data: { error: 'Failed to fetch keys' } },
    });

    render(<SSHKeyManagement />);

    await waitFor(() => {
      expect(screen.getByText('Failed to fetch keys')).toBeInTheDocument();
    });
  });

  test('opens add SSH key modal when button is clicked', async () => {
    (sshKeyApi.getSSHKeys as jest.Mock).mockResolvedValue(mockSSHKeys);

    render(<SSHKeyManagement />);

    await waitFor(() => {
      expect(screen.getByText('My laptop key')).toBeInTheDocument();
    });

    fireEvent.click(screen.getAllByText('Add SSH Key')[0]);

    await waitFor(() => {
      expect(screen.getByLabelText('Title')).toBeInTheDocument();
      expect(screen.getByLabelText('SSH Public Key')).toBeInTheDocument();
    });
  });

  test('validates SSH key format', async () => {
    (sshKeyApi.getSSHKeys as jest.Mock).mockResolvedValue([]);

    render(<SSHKeyManagement />);

    await waitFor(() => {
      expect(screen.getByText('No SSH keys')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Add your first SSH key'));

    await waitFor(() => {
      expect(screen.getByLabelText('Title')).toBeInTheDocument();
    });

    const titleInput = screen.getByLabelText('Title');
    const keyInput = screen.getByLabelText('SSH Public Key');
    const submitButton = screen
      .getAllByRole('button', { name: /Add SSH Key/ })
      .find((btn) => !(btn as HTMLButtonElement).disabled);

    // Fill in title but invalid key
    fireEvent.change(titleInput, { target: { value: 'Test key' } });
    fireEvent.change(keyInput, { target: { value: 'invalid-key-format' } });

    expect(submitButton).toBeDisabled();

    // Fill in valid key
    fireEvent.change(keyInput, {
      target: {
        value:
          'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIG4rT3vTt99Ox5kndS4HmgTrKBT8SKzhK4rhGkEVGlCI test@example.com',
      },
    });

    expect(submitButton).not.toBeDisabled();
  });

  test('creates new SSH key successfully', async () => {
    const newKey = {
      id: '3',
      title: 'New key',
      fingerprint: 'SHA256:new123key456',
      key_type: 'ssh-ed25519',
      last_used_at: null,
      created_at: '2024-01-20T00:00:00Z',
    };

    (sshKeyApi.getSSHKeys as jest.Mock).mockResolvedValue([]);
    (sshKeyApi.createSSHKey as jest.Mock).mockResolvedValue(newKey);

    render(<SSHKeyManagement />);

    await waitFor(() => {
      expect(screen.getByText('No SSH keys')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Add your first SSH key'));

    await waitFor(() => {
      expect(screen.getByLabelText('Title')).toBeInTheDocument();
    });

    const titleInput = screen.getByLabelText('Title');
    const keyInput = screen.getByLabelText('SSH Public Key');

    fireEvent.change(titleInput, { target: { value: 'New key' } });
    fireEvent.change(keyInput, {
      target: {
        value:
          'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIG4rT3vTt99Ox5kndS4HmgTrKBT8SKzhK4rhGkEVGlCI test@example.com',
      },
    });

    const submitButton = screen
      .getAllByRole('button', { name: /Add SSH Key/ })
      .find((btn) => !(btn as HTMLButtonElement).disabled);
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(sshKeyApi.createSSHKey).toHaveBeenCalledWith({
        title: 'New key',
        key_data:
          'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIG4rT3vTt99Ox5kndS4HmgTrKBT8SKzhK4rhGkEVGlCI test@example.com',
      });
    });
  });

  test('deletes SSH key successfully', async () => {
    (sshKeyApi.getSSHKeys as jest.Mock).mockResolvedValue(mockSSHKeys);
    (sshKeyApi.deleteSSHKey as jest.Mock).mockResolvedValue({});

    render(<SSHKeyManagement />);

    await waitFor(() => {
      expect(screen.getByText('My laptop key')).toBeInTheDocument();
    });

    const deleteButtons = screen.getAllByText('Delete');
    fireEvent.click(deleteButtons[0]);

    await waitFor(() => {
      expect(sshKeyApi.deleteSSHKey).toHaveBeenCalledWith('1');
    });
  });

  test('displays correct key type icons', async () => {
    (sshKeyApi.getSSHKeys as jest.Mock).mockResolvedValue(mockSSHKeys);

    render(<SSHKeyManagement />);

    await waitFor(() => {
      // Check that key type badges are displayed
      expect(screen.getByText('ssh-ed25519')).toBeInTheDocument();
      expect(screen.getByText('ssh-rsa')).toBeInTheDocument();
    });
  });

  test('displays fingerprints and dates correctly', async () => {
    (sshKeyApi.getSSHKeys as jest.Mock).mockResolvedValue(mockSSHKeys);

    render(<SSHKeyManagement />);

    await waitFor(() => {
      expect(screen.getByText('SHA256:abc123def456')).toBeInTheDocument();
      expect(screen.getByText('SHA256:xyz789uvw012')).toBeInTheDocument();

      // Check last used dates
      expect(screen.getByText('1/15/2024')).toBeInTheDocument(); // last_used_at for first key
      expect(screen.getByText('Never')).toBeInTheDocument(); // null last_used_at for second key
    });
  });
});
