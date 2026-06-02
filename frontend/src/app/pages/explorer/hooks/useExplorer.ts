import { useMutation } from '@tanstack/react-query';
import { verifyTx, validateChain } from '../../../../modules/blockchain';

export function useExplorer() {
  const verifyMutation = useMutation({
    mutationFn: (hash: string) => verifyTx(hash),
  });

  const validateMutation = useMutation({
    mutationFn: () => validateChain(),
  });

  return {
    data: verifyMutation.data,
    isLoading: verifyMutation.isPending,
    error: verifyMutation.error,
    verify: verifyMutation.mutateAsync,

    validationData: validateMutation.data,
    isValidating: validateMutation.isPending,
    validationError: validateMutation.error,
    validate: validateMutation.mutateAsync,
  };
}
