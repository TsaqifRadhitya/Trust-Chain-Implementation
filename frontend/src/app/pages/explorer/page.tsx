import React, { useState } from 'react';
import { Search, Link as LinkIcon, CheckCircle, Clock, ShieldCheck, ShieldAlert, AlertTriangle, Loader2 } from 'lucide-react';
import { motion } from 'framer-motion';
import { useParams } from 'react-router-dom';
import { useExplorer } from './hooks/useExplorer';
import Guard from '../../../components/Guard';

export default function BlockchainExplorer() {
  const [hash, setHash] = useState('');
  const [searched, setSearched] = useState(false);
  const { companyId } = useParams();
  const { 
    data, 
    isLoading, 
    error, 
    verify,
    validationData,
    isValidating,
    validationError,
    validate 
  } = useExplorer();

  const isPublic = !!companyId;

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!hash) return;
    setSearched(false);
    try {
      await verify(hash);
      setSearched(true);
    } catch {
      setSearched(true);
    }
  };

  const handleIntegrityCheck = async () => {
    try {
      await validate();
    } catch (err) {
      console.error(err);
    }
  };

  const getSanitizedPayload = (payload: Record<string, unknown>) => {
    if (isPublic) {
      const sanitized = { ...payload };
      delete sanitized.signature;
      return sanitized;
    }
    return payload;
  };

  return (
    <div className="space-y-6 max-w-4xl mx-auto pb-10">
      <div className="text-center py-10">
        <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary/10 text-primary mb-4">
          <LinkIcon className="w-8 h-8" />
        </div>
        <h1 className="text-3xl font-bold text-white tracking-tight">Immutable Audit Trail</h1>
        <p className="text-textMuted mt-2 max-w-lg mx-auto">Verify any transaction on the TrustChain network. Our distributed ledger guarantees tamper-proof records for your auditors.</p>
      </div>

      {/* Cryptographic Integrity Audit Panel */}
      <div className="bg-surface border border-slate-700/50 rounded-2xl p-6 relative overflow-hidden">
        <div className="absolute top-0 right-0 p-4 opacity-5">
          <ShieldCheck className="w-24 h-24 text-white" />
        </div>
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h2 className="text-lg font-bold text-white flex items-center gap-2">
              <ShieldCheck className="w-5 h-5 text-primary" />
              Cryptographic Integrity Audit
            </h2>
            <p className="text-xs text-textMuted mt-1">
              Verify the mathematical integrity of the Ganache blockchain. This scans all blocks to check parent hash continuity.
            </p>
          </div>
          <button
            onClick={handleIntegrityCheck}
            disabled={isValidating}
            className="flex items-center justify-center gap-2 bg-slate-800 hover:bg-slate-700 disabled:bg-slate-800/50 text-white border border-slate-700 hover:border-slate-600 px-5 py-2.5 rounded-xl font-medium text-sm transition-all duration-200"
          >
            {isValidating ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin text-primary" />
                Auditing Nodes...
              </>
            ) : (
              'Run Integrity Check'
            )}
          </button>
        </div>

        {/* Audit Results */}
        {validationData && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className="mt-6 border-t border-slate-700/50 pt-5 space-y-4"
          >
            {validationData.is_valid ? (
              <div className="flex items-start gap-3 bg-success/10 border border-success/30 p-4 rounded-xl text-success">
                <ShieldCheck className="w-6 h-6 shrink-0 mt-0.5" />
                <div>
                  <h3 className="font-bold text-sm">Blockchain Intact & Secure</h3>
                  <p className="text-xs text-success/80 mt-0.5">
                    All {validationData.total_blocks} blocks successfully validated. Cryptographic hash-chain links are 100% correct. No tampering detected.
                  </p>
                </div>
              </div>
            ) : (
              <div className="flex items-start gap-3 bg-red-500/10 border border-red-500/30 p-4 rounded-xl text-red-500">
                <ShieldAlert className="w-6 h-6 shrink-0 mt-0.5" />
                <div>
                  <h3 className="font-bold text-sm">Blockchain Integrity Compromised!</h3>
                  <p className="text-xs text-red-500/80 mt-0.5">
                    Detected corrupted or broken block links! The mathematical proof of immutability failed validation.
                  </p>
                </div>
              </div>
            )}

            {/* Block list detail */}
            <div className="space-y-2">
              <h4 className="text-xs font-semibold text-white uppercase tracking-wider">Audit Details</h4>
              <div className="max-h-48 overflow-y-auto space-y-1.5 pr-2 custom-scrollbar">
                {validationData.details.map((b) => (
                  <div 
                    key={b.height} 
                    className="flex justify-between items-center text-xs p-2 bg-slate-900/50 border border-slate-800 rounded-lg hover:border-slate-700/80 transition"
                  >
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-textMuted bg-slate-800 px-1.5 py-0.5 rounded">
                        Block #{b.height}
                      </span>
                      <span className="font-mono text-[10px] text-textMuted max-w-[200px] md:max-w-sm truncate">
                        Hash: {b.hash}
                      </span>
                    </div>
                    <div className="flex items-center gap-2">
                      {b.tx_count > 0 && (
                        <span className="text-[10px] text-accent/80 font-medium">
                          {b.tx_count} txs
                        </span>
                      )}
                      <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${
                        b.status === 'OK' ? 'bg-success/15 text-success' : 'bg-red-500/15 text-red-500'
                      }`}>
                        {b.status}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </motion.div>
        )}

        {validationError && (
          <div className="mt-4 p-4 bg-red-500/10 border border-red-500/30 text-red-500 rounded-xl text-xs flex items-center gap-2">
            <AlertTriangle className="w-5 h-5 shrink-0" />
            <span>Failed to run integrity audit: {validationError.message || 'Node connection error'}</span>
          </div>
        )}
      </div>

      <form onSubmit={handleSearch} className="relative group">
        <div className="absolute -inset-1 bg-linear-to-r from-primary to-accent rounded-xl blur opacity-25 group-hover:opacity-50 transition duration-1000 group-hover:duration-200"></div>
        <div className="relative flex items-center bg-surface border border-slate-700 p-2 rounded-xl">
          <Search className="w-6 h-6 ml-3 text-slate-400" />
          <input 
            type="text" 
            value={hash}
            onChange={(e) => setHash(e.target.value)}
            placeholder="Search by Txn Hash..." 
            className="flex-1 bg-transparent border-none focus:ring-0 text-white px-4 placeholder-slate-500"
          />
          <button 
            type="submit"
            className="bg-primary hover:bg-primary/90 text-white px-6 py-3 rounded-lg font-medium transition-colors"
          >
            {isLoading ? 'Verifying...' : 'Verify'}
          </button>
        </div>
      </form>

      {isLoading && (
        <div className="py-20 flex flex-col items-center justify-center text-primary">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mb-4"></div>
          <p className="animate-pulse">Querying TrustChain Nodes...</p>
        </div>
      )}

      {searched && !isLoading && error && (
        <motion.div 
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-red-500/10 border border-red-500/50 text-red-500 p-6 rounded-2xl text-center"
        >
          <p className="font-bold text-lg">Transaction Not Found</p>
          <p className="text-sm mt-1">We couldn't find a transaction with the given hash on the TrustChain network.</p>
        </motion.div>
      )}

      {searched && !isLoading && !error && data && (
        <motion.div 
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-surface border border-slate-700/50 rounded-2xl overflow-hidden"
        >
          <div className="p-6 border-b border-slate-700/50 flex justify-between items-start">
            <div>
              <h2 className="text-xl font-bold text-white flex items-center">
                <CheckCircle className="w-6 h-6 text-success mr-2" />
                Transaction Verified
              </h2>
              <p className="text-sm text-textMuted mt-1 break-all">Hash: {data.hash}</p>
            </div>
            <span className="px-3 py-1 bg-success/20 text-success border border-success/30 rounded-full text-xs font-bold uppercase tracking-wide">
              {data.status}
            </span>
          </div>
          
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-4">
              <div>
                <p className="text-xs text-textMuted mb-1">Status</p>
                <p className="text-white font-medium flex items-center">
                  <CheckCircle className="w-4 h-4 text-success mr-2" /> Confirmed (Block #{data.blockHeight})
                </p>
              </div>
              <div>
                <p className="text-xs text-textMuted mb-1">Timestamp</p>
                <p className="text-white font-medium flex items-center">
                  <Clock className="w-4 h-4 text-slate-400 mr-2" /> {data.timestamp}
                </p>
              </div>
              <Guard>
                <div>
                  <p className="text-xs text-textMuted mb-1">From (ERP Backend)</p>
                  <p className="text-primary font-mono text-sm break-all">{data.from}</p>
                </div>
                <div>
                  <p className="text-xs text-textMuted mb-1">To (Smart Contract)</p>
                  <p className="text-primary font-mono text-sm break-all">{data.to}</p>
                </div>
              </Guard>
            </div>
            
            <div className="p-4 bg-slate-800/50 rounded-xl border border-slate-700/50">
              <h3 className="text-sm font-semibold text-white mb-3">AI Attestation Payload</h3>
              <pre className="text-xs text-accent font-mono overflow-x-auto whitespace-pre-wrap">
{JSON.stringify(getSanitizedPayload(data.payload), null, 2)}
              </pre>
            </div>
          </div>
        </motion.div>
      )}
    </div>
  );
}
