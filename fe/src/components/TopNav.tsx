import { useState, useRef, useEffect } from 'react';
import { Bell, Search, User, LogOut, ChevronDown, ShieldAlert } from 'lucide-react';
import { useNavigate, useParams } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { useAuth } from '../app/hooks/AuthContext';
import Guard from './Guard';

const NOTIFICATIONS = [
  { id: 1, msg: 'AI flagged TX-8823: Volume Anomaly (Score 89)', time: '2 min ago', read: false },
  { id: 2, msg: 'CASE-091 escalated to Compliance review', time: '15 min ago', read: false },
  { id: 3, msg: 'Blockchain sync completed: 1,402 blocks confirmed', time: '1 hr ago', read: true },
];

export default function TopNav() {
  const navigate = useNavigate();
  const { companyId } = useParams();
  const { user, logout } = useAuth();
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const [notifOpen, setNotifOpen] = useState(false);
  const userRef = useRef<HTMLDivElement>(null);
  const notifRef = useRef<HTMLDivElement>(null);

  const unreadCount = NOTIFICATIONS.filter((n) => !n.read).length;

  // Close dropdowns on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (userRef.current && !userRef.current.contains(e.target as Node)) setUserMenuOpen(false);
      if (notifRef.current && !notifRef.current.contains(e.target as Node)) setNotifOpen(false);
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <header className="h-16 bg-surface/50 backdrop-blur-md border-b border-slate-700/50 flex items-center justify-between px-4 md:px-6 z-10">
      <Guard
        publicOnly
        fallback={
          <div className="flex-1 max-w-xl">
            <div className="relative">
              <Search className="w-5 h-5 absolute left-3 top-1/2 -translate-y-1/2 text-textMuted" />
              <input
                type="text"
                placeholder="Search transactions, cases, or hashes..."
                className="w-full bg-slate-800/50 border border-slate-700 text-white text-sm rounded-full pl-10 pr-4 py-2 focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-all"
              />
            </div>
          </div>
        }
      >
        <div className="flex items-center space-x-3">
          <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            Public Auditor View
          </span>
          <span className="text-slate-300 text-sm font-semibold">Company ID: <span className="text-white font-mono">{companyId}</span></span>
        </div>
      </Guard>

      <div className="flex items-center space-x-4 ml-6">
        <Guard>
          {/* Notification Bell */}
          <div className="relative" ref={notifRef}>
            <button
              onClick={() => { setNotifOpen((v) => !v); setUserMenuOpen(false); }}
              className="relative p-2 text-textMuted hover:text-white transition-colors rounded-full hover:bg-slate-800"
            >
              <Bell className="w-5 h-5" />
              {unreadCount > 0 && (
                <span className="absolute top-1 right-1 w-4 h-4 bg-danger rounded-full border border-surface text-[9px] font-bold text-white flex items-center justify-center">
                  {unreadCount}
                </span>
              )}
            </button>
            <AnimatePresence>
              {notifOpen && (
                <motion.div
                  initial={{ opacity: 0, y: -8, scale: 0.95 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  exit={{ opacity: 0, y: -8, scale: 0.95 }}
                  className="absolute right-0 mt-2 w-80 bg-surface border border-slate-700 rounded-2xl shadow-2xl z-30 overflow-hidden"
                >
                  <div className="px-4 py-3 border-b border-slate-700/50 flex items-center justify-between">
                    <h3 className="text-sm font-semibold text-white">Notifications</h3>
                    <span className="text-xs text-primary">{unreadCount} unread</span>
                  </div>
                  <div className="divide-y divide-slate-700/50">
                    {NOTIFICATIONS.map((n) => (
                      <div key={n.id} className={`px-4 py-3 flex gap-3 ${!n.read ? 'bg-primary/5' : ''}`}>
                        <ShieldAlert className="w-4 h-4 mt-0.5 shrink-0 text-danger" />
                        <div>
                          <p className={`text-xs ${!n.read ? 'text-slate-200' : 'text-slate-400'}`}>{n.msg}</p>
                          <p className="text-[10px] text-slate-500 mt-0.5">{n.time}</p>
                        </div>
                      </div>
                    ))}
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>

          <div className="h-8 w-px bg-slate-700/50 mx-2" />

          {/* User Menu */}
          <div className="relative" ref={userRef}>
            <button
              onClick={() => { setUserMenuOpen((v) => !v); setNotifOpen(false); }}
              className="flex items-center space-x-3 text-textMuted hover:text-white transition-colors"
            >
              <div className="w-8 h-8 rounded-full bg-primary/20 border border-primary/30 flex items-center justify-center text-primary">
                <User className="w-4 h-4" />
              </div>
              <span className="hidden md:inline-block text-sm font-medium">{user?.name ?? 'Admin SSO'}</span>
              <ChevronDown className="w-3 h-3 hidden md:block" />
            </button>
            <AnimatePresence>
              {userMenuOpen && (
                <motion.div
                  initial={{ opacity: 0, y: -8, scale: 0.95 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  exit={{ opacity: 0, y: -8, scale: 0.95 }}
                  className="absolute right-0 mt-2 w-52 bg-surface border border-slate-700 rounded-2xl shadow-2xl z-30 overflow-hidden"
                >
                  <div className="px-4 py-3 border-b border-slate-700/50">
                    <p className="text-sm font-semibold text-white">{user?.name}</p>
                    <p className="text-xs text-textMuted">{user?.role}</p>
                    <p className="text-xs text-slate-500 mt-0.5">{user?.email}</p>
                  </div>
                  <button
                    onClick={handleLogout}
                    className="w-full flex items-center px-4 py-3 text-sm text-danger hover:bg-danger/10 transition-colors"
                  >
                    <LogOut className="w-4 h-4 mr-2" /> Sign Out
                  </button>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </Guard>
      </div>
    </header>
  );
}
