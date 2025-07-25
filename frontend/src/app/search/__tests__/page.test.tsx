import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { useSearchParams } from 'next/navigation';
import SearchPage from '../page';
import { searchApi } from '@/lib/api';

// Mock Next.js router
jest.mock('next/navigation', () => ({
  useSearchParams: jest.fn(),
}));

// Mock the search API
jest.mock('@/lib/api', () => ({
  searchApi: {
    globalSearch: jest.fn(),
  },
}));

const mockUseSearchParams = useSearchParams as jest.MockedFunction<typeof useSearchParams>;
const mockGlobalSearch = searchApi.globalSearch as jest.MockedFunction<typeof searchApi.globalSearch>;

describe('SearchPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUseSearchParams.mockReturnValue({
      get: jest.fn(() => null),
    } as unknown as URLSearchParams);
  });

  it('renders search form correctly', () => {
    render(<SearchPage />);
    
    expect(screen.getByPlaceholderText('Search repositories, issues, users, and commits...')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Search' })).toBeInTheDocument();
  });

  it('renders search type filters', () => {
    render(<SearchPage />);
    
    expect(screen.getByRole('button', { name: 'All' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Repositories' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Issues' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Users' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Commits' })).toBeInTheDocument();
  });

  it('handles search input changes', () => {
    render(<SearchPage />);
    
    const searchInput = screen.getByPlaceholderText('Search repositories, issues, users, and commits...');
    fireEvent.change(searchInput, { target: { value: 'test query' } });
    
    expect((searchInput as HTMLInputElement).value).toBe('test query');
  });

  it('performs search when form is submitted', async () => {
    const mockResults = {
      data: {
        users: [],
        repositories: [
          {
            id: '1',
            name: 'test-repo',
            description: 'Test repository',
            owner_id: 'user1',
            owner_type: 'user',
            visibility: 'public',
            stars_count: 10,
            forks_count: 5,
            primary_language: 'JavaScript',
            created_at: '2023-01-01',
            updated_at: '2023-01-02',
          },
        ],
        issues: [],
        organizations: [],
        commits: [],
        total_count: 1,
      },
    };

    mockGlobalSearch.mockResolvedValue(mockResults);

    render(<SearchPage />);
    
    const searchInput = screen.getByPlaceholderText('Search repositories, issues, users, and commits...');
    const searchButton = screen.getByRole('button', { name: 'Search' });
    
    fireEvent.change(searchInput, { target: { value: 'test' } });
    fireEvent.click(searchButton);

    await waitFor(() => {
      expect(mockGlobalSearch).toHaveBeenCalledWith('test', {
        type: undefined,
        page: 1,
        per_page: 30,
      });
    });

    await waitFor(() => {
      expect(screen.getByText('Repositories')).toBeInTheDocument();
      expect(screen.getByText('test-repo')).toBeInTheDocument();
    });
  });

  it('handles search type changes', async () => {
    const mockResults = {
      data: {
        users: [
          {
            id: '1',
            username: 'testuser',
            full_name: 'Test User',
            email: 'test@example.com',
            bio: 'Test bio',
            avatar_url: '',
            company: 'Test Company',
            location: 'Test Location',
          },
        ],
        repositories: [],
        issues: [],
        organizations: [],
        commits: [],
        total_count: 1,
      },
    };

    mockGlobalSearch.mockResolvedValue(mockResults);

    render(<SearchPage />);
    
    const searchInput = screen.getByPlaceholderText('Search repositories, issues, users, and commits...');
    const usersButton = screen.getByRole('button', { name: 'Users' });
    
    fireEvent.change(searchInput, { target: { value: 'test' } });
    fireEvent.click(usersButton);

    await waitFor(() => {
      expect(mockGlobalSearch).toHaveBeenCalledWith('test', {
        type: 'users',
        page: 1,
        per_page: 30,
      });
    });

    await waitFor(() => {
      expect(screen.getByText('Users')).toBeInTheDocument();
      expect(screen.getByText('testuser')).toBeInTheDocument();
    });
  });

  it('displays loading state during search', async () => {
    mockGlobalSearch.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));

    render(<SearchPage />);
    
    const searchInput = screen.getByPlaceholderText('Search repositories, issues, users, and commits...');
    const searchButton = screen.getByRole('button', { name: 'Search' });
    
    fireEvent.change(searchInput, { target: { value: 'test' } });
    fireEvent.click(searchButton);

    expect(screen.getByText('Searching...')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Searching...' })).toBeDisabled();
  });

  it('handles search errors', async () => {
    mockGlobalSearch.mockRejectedValue(new Error('Search failed'));

    render(<SearchPage />);
    
    const searchInput = screen.getByPlaceholderText('Search repositories, issues, users, and commits...');
    const searchButton = screen.getByRole('button', { name: 'Search' });
    
    fireEvent.change(searchInput, { target: { value: 'test' } });
    fireEvent.click(searchButton);

    await waitFor(() => {
      expect(screen.getByText('Failed to perform search. Please try again.')).toBeInTheDocument();
    });
  });

  it('displays no results message when search returns empty', async () => {
    const mockResults = {
      data: {
        users: [],
        repositories: [],
        issues: [],
        organizations: [],
        commits: [],
        total_count: 0,
      },
    };

    mockGlobalSearch.mockResolvedValue(mockResults);

    render(<SearchPage />);
    
    const searchInput = screen.getByPlaceholderText('Search repositories, issues, users, and commits...');
    const searchButton = screen.getByRole('button', { name: 'Search' });
    
    fireEvent.change(searchInput, { target: { value: 'nonexistent' } });
    fireEvent.click(searchButton);

    await waitFor(() => {
      expect(screen.getByText('No results found')).toBeInTheDocument();
      expect(screen.getByText('Try adjusting your search query or search in a different category.')).toBeInTheDocument();
    });
  });

  it('initializes search from URL parameters', () => {
    mockUseSearchParams.mockReturnValue({
      get: jest.fn((param) => param === 'q' ? 'initial query' : null),
    } as unknown as URLSearchParams);

    const mockResults = {
      data: {
        users: [],
        repositories: [],
        issues: [],
        organizations: [],
        commits: [],
        total_count: 0,
      },
    };

    mockGlobalSearch.mockResolvedValue(mockResults);

    render(<SearchPage />);
    
    const searchInput = screen.getByPlaceholderText('Search repositories, issues, users, and commits...');
    expect((searchInput as HTMLInputElement).value).toBe('initial query');
  });

  it('displays repository results with correct formatting', async () => {
    const mockResults = {
      data: {
        users: [],
        repositories: [
          {
            id: '1',
            name: 'awesome-project',
            description: 'An awesome web application',
            owner_id: 'user1',
            owner_type: 'user',
            visibility: 'public',
            stars_count: 150,
            forks_count: 45,
            primary_language: 'TypeScript',
            created_at: '2023-01-01',
            updated_at: '2023-01-02',
          },
        ],
        issues: [],
        organizations: [],
        commits: [],
        total_count: 1,
      },
    };

    mockGlobalSearch.mockResolvedValue(mockResults);

    render(<SearchPage />);
    
    const searchInput = screen.getByPlaceholderText('Search repositories, issues, users, and commits...');
    fireEvent.change(searchInput, { target: { value: 'awesome' } });
    fireEvent.submit(searchInput.closest('form')!);

    await waitFor(() => {
      expect(screen.getByText('awesome-project')).toBeInTheDocument();
      expect(screen.getByText('An awesome web application')).toBeInTheDocument();
      expect(screen.getByText('TypeScript')).toBeInTheDocument();
      expect(screen.getByText('⭐ 150')).toBeInTheDocument();
      expect(screen.getByText('🍴 45')).toBeInTheDocument();
      expect(screen.getByText('public')).toBeInTheDocument();
    });
  });

  it('displays issue results with correct formatting', async () => {
    const mockResults = {
      data: {
        users: [],
        repositories: [],
        issues: [
          {
            id: '1',
            number: 42,
            title: 'Fix search functionality',
            body: 'The search feature needs improvements...',
            state: 'open',
            repository_id: 'repo1',
            user_id: 'user1',
            created_at: '2023-01-01',
            updated_at: '2023-01-02',
          },
        ],
        organizations: [],
        commits: [],
        total_count: 1,
      },
    };

    mockGlobalSearch.mockResolvedValue(mockResults);

    render(<SearchPage />);
    
    const searchInput = screen.getByPlaceholderText('Search repositories, issues, users, and commits...');
    fireEvent.change(searchInput, { target: { value: 'search' } });
    fireEvent.submit(searchInput.closest('form')!);

    await waitFor(() => {
      expect(screen.getByText('Fix search functionality')).toBeInTheDocument();
      expect(screen.getByText('#42 opened on 1/1/2023')).toBeInTheDocument();
      expect(screen.getByText(/The search feature needs improvements/)).toBeInTheDocument();
    });
  });
});