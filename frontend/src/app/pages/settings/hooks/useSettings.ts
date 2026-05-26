import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getConfig, saveConfig } from '../../../../modules/setting';
import { QUERY_KEYS } from '../../../../constant/query-key';
import type { Config } from '../../../../modules/setting/type';

export function useSettings() {
  const queryClient = useQueryClient();

  const configQuery = useQuery({
    queryKey: [QUERY_KEYS.CONFIG],
    queryFn: getConfig,
  });

  const saveMutation = useMutation({
    mutationFn: (config: Config) => saveConfig(config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.CONFIG] });
    },
  });

  return {
    config: configQuery.data,
    isLoading: configQuery.isLoading,
    isSaving: saveMutation.isPending,
    save: saveMutation.mutateAsync,
  };
}
