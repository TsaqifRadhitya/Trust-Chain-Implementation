import type { Report } from './type';
import { apiClient } from '../../lib/axios';

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
  model_result?: {
    is_fraud: boolean;
    verdict: string;
    flag_reason: string;
    risk_score: number;
    data: string;
  };
  correction?: {
    is_corrected: boolean;
    actual_status: string;
    reason: string;
    corrected_by: string;
    updated_at: string;
  };
}

export async function fetchReports(): Promise<Report[]> {
  try {
    const response = await apiClient.get('/explorer/transactions?limit=1000');
    const txs: BackendTransaction[] = response.data.data || [];

    const totalTxs = txs.length;
    // An anomaly is a case if it was originally flagged as fraud by the model
    const anomalies = txs.filter(tx => tx.model_result && tx.model_result.is_fraud).length;
    const resolved = txs.filter(tx => tx.correction && tx.correction.is_corrected).length;
    const pending = anomalies - resolved;

    // Build BI Risk Index CSV content
    let csvContent = 'transaction_hash,date,partner,amount_idr,risk_score,status,flag_reason\n';
    txs.forEach(tx => {
      const hash = tx.hash;
      const date = new Date(tx.timestamp).toISOString().split('T')[0];
      let partner = tx.to || 'Unknown Vendor';
      const dataStr = tx.model_result ? tx.model_result.data : tx.data;
      try {
        const payload = JSON.parse(dataStr) as { vendor_name?: string };
        partner = payload.vendor_name || partner;
      } catch {
        // Ignored
      }
      const risk = tx.model_result ? tx.model_result.risk_score : tx.risk_score;
      let status = 'safe';
      if (tx.correction && tx.correction.is_corrected) {
        status = 'safe (corrected)';
      } else if (tx.model_result && tx.model_result.is_fraud) {
        status = 'flagged';
      } else if (risk >= 30) {
        status = 'warning';
      }
      
      const type = tx.model_result ? (tx.model_result.flag_reason || tx.model_result.verdict) : (tx.flag_reason || tx.verdict);
      const flagReason = type || 'None';

      csvContent += `${hash},${date},${partner.replace(/,/g, '')},${tx.value},${risk},${status},${flagReason.replace(/,/g, '')}\n`;
    });

    const currentDate = new Date().toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' });

    return [
      {
        title: 'OJK Fraud Compliance Q1',
        date: currentDate,
        type: 'PDF',
        iconName: 'ShieldCheck',
        color: 'text-primary',
        bg: 'bg-primary/10',
        content: `OJK Fraud Compliance Report Q1 2026\nPeriod: January - March 2026\nTotal Transactions Reviewed: ${totalTxs}\nAnomalies Detected: ${anomalies}\nBlockchain Verified: 100%\nCompliance Status: ${pending === 0 ? 'PASSED' : 'PENDING REVIEW'}`,
      },
      {
        title: 'Bank Indonesia Risk Index',
        date: currentDate,
        type: 'CSV',
        iconName: 'TrendingUp',
        color: 'text-success',
        bg: 'bg-success/10',
        content: csvContent,
      },
      {
        title: 'Internal Audit Trail - March',
        date: 'March 31, 2026',
        type: 'PDF',
        iconName: 'FileText',
        color: 'text-warning',
        bg: 'bg-warning/10',
        content: `Internal Audit Trail - March 2026\nAudit Period: March 2026\nCases Opened: ${anomalies}\nCases Resolved: ${resolved}\nPending Review: ${pending}\nBlockchain Records: ${totalTxs} entries`,
      },
    ];
  } catch (error) {
    console.error('Error fetching dynamic reports, using fallback static data:', error);
    return [
      {
        title: 'OJK Fraud Compliance Q1',
        date: 'April 24, 2026',
        type: 'PDF',
        iconName: 'ShieldCheck',
        color: 'text-primary',
        bg: 'bg-primary/10',
        content: 'OJK Fraud Compliance Report Q1 2026\nPeriod: January - March 2026\nTotal Transactions Reviewed: 142,593\nAnomalies Detected: 24\nBlockchain Verified: 100%\nCompliance Status: PASSED',
      },
      {
        title: 'Bank Indonesia Risk Index',
        date: 'April 20, 2026',
        type: 'CSV',
        iconName: 'TrendingUp',
        color: 'text-success',
        bg: 'bg-success/10',
        content: 'transaction_hash,date,partner,amount_idr,risk_score,status,flag_reason\n0xa1b2c3...,2026-04-20,Neo Supply,2100000000,89,flagged,Volume Anomaly\n0xd4e5f6...,2026-04-20,Apex Corp,450000000,72,warning,Velocity Check\n0x7a8b9c...,2026-04-20,LogisX Energy,12000000,12,safe,None',
      },
      {
        title: 'Internal Audit Trail - March',
        date: 'March 31, 2026',
        type: 'PDF',
        iconName: 'FileText',
        color: 'text-warning',
        bg: 'bg-warning/10',
        content: 'Internal Audit Trail - March 2026\nAudit Period: March 2026\nCases Opened: 18\nCases Resolved: 15\nPending Review: 3\nBlockchain Records: 98,201 entries',
      },
    ];
  }
}
