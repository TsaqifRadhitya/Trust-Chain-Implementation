export type CaseStatus = 'Open' | 'In Review' | 'Resolved';

export interface Case {
  id: string;
  txId: string;
  date: string;
  status: CaseStatus;
  partner: string;
  amount: string;
  risk: number;
  type: string;
  originalHash: string;
}

export type TxStatus = 'safe' | 'warning' | 'flagged';

export interface Transaction {
  id: string;
  partner: string;
  amount: string;
  status: TxStatus;
  aiScore: number;
}

export interface TrendData {
  time: string;
  risk: number;
}

export interface VerifyTxResponse {
  hash: string;
  status: string;
  timestamp: string;
  from: string;
  to: string;
  blockHeight: number;
  payload: Record<string, unknown>;
}

export interface ValidationDetail {
  height: number;
  hash: string;
  parent_hash: string;
  previous_block_hash: string;
  timestamp: string;
  tx_count: number;
  status: 'OK' | 'CORRUPTED';
  error?: string;
}

export interface ChainValidationResponse {
  is_valid: boolean;
  total_blocks: number;
  validated_blocks: number;
  details: ValidationDetail[];
}
