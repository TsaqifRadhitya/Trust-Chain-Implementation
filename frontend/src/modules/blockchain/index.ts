import type { Case, CaseStatus, Transaction, TrendData, VerifyTxResponse, ChainValidationResponse } from './type';
import { apiClient } from '../../lib/axios';

export const FLAG_REASONS: Record<string, string> = {
  'Volume Anomaly': 'Transaction volume is 400% higher than the historical average for this partner.',
  'Geographic Mismatch': 'Geographic mismatch detected in origin IP vs. registered address.',
  'Duplicate Detection': 'Possible duplicate transaction detected within a 2-hour window.',
  'Velocity Check': 'Transaction frequency exceeds normal velocity threshold for this account.',
};

function getStoredCaseStatuses(): Record<string, CaseStatus> {
  const stored = localStorage.getItem('tc_case_statuses');
  if (stored) {
    try {
      return JSON.parse(stored);
    } catch {
      // Fallback
    }
  }
  return {};
}

interface BackendTransaction {
  hash: string;
  block_height: number;
  timestamp: string;
  status: string;
  from: string;
  to: string;
  value: number;
  data: string;
  is_fraud: boolean;
  risk_score: number;
  flag_reason: string;
  verdict: string;
}

export async function fetchCases(): Promise<Case[]> {
  try {
    const response = await apiClient.get('/explorer/transactions?limit=100');
    const txs: BackendTransaction[] = response.data.data || [];
    
    // Filter fraud transactions
    const fraudTxs = txs.filter((tx) => tx.is_fraud === true);
    
    const statuses = getStoredCaseStatuses();
    
    return fraudTxs.map((tx) => {
      // Extract short ID
      const shortId = `CASE-${tx.hash.substring(2, 6).toUpperCase()}`;
      
      // Try to parse original ERP payload for context
      let partner = tx.to || 'Unknown Vendor';
      try {
        const payload = JSON.parse(tx.data) as { vendor_name?: string };
        partner = payload.vendor_name || partner;
      } catch {
        // Ignored
      }

      // Find saved status, default to Open
      const savedStatus = statuses[shortId] || 'Open';

      // Format currency (IDR)
      const amountFormat = new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR'
      }).format(tx.value);

      return {
        id: shortId,
        txId: `TX-${tx.hash.substring(2, 8).toUpperCase()}`,
        date: new Date(tx.timestamp).toISOString().split('T')[0],
        status: savedStatus,
        partner: partner,
        amount: amountFormat,
        risk: tx.risk_score,
        type: tx.flag_reason || tx.verdict || 'Anomaly',
        originalHash: tx.hash
      };
    });
  } catch (error) {
    console.error('Error fetching cases from API:', error);
    return [];
  }
}

export async function updateCaseStatus(caseId: string, status: CaseStatus): Promise<Case> {
  const statuses = getStoredCaseStatuses();
  statuses[caseId] = status;
  localStorage.setItem('tc_case_statuses', JSON.stringify(statuses));
  
  return { id: caseId, status } as Case;
}

export async function fetchTrendData(): Promise<TrendData[]> {
  try {
    const response = await apiClient.get('/explorer/transactions?limit=50');
    const txs: BackendTransaction[] = response.data.data || [];
    
    // Sort transactions chronologically (oldest first for charts)
    const sortedTxs = [...txs].reverse();

    return sortedTxs.map((tx) => {
      const timeStr = new Date(tx.timestamp).toLocaleTimeString('en-US', {
        hour: '2-digit',
        minute: '2-digit',
        hour12: false
      });
      return {
        time: timeStr,
        risk: tx.risk_score
      };
    });
  } catch (error) {
    console.error('Error fetching trend data from API:', error);
    return [];
  }
}

export async function fetchLiveTransactions(): Promise<Transaction[]> {
  try {
    const response = await apiClient.get('/explorer/transactions?limit=20');
    const txs: BackendTransaction[] = response.data.data || [];
    
    return txs.map((tx) => {
      let partner = tx.to || 'Unknown Vendor';
      try {
        const payload = JSON.parse(tx.data) as { vendor_name?: string };
        partner = payload.vendor_name || partner;
      } catch {
        // Ignored
      }

      // Format currency (IDR)
      const amountFormat = new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR'
      }).format(tx.value);

      // Determine visual status badge
      let status: 'safe' | 'warning' | 'flagged' = 'safe';
      if (tx.is_fraud) {
        status = 'flagged';
      } else if (tx.risk_score >= 30) {
        status = 'warning';
      }

      return {
        id: `TX-${tx.hash.substring(2, 8).toUpperCase()}`,
        partner: partner,
        amount: amountFormat,
        status: status,
        aiScore: tx.risk_score
      };
    });
  } catch (error) {
    console.error('Error fetching live transactions from API:', error);
    return [];
  }
}

export async function verifyTx(hash: string): Promise<VerifyTxResponse> {
  try {
    const response = await apiClient.get(`/explorer/transactions/${hash}`);
    const tx = response.data.data;
    
    let payload = {};
    try {
      payload = JSON.parse(tx.data) as Record<string, unknown>;
    } catch {
      // Ignored
    }

    return {
      hash: tx.hash,
      status: tx.status === 'success' ? 'Success' : 'Failed',
      timestamp: new Date(tx.timestamp).toLocaleString('en-US', { timeZoneName: 'short' }),
      from: tx.from,
      to: tx.to,
      blockHeight: tx.block_height,
      payload: {
        ...payload,
        aiScore: tx.risk_score,
        flags: tx.is_fraud ? [tx.flag_reason || tx.verdict] : [],
      },
    };
  } catch (error) {
    console.error('Error verifying transaction:', error);
    throw new Error('Transaction not found', { cause: error });
  }
}

export async function validateChain(): Promise<ChainValidationResponse> {
  try {
    const response = await apiClient.get('/explorer/blockchain/validate');
    return response.data.data;
  } catch (error) {
    console.error('Error validating blockchain chain:', error);
    throw error;
  }
}

export async function fetchDashboardStats() {
  try {
    const [statsRes, valRes] = await Promise.all([
      apiClient.get('/explorer/stats'),
      apiClient.get('/explorer/blockchain/validate')
    ]);
    
    const statsData = statsRes.data.data || { total_transactions: 0, total_anomalies: 0 };
    
    const validation = valRes.data.data;
    const verified = validation && validation.is_valid ? '100%' : 'Failed';
    
    return {
      processed: statsData.total_transactions,
      anomalies: statsData.total_anomalies,
      verified
    };
  } catch (error) {
    console.error('Error fetching dashboard stats:', error);
    return { processed: 0, anomalies: 0, verified: '100%' };
  }
}
