import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchCases, updateCaseStatus } from '../../../../modules/blockchain';
import type { CaseStatus } from '../../../../modules/blockchain/type';

export function useCases() {
  const queryClient = useQueryClient();

  const casesQuery = useQuery({
    queryKey: ['cases'],
    queryFn: fetchCases,
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: CaseStatus }) =>
      updateCaseStatus(id, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cases'] });
    },
  });

  return {
    cases: casesQuery.data || [],
    isLoading: casesQuery.isLoading,
    isUpdating: updateMutation.isPending,
    updateStatus: updateMutation.mutateAsync,
  };
}
