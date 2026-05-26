export interface Report {
  title: string;
  date: string;
  type: 'PDF' | 'CSV';
  iconName: 'ShieldCheck' | 'TrendingUp' | 'FileText';
  color: string;
  bg: string;
  content: string;
}
