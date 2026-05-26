import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchCases, updateCaseStatus } from '../../../../modules/blockchain';
import { QUERY_KEYS } from '../../../../constant/query-key';
import type { CaseStatus } from '../../../../modules/blockchain/type';

export function useCases() {
  const queryClient = useQueryClient();

  const casesQuery = useQuery({
    queryKey: [QUERY_KEYS.CASES],
    queryFn: fetchCases,
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: CaseStatus }) =>
      updateCaseStatus(id, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.CASES] });
    },
  });

  return {
    cases: casesQuery.data || [],
    isLoading: casesQuery.isLoading,
    isUpdating: updateMutation.isPending,
    updateStatus: updateMutation.mutateAsync,
  };
}
