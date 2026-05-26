import { useState, useEffect } from 'react';
import { Save, Link as Database, Key } from 'lucide-react';
import { useToast } from '../../../components/Toast';
import { useSettings } from './hooks/useSettings';

export default function Configuration() {
  const { toast } = useToast();
  const { config, isLoading, isSaving, save } = useSettings();

  const [erpType, setErpType] = useState('');
  const [endpoint, setEndpoint] = useState('');
  const [apiKey, setApiKey] = useState('');
  const [volumeSensitivity, setVolumeSensitivity] = useState(0);
  const [geoThreshold, setGeoThreshold] = useState(0);
  const [velocityLimit, setVelocityLimit] = useState(0);

  useEffect(() => {
    if (config) {
      setErpType(config.erpType);
      setEndpoint(config.endpoint);
      setApiKey(config.apiKey);
      setVolumeSensitivity(config.volumeSensitivity);
      setGeoThreshold(config.geoThreshold);
      setVelocityLimit(config.velocityLimit);
    }
  }, [config]);

  const getSensitivityLabel = (val: number) => {
    if (val >= 80) return { label: 'High', color: 'text-danger' };
    if (val >= 50) return { label: 'Medium', color: 'text-warning' };
    return { label: 'Low', color: 'text-success' };
  };

  const handleSave = async () => {
    try {
      await save({
        erpType,
        endpoint,
        apiKey,
        volumeSensitivity,
        geoThreshold,
        velocityLimit,
      });
      toast('Konfigurasi berhasil disimpan!', 'success');
    } catch (e: any) {
      toast(`Gagal menyimpan: ${e.message}`, 'error');
    }
  };

  if (isLoading || !config) {
    return (
      <div className="flex items-center justify-center py-20 text-textMuted">
        Loading configuration...
      </div>
    );
  }

  return (
    <div className="space-y-6 max-w-4xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight">System Configuration</h1>
          <p className="text-sm text-textMuted mt-1">Manage ERP connections, AI parameters, and Blockchain settings.</p>
        </div>
        <button
          onClick={handleSave}
          disabled={isSaving}
          className="flex items-center px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors text-sm font-medium disabled:opacity-70"
        >
          {isSaving ? (
            <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin mr-2" />
          ) : (
            <Save className="w-4 h-4 mr-2" />
          )}
          {isSaving ? 'Menyimpan...' : 'Save Changes'}
        </button>
      </div>

      <div className="space-y-6">
        {/* ERP Connector */}
        <div className="bg-surface border border-slate-700/50 rounded-2xl p-6">
          <div className="flex items-center mb-6">
            <div className="w-10 h-10 rounded-full bg-slate-800 flex items-center justify-center mr-4">
              <Database className="w-5 h-5 text-primary" />
            </div>
            <div>
              <h2 className="text-lg font-bold text-white">ERP API Connector</h2>
              <p className="text-sm text-textMuted">Connect your internal systems to TrustChain AI.</p>
            </div>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-1">ERP System Type</label>
              <select
                value={erpType}
                onChange={(e) => setErpType(e.target.value)}
                className="w-full bg-slate-800/50 border border-slate-700 rounded-lg py-2 px-3 text-white focus:outline-none focus:border-primary"
              >
                <option>SAP S/4HANA</option>
                <option>Oracle NetSuite</option>
                <option>Microsoft Dynamics</option>
                <option>Custom REST API</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-1">Endpoint URL</label>
              <input
                type="text"
                value={endpoint}
                onChange={(e) => setEndpoint(e.target.value)}
                className="w-full bg-slate-800/50 border border-slate-700 rounded-lg py-2 px-3 text-white focus:outline-none focus:border-primary"
              >
              </input>
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-slate-300 mb-1">API Key</label>
              <input
                type="text"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                className="w-full bg-slate-800/50 border border-slate-700 rounded-lg py-2 px-3 text-white font-mono text-sm focus:outline-none focus:border-primary"
              />
            </div>
          </div>
          <div className="mt-4 flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-success animate-pulse" />
            <span className="text-xs text-success font-medium">Connected to {erpType}</span>
          </div>
        </div>

        {/* AI Parameters */}
        <div className="bg-surface border border-slate-700/50 rounded-2xl p-6">
          <div className="flex items-center mb-6">
            <div className="w-10 h-10 rounded-full bg-slate-800 flex items-center justify-center mr-4">
              <Key className="w-5 h-5 text-accent" />
            </div>
            <div>
              <h2 className="text-lg font-bold text-white">AI Risk Parameters</h2>
              <p className="text-sm text-textMuted">Tune the sensitivity of the AI fraud detection engine.</p>
            </div>
          </div>

          <div className="space-y-8">
            {/* Volume Anomaly */}
            <div>
              <div className="flex justify-between mb-2">
                <label className="block text-sm font-medium text-slate-300">Volume Anomaly Sensitivity</label>
                <span className={`text-sm font-bold ${getSensitivityLabel(volumeSensitivity).color}`}>
                  {getSensitivityLabel(volumeSensitivity).label} ({volumeSensitivity}%)
                </span>
              </div>
              <input
                type="range" min="0" max="100"
                value={volumeSensitivity}
                onChange={(e) => setVolumeSensitivity(Number(e.target.value))}
                className="w-full accent-accent cursor-pointer"
              />
              <div className="flex justify-between text-xs text-slate-500 mt-1"><span>Low</span><span>High</span></div>
            </div>

            {/* Geographic Deviation */}
            <div>
              <div className="flex justify-between mb-2">
                <label className="block text-sm font-medium text-slate-300">Geographic Deviation Threshold</label>
                <span className={`text-sm font-bold ${getSensitivityLabel(geoThreshold).color}`}>
                  {getSensitivityLabel(geoThreshold).label} ({geoThreshold}%)
                </span>
              </div>
              <input
                type="range" min="0" max="100"
                value={geoThreshold}
                onChange={(e) => setGeoThreshold(Number(e.target.value))}
                className="w-full accent-warning cursor-pointer"
              />
              <div className="flex justify-between text-xs text-slate-500 mt-1"><span>Low</span><span>High</span></div>
            </div>

            {/* Velocity Limit */}
            <div>
              <div className="flex justify-between mb-2">
                <label className="block text-sm font-medium text-slate-300">Transaction Velocity Limit</label>
                <span className={`text-sm font-bold ${getSensitivityLabel(velocityLimit).color}`}>
                  {getSensitivityLabel(velocityLimit).label} ({velocityLimit}%)
                </span>
              </div>
              <input
                type="range" min="0" max="100"
                value={velocityLimit}
                onChange={(e) => setVelocityLimit(Number(e.target.value))}
                className="w-full accent-primary cursor-pointer"
              />
              <div className="flex justify-between text-xs text-slate-500 mt-1"><span>Low</span><span>High</span></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
