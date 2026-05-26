import { FileText, Download, TrendingUp, ShieldCheck } from 'lucide-react';
import { useToast } from '../../../components/Toast';
import { useReports } from './hooks/useReports';
import type { Report } from '../../../modules/report/type';

const ICON_MAP = {
  ShieldCheck: ShieldCheck,
  TrendingUp: TrendingUp,
  FileText: FileText,
};

export default function Reports() {
  const { toast } = useToast();
  const { reports, isLoading } = useReports();

  const handleDownload = (report: Report) => {
    const mimeType = report.type === 'PDF' ? 'application/pdf' : 'text/csv';
    const fileName = `${report.title.replace(/\s+/g, '_')}.${report.type.toLowerCase()}`;
    const blob = new Blob([report.content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = fileName;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    toast(`${fileName} berhasil diunduh!`, 'success');
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Compliance Reports</h1>
          <p className="text-sm text-textMuted mt-1">Generate automated audit and regulatory reports (OJK, BI).</p>
        </div>
      </div>

      {isLoading ? (
        <div className="text-center py-16 text-textMuted">Loading compliance reports...</div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {reports.map((report, i) => {
            const IconComponent = ICON_MAP[report.iconName] || FileText;
            return (
              <div key={i} className="bg-surface border border-slate-700/50 rounded-2xl p-6 flex flex-col items-center text-center group hover:border-primary/50 transition-colors">
                <div className={`w-16 h-16 rounded-full ${report.bg} flex items-center justify-center mb-4`}>
                  <IconComponent className={`w-8 h-8 ${report.color}`} />
                </div>
                <h3 className="text-lg font-bold text-white mb-1">{report.title}</h3>
                <p className="text-sm text-textMuted mb-2">Generated on {report.date}</p>
                <span className="text-xs text-slate-500 mb-6 px-2 py-0.5 rounded bg-slate-800 border border-slate-700">{report.type}</span>
                <button
                  onClick={() => handleDownload(report)}
                  className="w-full flex items-center justify-center px-4 py-2 bg-slate-800 hover:bg-primary hover:text-white text-slate-300 rounded-lg transition-colors font-medium text-sm border border-slate-700 group-hover:border-primary"
                >
                  <Download className="w-4 h-4 mr-2" /> Download {report.type}
                </button>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
