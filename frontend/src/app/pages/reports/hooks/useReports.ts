import { useQuery } from '@tanstack/react-query';
import { fetchReports } from '../../../../modules/report';
import { QUERY_KEYS } from '../../../../constant/query-key';

export function useReports() {
  const reportsQuery = useQuery({
    queryKey: [QUERY_KEYS.REPORTS],
    queryFn: fetchReports,
  });

  return {
    reports: reportsQuery.data || [],
    isLoading: reportsQuery.isLoading,
    error: reportsQuery.error,
  };
}
