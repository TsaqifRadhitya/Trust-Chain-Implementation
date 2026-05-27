import type { Case, CaseStatus, Transaction, TrendData, VerifyTxResponse } from './type';

const INITIAL_CASES: Case[] = [
  { id: 'CASE-092', txId: 'TX-8823', date: '2026-04-24', status: 'Open', partner: 'Neo Supply', amount: '$2,100,000.00', risk: 89, type: 'Volume Anomaly' },
  { id: 'CASE-091', txId: 'TX-8110', date: '2026-04-23', status: 'In Review', partner: 'Apex Corp', amount: '$450,000.00', risk: 72, type: 'Geographic Mismatch' },
  { id: 'CASE-090', txId: 'TX-7994', date: '2026-04-22', status: 'Resolved', partner: 'LogisX Energy', amount: '$12,400.00', risk: 45, type: 'Velocity Check' },
  { id: 'CASE-089', txId: 'TX-7821', date: '2026-04-21', status: 'Open', partner: 'CyberLog', amount: '$890,000.00', risk: 81, type: 'Duplicate Detection' },
  { id: 'CASE-088', txId: 'TX-7700', date: '2026-04-20', status: 'In Review', partner: 'Global Manuf.', amount: '$312,000.00', risk: 63, type: 'Volume Anomaly' },
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

function getStoredCases(): Case[] {
  const stored = localStorage.getItem('tc_cases');
  if (stored) {
    try {
      return JSON.parse(stored);
    } catch {
      // Fallback
    }
  }
  localStorage.setItem('tc_cases', JSON.stringify(INITIAL_CASES));
  return INITIAL_CASES;
}

export async function fetchCases(): Promise<Case[]> {
  await new Promise((res) => setTimeout(res, 500));
  return getStoredCases();
}

export async function updateCaseStatus(caseId: string, status: CaseStatus): Promise<Case> {
  await new Promise((res) => setTimeout(res, 500));
  const cases = getStoredCases();
  const index = cases.findIndex((c) => c.id === caseId);
  if (index === -1) {
    throw new Error('Case not found');
  }

  cases[index].status = status;
  localStorage.setItem('tc_cases', JSON.stringify(cases));
  return cases[index];
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
  await new Promise((res) => setTimeout(res, 1200));

  return {
    hash: hash || '0x7f8ba1689234b678de9a4b2c1234a4c58d0426f8',
    status: 'Success',
    timestamp: '2026-04-24 10:25:34 UTC',
    from: '0x4D2B8A043C0E56A9D...2B',
    to: '0x99F1B8A043C0E56A9...1F',
    payload: {
      txId: 'TX-8823',
      aiScore: 89,
      flags: ['Volume Anomaly'],
      model_version: 'v2.4.1',
      signature: '0xfa89...11',
    },
  };
}
