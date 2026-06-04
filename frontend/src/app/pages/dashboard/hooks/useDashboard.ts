import { useQuery } from '@tanstack/react-query';
import { fetchTrendData, fetchLiveTransactions, fetchDashboardStats } from '../../../../modules/blockchain';
import { QUERY_KEYS } from '../../../../constant/query-key';

export function useDashboard() {
  const trendQuery = useQuery({
    queryKey: [QUERY_KEYS.TREND_DATA],
    queryFn: fetchTrendData,
    refetchInterval: 10000,
  });

  const txQuery = useQuery({
    queryKey: [QUERY_KEYS.LIVE_TRANSACTIONS],
    queryFn: fetchLiveTransactions,
    refetchInterval: 5000,
  });

  const statsQuery = useQuery({
    queryKey: ['DASHBOARD_STATS'],
    queryFn: fetchDashboardStats,
    refetchInterval: 10000,
  });

  return {
    trendData: trendQuery.data || [],
    isLoadingTrend: trendQuery.isLoading,
    liveTxs: txQuery.data || [],
    isLoadingTxs: txQuery.isLoading,
    stats: statsQuery.data || { processed: 0, anomalies: 0, verified: '100%' },
    isLoadingStats: statsQuery.isLoading,
  };
}
