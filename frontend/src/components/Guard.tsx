import React from 'react';
import { useParams } from 'react-router-dom';

interface GuardProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
  /**
   * If true, renders children only when in public mode (companyId is present).
   * If false or omitted, renders children only when NOT in public mode (private mode).
   */
  publicOnly?: boolean;
}

export default function Guard({ children, fallback = null, publicOnly = false }: GuardProps) {
  const { companyId } = useParams();
  const isPublic = !!companyId;

  if (publicOnly) {
    return isPublic ? <>{children}</> : <>{fallback}</>;
  }

  return !isPublic ? <>{children}</> : <>{fallback}</>;
}
