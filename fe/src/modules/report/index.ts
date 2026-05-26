import type { Report } from './type';

const REPORTS: Report[] = [
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
    content: 'date,partner,risk_score,status\n2026-04-20,Neo Supply,89,flagged\n2026-04-20,Apex Corp,72,warning\n2026-04-20,LogisX Energy,12,safe',
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

export async function fetchReports(): Promise<Report[]> {
  await new Promise((res) => setTimeout(res, 500));
  return REPORTS;
}
