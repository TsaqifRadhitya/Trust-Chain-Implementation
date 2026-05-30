import { useState, useMemo } from 'react';
import { Search, Filter, ShieldAlert, ChevronDown } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { useParams } from 'react-router-dom';
import { useToast } from '../../../components/Toast';
import { useCases } from './hooks/useCases';
import { FLAG_REASONS } from '../../../modules/blockchain';
import type { Case, CaseStatus } from '../../../modules/blockchain/type';
import Guard from '../../../components/Guard';

const STATUS_FILTERS = ['All', 'Open', 'In Review', 'Resolved'] as const;
type FilterOption = typeof STATUS_FILTERS[number];

export default function CaseManagement() {
  const { toast } = useToast();
  const { companyId } = useParams();
  const { cases, isLoading, updateStatus } = useCases();
  const [selectedCase, setSelectedCase] = useState<Case | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<FilterOption>('All');
  const [filterOpen, setFilterOpen] = useState(false);

  const isPublic = !!companyId;

  const filteredCases = useMemo(() => {
    return cases.filter((c) => {
      const q = searchQuery.toLowerCase();
      const matchSearch = !q || c.id.toLowerCase().includes(q) || c.partner.toLowerCase().includes(q)
        || c.type.toLowerCase().includes(q) || c.txId.toLowerCase().includes(q);
      const matchStatus = statusFilter === 'All' || c.status === statusFilter;
      return matchSearch && matchStatus;
    });
  }, [cases, searchQuery, statusFilter]);

  const handleUpdateStatus = async (caseId: string, newStatus: CaseStatus) => {
    try {
      await updateStatus({ id: caseId, status: newStatus });
      setSelectedCase(null);
      if (newStatus === 'In Review') {
        toast(`${caseId} dieskalasi ke Compliance.`, 'warning');
      } else {
        toast(`${caseId} ditandai aman dan diselesaikan.`, 'success');
      }
    } catch (e: unknown) {
      toast(`Gagal mengubah status: ${(e as Error).message}`, 'error');
    }
  };

  const tableHeaders = isPublic
    ? ['Case ID', 'Transaction Info', 'AI Flag Reason', 'Risk Score', 'Status']
    : ['Case ID', 'Transaction Info', 'AI Flag Reason', 'Risk Score', 'Status', 'Action'];

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Alerts & Cases</h1>
          <p className="text-sm text-textMuted mt-1">
            Investigate AI-flagged transactions.{' '}
            <span className="text-primary font-medium">{filteredCases.length} kasus</span>
          </p>
        </div>
        <div className="flex space-x-3">
          <div className="relative">
            <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
            <input type="text" value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Cari case ID, partner..." className="bg-surface border border-slate-700 text-sm py-2 pl-9 pr-4 rounded-lg focus:outline-none focus:border-primary text-white w-52" />
          </div>
          <div className="relative">
            <button onClick={() => setFilterOpen((v) => !v)}
              className="flex items-center px-4 py-2 bg-surface border border-slate-700 rounded-lg text-sm text-white hover:bg-surfaceHover transition-colors">
              <Filter className="w-4 h-4 mr-2" />{statusFilter === 'All' ? 'Filter' : statusFilter}<ChevronDown className="w-3 h-3 ml-2 text-slate-400" />
            </button>
            <AnimatePresence>
              {filterOpen && (
                <motion.div initial={{ opacity: 0, y: -8 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -8 }}
                  className="absolute right-0 mt-2 w-36 bg-surface border border-slate-700 rounded-xl overflow-hidden shadow-2xl z-20">
                  {STATUS_FILTERS.map((s) => (
                    <button key={s} onClick={() => { setStatusFilter(s); setFilterOpen(false); }}
                      className={`w-full text-left px-4 py-2.5 text-sm transition-colors ${statusFilter === s ? 'text-primary bg-primary/10' : 'text-slate-300 hover:bg-surfaceHover'}`}>
                      {s}
                    </button>
                  ))}
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>
      </div>

      <div className="bg-surface border border-slate-700/50 rounded-2xl overflow-hidden">
        <div className="overflow-x-auto">
          {isLoading ? (
            <div className="text-center py-16 text-textMuted">Loading cases...</div>
          ) : (
            <table className="w-full text-left text-sm text-textMuted">
              <thead className="bg-slate-800/50 text-xs uppercase border-b border-slate-700/50">
                <tr>
                  {tableHeaders.map((h) => (
                    <th key={h} className="px-6 py-4 font-semibold text-slate-300">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {filteredCases.length === 0 ? (
                  <tr><td colSpan={isPublic ? 5 : 6} className="text-center py-16 text-textMuted">
                    <Search className="w-8 h-8 mx-auto mb-2 opacity-30" /><p>Tidak ada kasus yang cocok.</p>
                  </td></tr>
                ) : filteredCases.map((c) => (
                  <tr key={c.id} className="border-b border-slate-700/50 hover:bg-slate-800/30 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap text-white font-medium">
                      <button onClick={() => setSelectedCase(c)} className="hover:text-primary hover:underline transition-colors font-semibold text-left">{c.id}</button>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <p className="text-white">{c.partner}</p>
                      <p className="text-xs text-textMuted">{c.txId} • {c.amount}</p>
                      <p className="text-[10px] text-slate-500 font-mono mt-0.5 truncate w-48" title={c.originalHash}>{c.originalHash}</p>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="flex items-center text-danger"><ShieldAlert className="w-3 h-3 mr-1" />{c.type}</span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        <div className="w-[60px] bg-slate-700 rounded-full h-1.5 mr-2">
                          <div className={`h-1.5 rounded-full ${c.risk > 80 ? 'bg-danger' : c.risk > 50 ? 'bg-warning' : 'bg-success'}`} style={{ width: `${c.risk}%` }} />
                        </div>
                        <span className="text-white font-bold">{c.risk}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-2.5 py-1 rounded-full text-xs font-semibold ${
                        c.status === 'Open' ? 'bg-danger/20 text-danger border border-danger/30' :
                        c.status === 'In Review' ? 'bg-warning/20 text-warning border border-warning/30' :
                        'bg-success/20 text-success border border-success/30'}`}>{c.status}</span>
                    </td>
                    <Guard>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {c.status !== 'Resolved'
                          ? <button onClick={() => setSelectedCase(c)} className="text-primary hover:text-white font-medium transition-colors">Investigate</button>
                          : <span className="text-success text-xs font-medium">✓ Resolved</span>}
                      </td>
                    </Guard>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      <AnimatePresence>
        {selectedCase && (
          <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-sm"
            onClick={() => setSelectedCase(null)}>
            <motion.div initial={{ scale: 0.95, y: 20 }} animate={{ scale: 1, y: 0 }} exit={{ scale: 0.95, y: 20 }}
              className="bg-surface border border-slate-700 rounded-2xl p-6 w-full max-w-lg shadow-2xl relative"
              onClick={(e) => e.stopPropagation()}>
              <div className="flex items-start justify-between mb-2">
                <h3 className="text-xl font-bold text-white">Investigation: {selectedCase.id}</h3>
                <span className={`px-2.5 py-1 rounded-full text-xs font-semibold ${
                  selectedCase.status === 'Open' ? 'bg-danger/20 text-danger border border-danger/30' : 'bg-warning/20 text-warning border border-warning/30'}`}>
                  {selectedCase.status}
                </span>
              </div>
              <div className="mb-6">
                <p className="text-sm text-textMuted">{selectedCase.partner} • {selectedCase.txId} • {selectedCase.amount}</p>
                <p className="text-xs text-slate-500 font-mono mt-1 break-all">Hash: {selectedCase.originalHash}</p>
              </div>
              <div className="space-y-4 mb-6">
                <div className="bg-danger/10 border border-danger/20 p-4 rounded-xl">
                  <h4 className="text-danger font-semibold text-sm mb-1 flex items-center"><ShieldAlert className="w-4 h-4 mr-1"/>AI Flag Reason</h4>
                  <p className="text-sm text-slate-300">{FLAG_REASONS[selectedCase.type] ?? 'Anomaly detected by AI engine.'}</p>
                </div>
                <div className="bg-slate-800/50 border border-slate-700/50 p-4 rounded-xl">
                  <div className="flex justify-between text-sm mb-2">
                    <span className="text-textMuted">AI Risk Score</span>
                    <span className={`font-bold ${selectedCase.risk > 80 ? 'text-danger' : selectedCase.risk > 50 ? 'text-warning' : 'text-success'}`}>{selectedCase.risk}/100</span>
                  </div>
                  <div className="w-full bg-slate-700 rounded-full h-2">
                    <motion.div initial={{ width: 0 }} animate={{ width: `${selectedCase.risk}%` }} transition={{ duration: 0.6 }}
                      className={`h-2 rounded-full ${selectedCase.risk > 80 ? 'bg-danger' : selectedCase.risk > 50 ? 'bg-warning' : 'bg-success'}`} />
                  </div>
                </div>
              </div>
              <div className="flex space-x-3 justify-end">
                <button onClick={() => setSelectedCase(null)} className="px-4 py-2 rounded-lg border border-slate-600 text-white hover:bg-slate-700 transition-colors text-sm">Close</button>
                <Guard>
                  {selectedCase.status === 'Open' && (
                    <button onClick={() => handleUpdateStatus(selectedCase.id, 'In Review')} className="px-4 py-2 rounded-lg bg-warning text-white hover:bg-warning/90 transition-colors text-sm font-medium">Escalate to Compliance</button>
                  )}
                  <button onClick={() => handleUpdateStatus(selectedCase.id, 'Resolved')} className="px-4 py-2 rounded-lg bg-success text-white hover:bg-success/90 transition-colors text-sm font-medium">Mark as Safe</button>
                </Guard>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
