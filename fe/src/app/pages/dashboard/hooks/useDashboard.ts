import { useQuery } from '@tanstack/react-query';
import { useEffect, useState } from 'react';
import { fetchTrendData, fetchLiveTransactions } from '../../../../modules/blockchain';
import type { Transaction } from '../../../../modules/blockchain/type';

export function useDashboard() {
  const trendQuery = useQuery({
    queryKey: ['trendData'],
    queryFn: fetchTrendData,
  });

  const txQuery = useQuery({
    queryKey: ['liveTransactions'],
    queryFn: fetchLiveTransactions,
  });

  const [txs, setTxs] = useState<Transaction[]>([]);

  useEffect(() => {
    if (txQuery.data) {
      setTxs(txQuery.data);
    }
  }, [txQuery.data]);

  useEffect(() => {
    if (txs.length === 0) return;
    const interval = setInterval(() => {
      const isAnomaly = Math.random() > 0.8;
      const newTx: Transaction = {
        id: `TX-${Math.floor(1000 + Math.random() * 9000)}`,
        partner: ['LogisX Energy', 'Global Manuf.', 'Neo Supply', 'Apex Corp', 'CyberLog'][Math.floor(Math.random() * 5)],
        amount: `$${(Math.random() * 500000).toLocaleString('en-US', { minimumFractionDigits: 2 })}`,
        status: isAnomaly ? 'flagged' : (Math.random() > 0.7 ? 'warning' : 'safe'),
        aiScore: isAnomaly ? Math.floor(75 + Math.random() * 20) : Math.floor(5 + Math.random() * 50),
      };
      setTxs((prev) => [newTx, ...prev.slice(0, 3)]);
    }, 4000);
    return () => clearInterval(interval);
  }, [txs.length]);

  return {
    trendData: trendQuery.data || [],
    isLoadingTrend: trendQuery.isLoading,
    liveTxs: txs,
    isLoadingTxs: txQuery.isLoading,
  };
}
