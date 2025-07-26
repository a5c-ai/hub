'use client';

import { useState, useEffect } from 'react';
import { repoApi } from '@/lib/api';
import { 
  RepositoryStatistics, 
  RepositoryLanguage, 
  ContributorStats,
  RepositoryStatsResponse 
} from '@/types';

interface UseRepositoryStatsReturn {
  statistics: RepositoryStatistics | null;
  languages: RepositoryLanguage[];
  contributors: ContributorStats[];
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

export const useRepositoryStats = (owner: string, repo: string): UseRepositoryStatsReturn => {
  const [statistics, setStatistics] = useState<RepositoryStatistics | null>(null);
  const [languages, setLanguages] = useState<RepositoryLanguage[]>([]);
  const [contributors, setContributors] = useState<ContributorStats[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = async () => {
    if (!owner || !repo) return;

    try {
      setLoading(true);
      setError(null);

      // Fetch repository statistics
      const [statsResponse, languagesResponse] = await Promise.all([
        repoApi.getRepositoryStatistics(owner, repo),
        repoApi.getRepositoryLanguages(owner, repo),
      ]);

      // Handle statistics response
      if (statsResponse.data) {
        const statsData = statsResponse.data as RepositoryStatsResponse;
        
        if (statsData.stats) {
          setStatistics(statsData.stats);
        }

        // Convert languages object to array format
        if (statsData.languages) {
          const languagesArray: RepositoryLanguage[] = Object.entries(statsData.languages).map(
            ([language, bytes]) => ({
              id: `${owner}-${repo}-${language}`,
              repository_id: statsData.stats?.repository_id || '',
              language,
              bytes,
              percentage: 0, // Will be calculated below
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
            })
          );

          // Calculate percentages
          const totalBytes = Object.values(statsData.languages).reduce((sum, bytes) => sum + bytes, 0);
          languagesArray.forEach(lang => {
            lang.percentage = totalBytes > 0 ? (lang.bytes / totalBytes) * 100 : 0;
          });

          setLanguages(languagesArray);
        }

        // Handle contributors
        if (statsData.contributors) {
          const contributorsData: ContributorStats[] = statsData.contributors.map(contributor => ({
            name: contributor.name,
            email: contributor.email,
            commits: contributor.commits,
            additions: 0, // Not provided by current API
            deletions: 0, // Not provided by current API
            first_commit_date: '', // Not provided by current API
            last_commit_date: '', // Not provided by current API
          }));
          setContributors(contributorsData);
        }
      }

      // If we got languages separately, process them
      if (languagesResponse.data && (!statsResponse.data || 
          (typeof statsResponse.data === 'object' && statsResponse.data !== null && !('languages' in statsResponse.data)))) {
        const languagesData = languagesResponse.data as Record<string, number>;
        const languagesArray: RepositoryLanguage[] = Object.entries(languagesData).map(
          ([language, bytes]) => ({
            id: `${owner}-${repo}-${language}`,
            repository_id: statistics?.repository_id || '',
            language,
            bytes,
            percentage: 0,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          })
        );

        // Calculate percentages
        const totalBytes = Object.values(languagesData).reduce((sum, bytes) => sum + bytes, 0);
        languagesArray.forEach(lang => {
          lang.percentage = totalBytes > 0 ? (lang.bytes / totalBytes) * 100 : 0;
        });

        setLanguages(languagesArray);
      }

    } catch (err) {
      console.error('Failed to fetch repository statistics:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch repository statistics');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [owner, repo]);

  const refetch = () => {
    fetchData();
  };

  return {
    statistics,
    languages,
    contributors,
    loading,
    error,
    refetch,
  };
};