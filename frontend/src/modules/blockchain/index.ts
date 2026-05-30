import type { Case, CaseStatus, Transaction, TrendData, VerifyTxResponse } from './type';

const INITIAL_CASES: Case[] = [
  { id: 'CASE-092', txId: 'TX-8823', date: '2026-04-24', status: 'Open', partner: 'Neo Supply', amount: '$2,100,000.00', risk: 89, type: 'Volume Anomaly', originalHash: '0x1234567890abcdef' },
  { id: 'CASE-091', txId: 'TX-8110', date: '2026-04-23', status: 'In Review', partner: 'Apex Corp', amount: '$450,000.00', risk: 72, type: 'Geographic Mismatch', originalHash: '0xabcdef1234567890' },
  { id: 'CASE-090', txId: 'TX-7994', date: '2026-04-22', status: 'Resolved', partner: 'LogisX Energy', amount: '$12,400.00', risk: 45, type: 'Velocity Check', originalHash: '0xdeadbeefdeadbeef' },
  { id: 'CASE-089', txId: 'TX-7821', date: '2026-04-21', status: 'Open', partner: 'CyberLog', amount: '$890,000.00', risk: 81, type: 'Duplicate Detection', originalHash: '0xcafebabecafebabe' },
  { id: 'CASE-088', txId: 'TX-7700', date: '2026-04-20', status: 'In Review', partner: 'Global Manuf.', amount: '$312,000.00', risk: 63, type: 'Volume Anomaly', originalHash: '0x0987654321fedcba' },
];

export const FLAG_REASONS: Record<string, string> = {
  'Volume Anomaly': 'Transaction volume is 400% higher than the historical average for this partner.',
  'Geographic Mismatch': 'Geographic mismatch detected in origin IP vs. registered address.',
  'Duplicate Detection': 'Possible duplicate transaction detected within a 2-hour window.',
  'Velocity Check': 'Transaction frequency exceeds normal velocity threshold for this account.',
};

const CHART_TREND_DATA: TrendData[] = [
  { time: '10:00', risk: 12 },
  { time: '10:05', risk: 18 },
  { time: '10:10', risk: 15 },
  { time: '10:15', risk: 35 },
  { time: '10:20', risk: 25 },
  { time: '10:25', risk: 65 }, // anomaly
  { time: '10:30', risk: 42 },
  { time: '10:35', risk: 20 },
];

const INITIAL_TX_FEED: Transaction[] = [
  { id: 'TX-8821', partner: 'LogisX Energy', amount: '$124,000.00', status: 'safe', aiScore: 12 },
  { id: 'TX-8822', partner: 'Global Manuf.', amount: '$45,500.00', status: 'safe', aiScore: 8 },
  { id: 'TX-8823', partner: 'Neo Supply', amount: '$2,100,000.00', status: 'flagged', aiScore: 89 },
  { id: 'TX-8824', partner: 'Apex Corp', amount: '$12,400.00', status: 'warning', aiScore: 45 },
];

import { apiClient } from '../../lib/axios';

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
    
    if (fraudTxs.length === 0) {
      return [{
        id: 'DBG-001',
        txId: 'TX-DEBUG',
        date: new Date().toISOString().split('T')[0],
        status: 'Open',
        partner: `No fraud found in ${txs.length} txs`,
        amount: 'Rp0',
        risk: 99,
        type: 'DEBUG',
        originalHash: '0x000'
      }];
    }

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
        originalHash: tx.hash // keep original hash for reference
      };
    });
  } catch (error) {
    console.error('Error fetching cases from API:', error);
    // Fallback to initial cases if API fails
    return INITIAL_CASES;
  }
}

export async function updateCaseStatus(caseId: string, status: CaseStatus): Promise<Case> {
  const statuses = getStoredCaseStatuses();
  statuses[caseId] = status;
  localStorage.setItem('tc_case_statuses', JSON.stringify(statuses));
  
  // We need to return a dummy case to satisfy the type, 
  // the hook will invalidate and refetch anyway.
  return { id: caseId, status } as Case;
}

export async function fetchTrendData(): Promise<TrendData[]> {
  await new Promise((res) => setTimeout(res, 300));
  return CHART_TREND_DATA;
}

export async function fetchLiveTransactions(): Promise<Transaction[]> {
  await new Promise((res) => setTimeout(res, 400));
  return INITIAL_TX_FEED;
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
