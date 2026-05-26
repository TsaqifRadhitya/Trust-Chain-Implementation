import { NavLink, useParams } from 'react-router-dom';
import { LayoutDashboard, ShieldAlert, Link as LinkIcon, FileText, Settings, Hexagon } from 'lucide-react';
import clsx from 'clsx';
import { motion } from 'framer-motion';

export default function Sidebar() {
  const { companyId } = useParams();
  const isPublic = !!companyId;

  const items = isPublic
    ? [
        { name: 'Dashboard', path: `/public/${companyId}/dashboard`, icon: LayoutDashboard },
        { name: 'Alerts & Cases', path: `/public/${companyId}/cases`, icon: ShieldAlert },
        { name: 'Blockchain Explorer', path: `/public/${companyId}/explorer`, icon: LinkIcon },
      ]
    : [
        { name: 'Dashboard', path: '/dashboard', icon: LayoutDashboard },
        { name: 'Alerts & Cases', path: '/cases', icon: ShieldAlert },
        { name: 'Blockchain Explorer', path: '/explorer', icon: LinkIcon },
        { name: 'Compliance Reports', path: '/reports', icon: FileText },
        { name: 'Configuration', path: '/settings', icon: Settings },
      ];

  return (
    <div className="hidden md:flex flex-col w-64 bg-surface border-r border-slate-700/50 h-full">
      <div className="h-16 flex items-center px-6 border-b border-slate-700/50">
        <Hexagon className="w-8 h-8 text-primary" strokeWidth={1.5} />
        <span className="ml-3 font-semibold text-lg tracking-wide text-white">TrustChain <span className="text-primary font-bold">AI</span></span>
      </div>

      <nav className="flex-1 overflow-y-auto py-6 px-3 space-y-1">
        {items.map((item) => {
          const Icon = item.icon;
          return (
            <NavLink
              key={item.path}
              to={item.path}
              className={({ isActive }) =>
                clsx(
                  'flex items-center px-3 py-2.5 rounded-lg transition-all duration-200 group relative',
                  isActive 
                    ? 'text-white bg-primary/10'
                    : 'text-textMuted hover:text-white hover:bg-surfaceHover'
                )
              }
            >
              {({ isActive }) => (
                <>
                  {isActive && (
                    <motion.div 
                      layoutId="sidebar-active" 
                      className="absolute left-0 top-0 bottom-0 w-1 bg-primary rounded-r-full" 
                    />
                  )}
                  <Icon className={clsx('w-5 h-5 mr-3', isActive ? 'text-primary' : 'text-slate-400 group-hover:text-slate-300')} />
                  <span className="font-medium text-sm">{item.name}</span>
                </>
              )}
            </NavLink>
          );
        })}
      </nav>

      <div className="p-4 border-t border-slate-700/50">
        <div className="rounded-xl bg-linear-to-tr from-primary/20 to-accent/20 p-4 border border-primary/20 relative overflow-hidden">
          <div className="absolute top-0 right-0 p-2 opacity-20">
            <Hexagon className="w-16 h-16" />
          </div>
          <h4 className="text-white text-sm font-semibold mb-1">AI Engine Active</h4>
          <p className="text-xs text-textMuted">Processing 1.2k tx/s</p>
          <div className="mt-3 flex items-center space-x-2">
            <span className="flex h-2 w-2 relative">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2 w-2 bg-accent"></span>
            </span>
            <span className="text-xs text-accent font-medium uppercase tracking-wider">Secured</span>
          </div>
        </div>
      </div>
    </div>
  );
}
