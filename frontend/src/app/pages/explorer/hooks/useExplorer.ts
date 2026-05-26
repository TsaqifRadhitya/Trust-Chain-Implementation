import { useMutation } from '@tanstack/react-query';
import { verifyTx } from '../../../../modules/blockchain';

export function useExplorer() {
  const verifyMutation = useMutation({
    mutationFn: (hash: string) => verifyTx(hash),
  });

  return {
    data: verifyMutation.data,
    isLoading: verifyMutation.isPending,
    error: verifyMutation.error,
    verify: verifyMutation.mutateAsync,
  };
}
