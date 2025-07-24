import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { 
  Issue, 
  Comment, 
  Label, 
  Milestone, 
  CreateIssueRequest, 
  UpdateIssueRequest, 
  IssueFilters,
  PaginatedResponse
} from '@/types';
import { issueApi, commentApi, labelApi, milestoneApi } from '@/lib/api';

// Helper function to extract error message
const getErrorMessage = (error: unknown): string => {
  if (error instanceof Error) {
    return error.message;
  }
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { data?: { error?: string } } }).response;
    return response?.data?.error || 'An error occurred';
  }
  return 'An error occurred';
};

interface IssueState {
  // Issue listing state
  issues: Issue[];
  isLoadingIssues: boolean;
  issuesError: string | null;
  issuesTotal: number;
  currentPage: number;
  totalPages: number;
  filters: IssueFilters;

  // Current issue detail state
  currentIssue: Issue | null;
  isLoadingCurrentIssue: boolean;
  currentIssueError: string | null;

  // Comments state
  comments: Comment[];
  isLoadingComments: boolean;
  commentsError: string | null;
  commentsTotal: number;

  // Labels state
  labels: Label[];
  isLoadingLabels: boolean;
  labelsError: string | null;

  // Milestones state
  milestones: Milestone[];
  isLoadingMilestones: boolean;
  milestonesError: string | null;

  // Operations state
  isCreating: boolean;
  isUpdating: boolean;
  isDeleting: boolean;
  operationError: string | null;
}

interface IssueActions {
  // Issue operations
  fetchIssues: (owner: string, repo: string, filters?: IssueFilters) => Promise<void>;
  searchIssues: (owner: string, repo: string, query: string, filters?: IssueFilters) => Promise<void>;
  fetchIssue: (owner: string, repo: string, number: number) => Promise<void>;
  createIssue: (owner: string, repo: string, data: CreateIssueRequest) => Promise<Issue>;
  updateIssue: (owner: string, repo: string, number: number, data: UpdateIssueRequest) => Promise<void>;
  closeIssue: (owner: string, repo: string, number: number, reason?: string) => Promise<void>;
  reopenIssue: (owner: string, repo: string, number: number) => Promise<void>;
  lockIssue: (owner: string, repo: string, number: number, reason?: string) => Promise<void>;
  unlockIssue: (owner: string, repo: string, number: number) => Promise<void>;

  // Comment operations
  fetchComments: (owner: string, repo: string, issueNumber: number) => Promise<void>;
  createComment: (owner: string, repo: string, issueNumber: number, body: string) => Promise<void>;
  updateComment: (owner: string, repo: string, issueNumber: number, commentId: string, body: string) => Promise<void>;
  deleteComment: (owner: string, repo: string, issueNumber: number, commentId: string) => Promise<void>;

  // Labels operations
  fetchLabels: (owner: string, repo: string) => Promise<void>;

  // Milestones operations
  fetchMilestones: (owner: string, repo: string) => Promise<void>;

  // Filter and state management
  setFilters: (filters: Partial<IssueFilters>) => void;
  clearCurrentIssue: () => void;
  clearErrors: () => void;
}

type IssueStore = IssueState & IssueActions;

export const useIssueStore = create<IssueStore>()(
  devtools(
    (set, get) => ({
      // Initial state
      issues: [],
      isLoadingIssues: false,
      issuesError: null,
      issuesTotal: 0,
      currentPage: 1,
      totalPages: 1,
      filters: {
        state: 'open',
        sort: 'created',
        direction: 'desc',
        page: 1,
        per_page: 30,
      },

      currentIssue: null,
      isLoadingCurrentIssue: false,
      currentIssueError: null,

      comments: [],
      isLoadingComments: false,
      commentsError: null,
      commentsTotal: 0,

      labels: [],
      isLoadingLabels: false,
      labelsError: null,

      milestones: [],
      isLoadingMilestones: false,
      milestonesError: null,

      isCreating: false,
      isUpdating: false,
      isDeleting: false,
      operationError: null,

      // Issue operations
      fetchIssues: async (owner: string, repo: string, filters?: IssueFilters) => {
        set({ isLoadingIssues: true, issuesError: null });
        
        try {
          const finalFilters = { ...get().filters, ...filters };
          const response = await issueApi.getIssues(owner, repo, finalFilters) as PaginatedResponse<Issue>;
          
          set({
            issues: response.data,
            issuesTotal: response.pagination.total,
            currentPage: response.pagination.page,
            totalPages: response.pagination.total_pages,
            filters: finalFilters,
            isLoadingIssues: false,
          });
        } catch (error: unknown) {
          set({
            issuesError: getErrorMessage(error),
            isLoadingIssues: false,
          });
        }
      },

      searchIssues: async (owner: string, repo: string, query: string, filters?: IssueFilters) => {
        set({ isLoadingIssues: true, issuesError: null });
        
        try {
          const finalFilters = { ...get().filters, ...filters };
          const response = await issueApi.searchIssues(owner, repo, query, finalFilters) as PaginatedResponse<Issue>;
          
          set({
            issues: response.data,
            issuesTotal: response.pagination.total,
            currentPage: response.pagination.page,
            totalPages: response.pagination.total_pages,
            filters: finalFilters,
            isLoadingIssues: false,
          });
        } catch (error: unknown) {
          set({
            issuesError: getErrorMessage(error),
            isLoadingIssues: false,
          });
        }
      },

      fetchIssue: async (owner: string, repo: string, number: number) => {
        set({ isLoadingCurrentIssue: true, currentIssueError: null });
        
        try {
          const response = await issueApi.getIssue(owner, repo, number) as { data: Issue };
          set({
            currentIssue: response.data,
            isLoadingCurrentIssue: false,
          });
        } catch (error: unknown) {
          set({
            currentIssueError: getErrorMessage(error),
            isLoadingCurrentIssue: false,
          });
        }
      },

      createIssue: async (owner: string, repo: string, data: CreateIssueRequest) => {
        set({ isCreating: true, operationError: null });
        
        try {
          const response = await issueApi.createIssue(owner, repo, data) as { data: Issue };
          set({ isCreating: false });
          
          // Add the new issue to the list if it matches current filters
          const state = get();
          if (state.filters.state === 'open' || !state.filters.state) {
            set({ issues: [response.data, ...state.issues] });
          }
          
          return response.data;
        } catch (error: unknown) {
          set({
            operationError: getErrorMessage(error),
            isCreating: false,
          });
          throw error;
        }
      },

      updateIssue: async (owner: string, repo: string, number: number, data: UpdateIssueRequest) => {
        set({ isUpdating: true, operationError: null });
        
        try {
          const response = await issueApi.updateIssue(owner, repo, number, data) as { data: Issue };
          set({ isUpdating: false });
          
          // Update current issue if it's loaded
          const state = get();
          if (state.currentIssue && state.currentIssue.number === number) {
            set({ currentIssue: response.data });
          }
          
          // Update in issues list
          const updatedIssues = state.issues.map(issue => 
            issue.number === number ? response.data : issue
          );
          set({ issues: updatedIssues });
        } catch (error: unknown) {
          set({
            operationError: getErrorMessage(error),
            isUpdating: false,
          });
          throw error;
        }
      },

      closeIssue: async (owner: string, repo: string, number: number, reason?: string) => {
        await get().updateIssue(owner, repo, number, { state: 'closed', state_reason: reason });
      },

      reopenIssue: async (owner: string, repo: string, number: number) => {
        await get().updateIssue(owner, repo, number, { state: 'open' });
      },

      lockIssue: async (owner: string, repo: string, number: number, reason?: string) => {
        set({ isUpdating: true, operationError: null });
        
        try {
          const response = await issueApi.lockIssue(owner, repo, number, reason) as { data: Issue };
          set({ isUpdating: false });
          
          // Update current issue if it's loaded
          const state = get();
          if (state.currentIssue && state.currentIssue.number === number) {
            set({ currentIssue: response.data });
          }
        } catch (error: unknown) {
          set({
            operationError: getErrorMessage(error),
            isUpdating: false,
          });
          throw error;
        }
      },

      unlockIssue: async (owner: string, repo: string, number: number) => {
        set({ isUpdating: true, operationError: null });
        
        try {
          const response = await issueApi.unlockIssue(owner, repo, number) as { data: Issue };
          set({ isUpdating: false });
          
          // Update current issue if it's loaded
          const state = get();
          if (state.currentIssue && state.currentIssue.number === number) {
            set({ currentIssue: response.data });
          }
        } catch (error: unknown) {
          set({
            operationError: getErrorMessage(error),
            isUpdating: false,
          });
          throw error;
        }
      },

      // Comment operations
      fetchComments: async (owner: string, repo: string, issueNumber: number) => {
        set({ isLoadingComments: true, commentsError: null });
        
        try {
          const response = await commentApi.getComments(owner, repo, issueNumber) as PaginatedResponse<Comment>;
          set({
            comments: response.data,
            commentsTotal: response.pagination.total,
            isLoadingComments: false,
          });
        } catch (error: unknown) {
          set({
            commentsError: getErrorMessage(error),
            isLoadingComments: false,
          });
        }
      },

      createComment: async (owner: string, repo: string, issueNumber: number, body: string) => {
        set({ isCreating: true, operationError: null });
        
        try {
          const response = await commentApi.createComment(owner, repo, issueNumber, body) as { data: Comment };
          set({ isCreating: false });
          
          // Add the new comment to the list
          const state = get();
          set({ 
            comments: [...state.comments, response.data],
            commentsTotal: state.commentsTotal + 1,
          });
          
          // Update issue comment count if current issue is loaded
          if (state.currentIssue && state.currentIssue.number === issueNumber) {
            set({
              currentIssue: {
                ...state.currentIssue,
                comments_count: state.currentIssue.comments_count + 1,
              },
            });
          }
        } catch (error: unknown) {
          set({
            operationError: getErrorMessage(error),
            isCreating: false,
          });
          throw error;
        }
      },

      updateComment: async (owner: string, repo: string, issueNumber: number, commentId: string, body: string) => {
        set({ isUpdating: true, operationError: null });
        
        try {
          const response = await commentApi.updateComment(owner, repo, issueNumber, commentId, body) as { data: Comment };
          set({ isUpdating: false });
          
          // Update comment in list
          const state = get();
          const updatedComments = state.comments.map(comment =>
            comment.id === commentId ? response.data : comment
          );
          set({ comments: updatedComments });
        } catch (error: unknown) {
          set({
            operationError: getErrorMessage(error),
            isUpdating: false,
          });
          throw error;
        }
      },

      deleteComment: async (owner: string, repo: string, issueNumber: number, commentId: string) => {
        set({ isDeleting: true, operationError: null });
        
        try {
          await commentApi.deleteComment(owner, repo, issueNumber, commentId);
          set({ isDeleting: false });
          
          // Remove comment from list
          const state = get();
          const updatedComments = state.comments.filter(comment => comment.id !== commentId);
          set({ 
            comments: updatedComments,
            commentsTotal: state.commentsTotal - 1,
          });
          
          // Update issue comment count if current issue is loaded
          if (state.currentIssue && state.currentIssue.number === issueNumber) {
            set({
              currentIssue: {
                ...state.currentIssue,
                comments_count: state.currentIssue.comments_count - 1,
              },
            });
          }
        } catch (error: unknown) {
          set({
            operationError: getErrorMessage(error),
            isDeleting: false,
          });
          throw error;
        }
      },

      // Labels operations
      fetchLabels: async (owner: string, repo: string) => {
        set({ isLoadingLabels: true, labelsError: null });
        
        try {
          const response = await labelApi.getLabels(owner, repo) as PaginatedResponse<Label>;
          set({
            labels: response.data,
            isLoadingLabels: false,
          });
        } catch (error: unknown) {
          set({
            labelsError: getErrorMessage(error),
            isLoadingLabels: false,
          });
        }
      },

      // Milestones operations
      fetchMilestones: async (owner: string, repo: string) => {
        set({ isLoadingMilestones: true, milestonesError: null });
        
        try {
          const response = await milestoneApi.getMilestones(owner, repo) as PaginatedResponse<Milestone>;
          set({
            milestones: response.data,
            isLoadingMilestones: false,
          });
        } catch (error: unknown) {
          set({
            milestonesError: getErrorMessage(error),
            isLoadingMilestones: false,
          });
        }
      },

      // Filter and state management
      setFilters: (filters: Partial<IssueFilters>) => {
        const currentFilters = get().filters;
        set({ filters: { ...currentFilters, ...filters } });
      },

      clearCurrentIssue: () => {
        set({
          currentIssue: null,
          currentIssueError: null,
          comments: [],
          commentsError: null,
          commentsTotal: 0,
        });
      },

      clearErrors: () => {
        set({
          issuesError: null,
          currentIssueError: null,
          commentsError: null,
          labelsError: null,
          milestonesError: null,
          operationError: null,
        });
      },
    }),
    {
      name: 'issue-store',
    }
  )
);