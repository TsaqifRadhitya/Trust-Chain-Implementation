import { useQuery } from '@tanstack/react-query';
import { fetchReports } from '../../../../modules/report';

export function useReports() {
  const reportsQuery = useQuery({
    queryKey: ['reports'],
    queryFn: fetchReports,
  });

  return {
    reports: reportsQuery.data || [],
    isLoading: reportsQuery.isLoading,
    error: reportsQuery.error,
  };
}
